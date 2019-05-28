package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// Parameters is  a map of parameter names and values
type Parameters map[string]string

// SystemsManagerParameter defines an AWS Systems Manager Parameter
type SystemsManagerParameter struct {
	Name  string
	Type  string
	Value string
}

// SystemsManagerParameters is  a map of parameter names and SystemsManagerParameter objects
type SystemsManagerParameters map[string]*SystemsManagerParameter

var profile, prefix, output, input, value, env, domain, project string
var overwrite bool
var download *flag.FlagSet
var upload *flag.FlagSet
var scrape *flag.FlagSet
var promote *flag.FlagSet
var validate *flag.FlagSet
var allparams = make(map[string]*SystemsManagerParameter)

// CreateSession returns a new AWS session
func CreateSession() (sess *session.Session, err error) {

	awsRegion := aws.String("eu-west-1") // example "eu-west-1"   EU (Ireland)
	if profile != "" {
		os.Setenv("AWS_PROFILE", profile)
	}

	sess, err = session.NewSession(&aws.Config{
		Region: awsRegion,
	})

	return sess, err
}

// SetParameter sets a parameter on parameter store, paramType can be one of these: String, StringList, SecureString
func SetParameter(paramName string, paramType string, paramValue string, overwrite bool) (err error) {
	sess, err := CreateSession()
	if err != nil {
		return err
	}

	svc := ssm.New(sess)

	input := &ssm.PutParameterInput{
		Name:      aws.String(paramName),
		Type:      aws.String(paramType),
		Value:     aws.String(paramValue),
		Overwrite: aws.Bool(overwrite),
	}
	_, err = svc.PutParameter(input)
	if err != nil {
		return err
	}

	return nil
}

// DeleteParameter deletes a parameter on parameter store
func DeleteParameter(paramName string) (err error) {
	sess, err := CreateSession()
	if err != nil {
		return err
	}

	svc := ssm.New(sess)

	input := &ssm.DeleteParameterInput{
		Name: aws.String(paramName),
	}
	_, err = svc.DeleteParameter(input)
	if err != nil {
		return err
	}

	return nil
}

// GetParameterByName deletes a parameter on parameter store
func GetParameterByName(paramName string) (param SystemsManagerParameter, err error) {
	sess, err := CreateSession()
	if err != nil {
		return param, err
	}

	svc := ssm.New(sess)
	var output *ssm.GetParameterOutput

	input := &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	}
	output, err = svc.GetParameter(input)
	if err != nil {
		return param, err
	}
	param = SystemsManagerParameter{Name: *output.Parameter.Name, Type: *output.Parameter.Type, Value: *output.Parameter.Value}

	return param, nil
}

// GetParametersByValue scrape the entire parameter store searching for all keys with a specific value
func GetParametersByValue(paramValue string) (params SystemsManagerParameters, err error) {
	if len(allparams) == 0 {
		allparams, err = GetSystemsManagerParametersByPath("/")
		if err != nil {
			println(err.Error())
		}
	}

	params = make(map[string]*SystemsManagerParameter)

	for _, value := range allparams {
		if value.Value == paramValue {
			//fmt.Println(value.Value)
			params[value.Name] = &SystemsManagerParameter{Name: value.Name, Type: value.Type, Value: value.Value}
		}
	}
	if len(params) == 0 {
		return params, errors.New("can't find any parameter with value " + paramValue)
	}
	return params, nil
}

// GetSystemsManagerParametersByPath retrieves the parameter from the AWS System Manager Parameter Store starting from the initial path recursively.
func GetSystemsManagerParametersByPath(path string) (params SystemsManagerParameters, err error) {
	sess, err := CreateSession()
	if err != nil {
		return nil, err
	}
	params = make(map[string]*SystemsManagerParameter)
	svc := ssm.New(sess)

	var output *ssm.GetParametersByPathOutput
	var nextToken *string
	for output == nil || nextToken != nil {
		input := &ssm.GetParametersByPathInput{
			MaxResults:     aws.Int64(10),
			Path:           aws.String(path),
			Recursive:      aws.Bool(true),
			WithDecryption: aws.Bool(true),
			NextToken:      nextToken,
		}
		output, err = svc.GetParametersByPath(input)
		if err != nil {
			return nil, err
		}
		nextToken = output.NextToken
		for _, par := range output.Parameters {
			params[*par.Name] = &SystemsManagerParameter{Name: *par.Name, Type: *par.Type, Value: *par.Value}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return params, nil
}

// DownloadParametersByPath retrieves the parameter from the AWS System Manager Parameter Store.
func DownloadParametersByPath(path string) {
	params, err := GetSystemsManagerParametersByPath(path)
	if err != nil {
		println(err.Error())
	}

	records := [][]string{}

	for key, value := range params {
		if output != "" {
			records = append(records, []string{key, value.Type, value.Value})
		} else {
			println(value.Type + " " + value.Name + " " + value.Value)
		}
	}
	if output != "" {
		fileName := fmt.Sprintf("download-%s-%s", output, time.Now().UTC().Format("20060102150405"))

		file, err := os.Create(getCsvPath(fileName))
		if err != nil {
			println(err.Error())
		}
		defer file.Close()

		w := csv.NewWriter(file)

		w.WriteAll(records) // calls Flush internally

		if err := w.Error(); err != nil {
			log.Fatalln("error writing csv:", err)
		}
	}
}

// UploadParametersFromCsv read parameters from CSV and write them to the AWS System Manager Parameter Store.
func UploadParametersFromCsv(filename string, overwrite bool) {

	// Open the file
	csvfile, err := os.Open(getCsvPath(filename))
	if err != nil {
		println(err.Error())
	}

	// Parse the file
	r := csv.NewReader(csvfile)

	records, err := r.ReadAll()
	if err != nil {
		println(err.Error())
	}
	fmt.Println(records)

	if len(records) > 0 {
		for _, row := range records {
			err := SetParameter(row[0], row[1], row[2], overwrite)
			if err != nil {
				println(err.Error())
			}
		}
	}
	println("ok")
}

// DownloadParametersByValue read parameters from Parameter store and return all keys with a specific value.
func DownloadParametersByValue(scrapevalue string) {
	params, err := GetSystemsManagerParametersByPath("/")
	if err != nil {
		println(err.Error())
	}

	records := [][]string{}
	fileName := fmt.Sprintf("scrape-%s-%s", output, time.Now().UTC().Format("20060102150405"))

	for key, value := range params {
		if value.Value == scrapevalue {
			if output != "" {
				records = append(records, []string{key, value.Type, value.Value})
			} else {
				println(value.Type + " " + value.Name + " " + value.Value)
			}
		}
	}
	if output != "" {
		file, err := os.Create(getCsvPath(fileName))
		if err != nil {
			println(err.Error())
		}
		defer file.Close()

		w := csv.NewWriter(file)

		w.WriteAll(records) // calls Flush internally

		if err := w.Error(); err != nil {
			log.Fatalln("error writing csv:", err)
		}
	}
	println("ok")
}

// PromoteParameters download all parameters linked to a project.
func PromoteParameters(env string, domain string, project string) {
	params, err := GetSystemsManagerParametersByPath("/" + env + "/" + domain + "/" + project)
	if err != nil {
		println(err.Error())
	}

	records := [][]string{}

	for key, value := range params {
		records = append(records, []string{key, value.Type, value.Value})
	}

	for _, value := range params {
		if strings.HasPrefix(value.Value, "/"+env+"/common") {
			common, err := GetParameterByName(value.Value)
			if err != nil {
				println(err.Error())
			} else {
				records = append(records, []string{common.Name, common.Type, common.Value})
			}
		}
	}

	fileName := fmt.Sprintf("promote-%s-%s-%s", project, env, time.Now().UTC().Format("20060102150405"))

	file, err := os.Create(getCsvPath(fileName))
	if err != nil {
		println(err.Error())
	}
	defer file.Close()

	w := csv.NewWriter(file)

	w.WriteAll(records) // calls Flush internally

	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}

	println("ok")
}

// ValidateParameters read parameters from a CSV and check for inconsistencies.
func ValidateParameters(filename string, env string) {
	// Open the file
	csvfile, err := os.Open(getCsvPath(filename))
	if err != nil {
		println(err.Error())
	}
	//params := make(map[string]*SystemsManagerParameter)
	params := make(SystemsManagerParameters)

	// Parse the file
	r := csv.NewReader(csvfile)

	records, err := r.ReadAll()
	if err != nil {
		println(err.Error())
	}
	//fmt.Println(records)

	// Create SystemsManagerParameters from CSV data
	for _, row := range records {
		params[row[0]] = &SystemsManagerParameter{Name: row[0], Type: row[1], Value: row[2]}
	}

	for _, param := range params {
		if strings.HasPrefix(param.Name, "/"+env+"/common") {
			commonvar, err := GetParameterByName(param.Name)
			if err != nil {
				//println(param.Name + " non esiste sul parameter store... cerco altri parametri in ambiente " + env + " con lo stesso valore:")
				commonvalues, err := GetParametersByValue(param.Value)
				if err != nil {
					//println("non ho trovato altre variabili comuni con il valore " + param.Value + " , la nuova variabile comune " + param.Name + " può essere inserita.")
					println("MISSING -> CREATE      - " + param.Name + " WITH VALUE " + param.Value)
				} else {
					println("MISSING -> DUPLICATE   - " + param.Name + " WITH VALUE " + param.Value)
					for _, commonvalue := range commonvalues {
						if strings.HasPrefix(commonvalue.Name, "/"+env+"/common") {
							println("- " + commonvalue.Name + " with value " + commonvalue.Value)
						}
					}
					//println("considera la possibilità di modificare il puntamento della variabile di progetto verso una di queste common")
				}
			} else {
				if commonvar.Value == param.Value {
					println("PRESENT -> MAINTAIN    - " + param.Name + " WITH VALUE " + param.Value)
				} else {
					println("PRESENT -> DESTRUCTIVE - " + param.Name + " WITH VALUE " + param.Value + " Caricare questo CSV potrebbe provocare problemi con altri progetti")
				}
			}
		} else {
			projectvar, err := GetParameterByName(param.Name)
			if err != nil {
				println("MISSING -> CREATE      - " + param.Name + " WITH VALUE " + param.Value + "")
			} else {
				if projectvar.Value == param.Value {
					println("PRESENT -> MAINTAIN    - " + param.Name + " WITH VALUE " + param.Value)
				} else {
					println("PRESENT -> OVERWRITE   - " + projectvar.Value + " WITH VALUE " + param.Value)
				}
			}
		}
	}
	println("ok")
}

func getCsvPath(filename string) string {
	return fmt.Sprintf("./%s.csv", filename)
}

func main() {
	flag.Parse()

	if len(os.Args) < 2 {

		flag.PrintDefaults()

		fmt.Printf("[download]\n")
		download.PrintDefaults()

		fmt.Printf("[upload]\n")
		upload.PrintDefaults()

		fmt.Printf("[scrape]\n")
		scrape.PrintDefaults()

		fmt.Printf("[promote]\n")
		promote.PrintDefaults()

		fmt.Printf("[validate]\n")
		validate.PrintDefaults()

		os.Exit(0)
	}

	switch os.Args[1] {

	case "download":
		download.Parse(os.Args[2:])
		if prefix == "" {
			download.PrintDefaults()
			os.Exit(1)
		}

		DownloadParametersByPath(prefix)

	case "upload":
		upload.Parse(os.Args[2:])
		if input == "" {
			upload.PrintDefaults()
			os.Exit(1)
		}

		UploadParametersFromCsv(input, overwrite)

	case "scrape":
		scrape.Parse(os.Args[2:])
		if value == "" {
			scrape.PrintDefaults()
			os.Exit(1)
		}

		DownloadParametersByValue(value)

	case "promote":
		promote.Parse(os.Args[2:])
		if env == "" || domain == "" || project == "" {
			promote.PrintDefaults()
			os.Exit(1)
		}

		PromoteParameters(env, domain, project)

	case "validate":
		validate.Parse(os.Args[2:])
		if input == "" || env == "" {
			validate.PrintDefaults()
			os.Exit(1)
		}

		ValidateParameters(input, env)

	default:
		flag.PrintDefaults()
		os.Exit(1)

	}
}

func init() {
	download = flag.NewFlagSet("SearchByPath", flag.ExitOnError)
	download.StringVar(&profile, "profile", "", "(optional) AWS profile")
	download.StringVar(&prefix, "prefix", "", "(required) prefix path to download")
	download.StringVar(&output, "output", "", "(optional) Output CSV file")
	upload = flag.NewFlagSet("Upload", flag.ExitOnError)
	upload.StringVar(&profile, "profile", "", "(optional) AWS profile")
	upload.StringVar(&input, "input", "", "(required) Input CSV file")
	upload.BoolVar(&overwrite, "overwrite", false, "(optional) Overwrite the value if the key already exists")
	scrape = flag.NewFlagSet("SearchByValue", flag.ExitOnError)
	scrape.StringVar(&profile, "profile", "", "(optional) AWS profile")
	scrape.StringVar(&value, "value", "", "(required) The Value to search")
	scrape.StringVar(&output, "output", "", "(optional) Output CSV file")
	promote = flag.NewFlagSet("Promote", flag.ExitOnError)
	promote.StringVar(&profile, "profile", "", "(optional) AWS profile")
	promote.StringVar(&env, "env", "", "(required) The source environment")
	promote.StringVar(&domain, "domain", "", "(required) The project domain")
	promote.StringVar(&project, "project", "", "(required) The project name")
	validate = flag.NewFlagSet("Validate", flag.ExitOnError)
	validate.StringVar(&profile, "profile", "", "(optional) AWS profile")
	validate.StringVar(&input, "input", "", "(required) Input CSV file")
	validate.StringVar(&env, "env", "", "(required) The target environment")
}
