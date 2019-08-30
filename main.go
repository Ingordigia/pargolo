package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/ingordigia/pargolo/util"
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

var profile, path, output, input, value, env, domain, filter, project string
var overwrite, recursive bool
var searchbypath *flag.FlagSet
var upload *flag.FlagSet
var searchbyvalue *flag.FlagSet
var export *flag.FlagSet
var validate *flag.FlagSet
var initialize *flag.FlagSet
var allparams = make(map[string]*SystemsManagerParameter)

// PrintMapToShell prints the parameters map to the shell standard Output
func PrintMapToShell(params SystemsManagerParameters) {
	maxKeyLength := 0
	maxTypeLenght := 0

	if output == "" {
		for _, value := range params {
			if maxKeyLength < len(value.Name) {
				maxKeyLength = len(value.Name)
			}
			if maxTypeLenght < len(value.Type) {
				maxTypeLenght = len(value.Type)
			}
		}
	}
	for _, value := range params {
		println(value.Type + strings.Repeat(" ", maxTypeLenght+1-len(value.Type)) + value.Name + strings.Repeat(" ", maxKeyLength+1-len(value.Name)) + value.Value)
	}
}

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
		allparams, err = GetParametersByPath("/")
		if err != nil {
			println(err.Error())
		}
	}

	params = make(map[string]*SystemsManagerParameter)

	for _, value := range allparams {
		if value.Value == paramValue {
			params[value.Name] = &SystemsManagerParameter{Name: value.Name, Type: value.Type, Value: value.Value}
		}
	}
	if len(params) == 0 {
		return params, errors.New("can't find any parameter with value " + paramValue)
	}
	return params, nil
}

// GetParametersByPath retrieves the parameter from the AWS System Manager Parameter Store starting from the initial path recursively.
func GetParametersByPath(path string) (params SystemsManagerParameters, err error) {
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
func DownloadParametersByPath(path string, recursive bool) {
	params, err := GetParametersByPath(path)
	if err != nil {
		println(err.Error())
	}

	records := [][]string{}

	for key, value := range params {
		if strings.Contains(value.Value, "/common/") && recursive {
			common, err := GetParameterByName(value.Value)
			if err != nil {
				println(err.Error())
			} else {
				param := params[key]
				param.Value = common.Value
				params[key] = param
			}
		}
	}

	if output != "" {
		for key, value := range params {
			records = append(records, []string{key, value.Type, value.Value})
		}
		fileName := fmt.Sprintf("searchbypath-%s-%s", output, time.Now().UTC().Format("20060102150405"))

		file, err := os.Create(getFilePath(fileName, "csv"))
		if err != nil {
			println(err.Error())
		}
		defer file.Close()

		w := csv.NewWriter(file)

		w.WriteAll(records) // calls Flush internally

		if err := w.Error(); err != nil {
			log.Fatalln("error writing csv:", err)
		}
	} else {
		PrintMapToShell(params)
	}
}

// UploadParametersFromCsv read parameters from CSV and write them to the AWS System Manager Parameter Store.
func UploadParametersFromCsv(filename string, overwrite bool) {

	// Open the file
	csvfile, err := os.Open(getFilePath(filename, "csv"))
	if err != nil {
		println(err.Error())
	}

	// Parse the file
	r := csv.NewReader(csvfile)

	records, err := r.ReadAll()
	if err != nil {
		println(err.Error())
	}

	if len(records) > 0 {
		for _, row := range records {
			err := SetParameter(row[0], row[1], row[2], overwrite)
			if err != nil {
				println(err.Error())
			}
		}
	}
}

// DownloadParametersByValue read parameters from Parameter store and return all keys with a specific value.
func DownloadParametersByValue(targetvalue string, filterpath string) {
	params, err := GetParametersByValue(targetvalue)
	if err != nil {
		println(err.Error())
	}

	records := [][]string{}
	fileName := fmt.Sprintf("searchbyvalue-%s-%s", output, time.Now().UTC().Format("20060102150405"))

	for key, value := range params {
		if strings.HasPrefix(value.Name, filterpath) {
			records = append(records, []string{key, value.Type, value.Value})
		} else {
			delete(params, value.Name)
		}
	}
	if output != "" {
		file, err := os.Create(getFilePath(fileName, "csv"))
		if err != nil {
			println(err.Error())
		}
		defer file.Close()

		w := csv.NewWriter(file)

		w.WriteAll(records) // calls Flush internally

		if err := w.Error(); err != nil {
			log.Fatalln("error writing csv:", err)
		}
	} else {
		PrintMapToShell(params)
	}
}

// ExportParameters download all parameters linked to a project.
func ExportParameters(env string, domain string, project string) {
	params, err := GetParametersByPath("/" + env + "/" + domain + "/" + project)
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

	fileName := fmt.Sprintf("export-%s-%s-%s", project, env, time.Now().UTC().Format("20060102150405"))

	file, err := os.Create(getFilePath(fileName, "csv"))
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

// ValidateParameters read parameters from a CSV and check for inconsistencies.
func ValidateParameters(filename string, env string) {
	// Open the file
	csvfile, err := os.Open(getFilePath(filename, "csv"))
	if err != nil {
		println(err.Error())
	}

	params := make(SystemsManagerParameters)

	// Parse the file
	r := csv.NewReader(csvfile)

	records, err := r.ReadAll()
	if err != nil {
		println(err.Error())
	}

	// Create SystemsManagerParameters from CSV data
	for _, row := range records {
		params[row[0]] = &SystemsManagerParameter{Name: row[0], Type: row[1], Value: row[2]}
	}

	for _, param := range params {
		if strings.HasPrefix(param.Name, "/"+env+"/common") {
			commonvar, err := GetParameterByName(param.Name)
			if err != nil {
				commonvalues, err := GetParametersByValue(param.Value)
				if err != nil {
					println("MISSING -> CREATE      - " + param.Name + " WITH VALUE " + param.Value)
				} else {
					println("MISSING -> DUPLICATE   - " + param.Name + " WITH VALUE " + param.Value)
					for _, commonvalue := range commonvalues {
						if strings.HasPrefix(commonvalue.Name, "/"+env+"/common") {
							println("- " + commonvalue.Name + " with value " + commonvalue.Value)
						}
					}
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
					println("PRESENT -> OVERWRITE   - " + projectvar.Name + " WITH VALUE " + param.Value)
				}
			}
		}
	}
}

// InitializeParameters read a Json config file and extract blank parameters and create a CSV file for pargolo upload.
func InitializeParameters(filename string, env string, domain string, project string) {

	// read data from file
	jsondatafromfile, err := ioutil.ReadFile(getFilePath(filename, "json"))
	if err != nil {
		fmt.Println(err)
	}

	// Create csv structure from json data
	actualCsv, err := util.NewJSONToCsvConverter().Convert(jsondatafromfile)
	if err != nil {
		fmt.Println(err)
	}

	//create output csvfile
	csvdatafile, err := os.Create("./data.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer csvdatafile.Close()

	writer := csv.NewWriter(csvdatafile)

	for _, key := range actualCsv {
		var record []string
		record = append(record, "/"+env+"/"+domain+"/"+project+"/"+key)
		record = append(record, "String")
		record = append(record, "")
		writer.Write(record)
	}

	// remember to flush!
	writer.Flush()
}

func getFilePath(filename string, extension string) string {
	if strings.HasSuffix(filename, extension) {
		return filename
	}
	return fmt.Sprintf("./%s.%s", filename, extension)
}

func main() {
	flag.Parse()

	if len(os.Args) < 2 {

		fmt.Printf("\n--- searchbypath ---\n")
		searchbypath.PrintDefaults()

		fmt.Printf("\n--- searchbyvalue ---\n")
		searchbyvalue.PrintDefaults()

		fmt.Printf("\n--- upload ---\n")
		upload.PrintDefaults()

		fmt.Printf("\n--- export ---\n")
		export.PrintDefaults()

		fmt.Printf("\n--- validate ---\n")
		validate.PrintDefaults()

		fmt.Printf("\n--- initialize ---\n")
		initialize.PrintDefaults()

		os.Exit(0)
	}

	switch os.Args[1] {

	case "searchbypath":
		searchbypath.Parse(os.Args[2:])
		if path == "" {
			searchbypath.PrintDefaults()
			os.Exit(1)
		}

		DownloadParametersByPath(path, recursive)

	case "upload":
		upload.Parse(os.Args[2:])
		if input == "" {
			upload.PrintDefaults()
			os.Exit(1)
		}

		UploadParametersFromCsv(input, overwrite)

	case "searchbyvalue":
		searchbyvalue.Parse(os.Args[2:])
		if value == "" {
			searchbyvalue.PrintDefaults()
			os.Exit(1)
		}

		DownloadParametersByValue(value, filter)

	case "export":
		export.Parse(os.Args[2:])
		if env == "" || domain == "" || project == "" {
			export.PrintDefaults()
			os.Exit(1)
		}

		ExportParameters(env, domain, project)

	case "validate":
		validate.Parse(os.Args[2:])
		if input == "" || env == "" {
			validate.PrintDefaults()
			os.Exit(1)
		}

		ValidateParameters(input, env)

	case "initialize":
		initialize.Parse(os.Args[2:])
		if input == "" || env == "" || domain == "" || project == "" {
			initialize.PrintDefaults()
			os.Exit(1)
		}

		InitializeParameters(input, env, domain, project)

	default:
		flag.PrintDefaults()
		os.Exit(1)

	}
}

func init() {
	searchbypath = flag.NewFlagSet("SearchByPath", flag.ExitOnError)
	searchbypath.StringVar(&profile, "profile", "", "(optional) AWS profile")
	searchbypath.StringVar(&path, "path", "", "(required) prefix path to download")
	searchbypath.StringVar(&output, "output", "", "(optional) Output CSV file")
	searchbypath.BoolVar(&recursive, "recursive", false, "(optional) Select if pargolo should recursively resolve parameters value")
	searchbyvalue = flag.NewFlagSet("SearchByValue", flag.ExitOnError)
	searchbyvalue.StringVar(&profile, "profile", "", "(optional) AWS profile")
	searchbyvalue.StringVar(&value, "value", "", "(required) The Value to search")
	searchbyvalue.StringVar(&filter, "filter", "", "(optional) Filters the results by path")
	searchbyvalue.StringVar(&output, "output", "", "(optional) Output CSV file")
	upload = flag.NewFlagSet("Upload", flag.ExitOnError)
	upload.StringVar(&profile, "profile", "", "(optional) AWS profile")
	upload.StringVar(&input, "input", "", "(required) Input CSV file")
	upload.BoolVar(&overwrite, "overwrite", false, "(optional) Overwrite the value if the key already exists")
	export = flag.NewFlagSet("Export", flag.ExitOnError)
	export.StringVar(&profile, "profile", "", "(optional) AWS profile")
	export.StringVar(&env, "env", "", "(required) The source environment")
	export.StringVar(&domain, "domain", "", "(required) The project domain")
	export.StringVar(&project, "project", "", "(required) The project name")
	validate = flag.NewFlagSet("Validate", flag.ExitOnError)
	validate.StringVar(&profile, "profile", "", "(optional) AWS profile")
	validate.StringVar(&input, "input", "", "(required) Input CSV file")
	validate.StringVar(&env, "env", "", "(required) The target environment")
	initialize = flag.NewFlagSet("Initialize", flag.ExitOnError)
	initialize.StringVar(&profile, "profile", "", "(optional) AWS profile")
	initialize.StringVar(&input, "input", "", "(required) Input JSON config file")
	initialize.StringVar(&env, "env", "", "(required) The source environment")
	initialize.StringVar(&domain, "domain", "", "(required) The project domain")
	initialize.StringVar(&project, "project", "", "(required) The project name")
}
