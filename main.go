package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
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

var profile, prefix, output, input, value, scrapeoutput string
var overwrite bool
var download *flag.FlagSet
var upload *flag.FlagSet
var scrape *flag.FlagSet

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
		time.Sleep(200 * time.Millisecond)
	}
	return params, nil
}

// DownloadParametersByPath retrieves the parameter from the AWS System Manager Parameter Store.
func DownloadParametersByPath(path string, filename string) {
	params, err := GetSystemsManagerParametersByPath(path)
	if err != nil {
		println(err.Error())
	}

	records := [][]string{}

	for key, value := range params {
		records = append(records, []string{key, value.Type, value.Value})
	}

	fileName := fmt.Sprintf("download-%s-%s", filename, time.Now().UTC().Format("20060102150405"))

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
func DownloadParametersByValue(scrapevalue string, filename string) {
	params, err := GetSystemsManagerParametersByPath("/")
	if err != nil {
		println(err.Error())
	}

	records := [][]string{}

	for key, value := range params {

		if value.Value == scrapevalue {
			fmt.Println(value.Value)
			records = append(records, []string{key, value.Type, value.Value})
		}
	}

	fileName := fmt.Sprintf("scrape-%s-%s", filename, time.Now().UTC().Format("20060102150405"))

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

		os.Exit(0)
	}

	switch os.Args[1] {

	case "download":
		download.Parse(os.Args[2:])
		if prefix == "" || output == "" {
			download.PrintDefaults()
			os.Exit(1)
		}

		DownloadParametersByPath(prefix, output)

	case "upload":
		upload.Parse(os.Args[2:])
		if input == "" {
			upload.PrintDefaults()
			os.Exit(1)
		}

		UploadParametersFromCsv(input, overwrite)

	case "scrape":
		scrape.Parse(os.Args[2:])
		if value == "" || scrapeoutput == "" {
			scrape.PrintDefaults()
			os.Exit(1)
		}

		DownloadParametersByValue(value, scrapeoutput)

	default:
		flag.PrintDefaults()
		os.Exit(1)

	}
}

func init() {
	download = flag.NewFlagSet("SearchByPath", flag.ExitOnError)
	download.StringVar(&profile, "profile", "", "(optional) AWS profile")
	download.StringVar(&prefix, "prefix", "", "(required) prefix path to download")
	download.StringVar(&output, "output", "", "(required) Output CSV file")
	upload = flag.NewFlagSet("Upload", flag.ExitOnError)
	upload.StringVar(&profile, "profile", "", "(optional) AWS profile")
	upload.StringVar(&input, "input", "", "(required) Input CSV file")
	upload.BoolVar(&overwrite, "overwrite", false, "(optional) Overwrite the value if the key already exists")
	scrape = flag.NewFlagSet("SearchByValue", flag.ExitOnError)
	scrape.StringVar(&profile, "profile", "", "(optional) AWS profile")
	scrape.StringVar(&value, "value", "", "(required) The Value to search")
	scrape.StringVar(&scrapeoutput, "output", "", "(required) Output CSV file")
}
