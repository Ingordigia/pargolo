package domains

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/diegoavanzini/pargolo/models"
	"github.com/diegoavanzini/pargolo/repositories"
	"github.com/ingordigia/pargolo/util"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var allparams = make(map[string]*models.SystemsManagerParameter)

// Parameters is  a map of parameter names and values
type Parameters map[string]string

var Profile, output string

type IPargolo interface {
	// UploadParametersFromCsv read parameters from CSV and write them to the AWS System Manager Parameter Store.
	UploadParametersFromCsv(filename string, overwrite bool)
	// DownloadParametersByValue read parameters from Parameter store and return all keys with a specific value.
	DownloadParametersByValue(targetvalue string, filterpath string)
	// ExportParameters download all parameters linked to a project.
	ExportParameters(env string, domain string, project string)
	// ValidateParameters read parameters from a CSV and check for inconsistencies.
	ValidateParameters(filename string, env string)
	// InitializeParameters read a Json config file and extract blank parameters and create a CSV file for pargolo upload.
	InitializeParameters(filename string, env string, domain string, project string)
	// DownloadParametersByPath retrieves the parameter from the AWS System Manager Parameter Store.
	DownloadParametersByPath(path string, recursive bool)
	// GetParametersByValue scrape the entire parameter store searching for all keys with a specific value
	GetParametersByValue(paramValue string) (params models.SystemsManagerParameters, err error)

	GetParameterByName(paramName string) (param models.SystemsManagerParameter, err error)
	SetParameter(paramName string, paramType string, paramValue string, overwrite bool) (err error)
	DeleteParameter(paramName string) (err error)
	GetParametersByPath(path string) (params models.SystemsManagerParameters, err error)
}

type Pargolo struct {
	repo repositories.IPargoloRepository
}

func NewPargolo(repository repositories.IPargoloRepository) (IPargolo, error) {
	pargolo := &Pargolo{}
	pargolo.repo = repository
	return pargolo, nil
}

// UploadParametersFromCsv read parameters from CSV and write them to the AWS System Manager Parameter Store.
func (pargolo *Pargolo) UploadParametersFromCsv(filename string, overwrite bool) {

	// Open the file
	csvfile, err := os.Open(pargolo.getCsvPath(filename))
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
			err := pargolo.repo.SetParameter(row[0], row[1], row[2], overwrite)
			if err != nil {
				println(err.Error())
			}
		}
	}
}

func (pargolo *Pargolo) SetParameter(paramName string, paramType string, paramValue string, overwrite bool) (err error) {
	return pargolo.repo.SetParameter(paramName, paramType, paramValue, overwrite)
}

func (pargolo *Pargolo) GetParameterByName(paramName string) (param models.SystemsManagerParameter, err error) {
	return pargolo.repo.GetParameterByName(paramName)
}

func (pargolo *Pargolo) DeleteParameter(paramName string) (err error) {
	return pargolo.repo.DeleteParameter(paramName)
}

func (pargolo *Pargolo) GetParametersByPath(path string) (params models.SystemsManagerParameters, err error) {
	return pargolo.repo.GetParametersByPath(path)
}

// DownloadParametersByValue read parameters from Parameter store and return all keys with a specific value.
func (pargolo *Pargolo) DownloadParametersByValue(targetvalue string, filterpath string) {
	params, err := pargolo.GetParametersByValue(targetvalue)
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
func (pargolo *Pargolo) ExportParameters(env string, domain string, project string) {
	params, err := pargolo.repo.GetParametersByPath("/" + env + "/" + domain + "/" + project)
	if err != nil {
		println(err.Error())
	}

	records := [][]string{}

	for key, value := range params {
		records = append(records, []string{key, value.Type, value.Value})
	}

	for _, value := range params {
		if strings.Contains(value.Value, "/common/") {
			common, err := pargolo.repo.GetParameterByName(value.Value)
			if err != nil {
				println(err.Error())
			} else {
				records = append(records, []string{common.Name, common.Type, common.Value})
			}
		}
	}

	fileName := fmt.Sprintf("export-%s-%s-%s", project, env, time.Now().UTC().Format("20060102150405"))

	file, err := os.Create(pargolo.getCsvPath(fileName))
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
func (pargolo *Pargolo) ValidateParameters(filename string, env string) {
	// Open the file
	csvfile, err := os.Open(pargolo.getCsvPath(filename))
	if err != nil {
		println(err.Error())
	}

	params := make(models.SystemsManagerParameters)

	// Parse the file
	r := csv.NewReader(csvfile)

	records, err := r.ReadAll()
	if err != nil {
		println(err.Error())
	}

	// Create SystemsManagerParameters from CSV data
	for _, row := range records {
		params[row[0]] = &models.SystemsManagerParameter{Name: row[0], Type: row[1], Value: row[2]}
	}

	for _, param := range params {
		if strings.Contains(param.Name, "/common/") {
			commonvar, err := pargolo.repo.GetParameterByName(param.Name)
			if err != nil {
				commonvalues, err := pargolo.GetParametersByValue(param.Value)
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
			projectvar, err := pargolo.repo.GetParameterByName(param.Name)
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
func (pargolo *Pargolo) InitializeParameters(filename string, env string, domain string, project string) {

	inputPath := filename
	// read data from file
	jsondatafromfile, err := ioutil.ReadFile(inputPath)
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

func (pargolo *Pargolo) getCsvPath(filename string) string {
	return fmt.Sprintf("./%s.csv", filename)
}

func getFilePath(filename string, extension string) string {
	if strings.HasSuffix(filename, extension) {
		return filename
	}
	return fmt.Sprintf("./%s.%s", filename, extension)
}

// PrintMapToShell prints the parameters map to the shell standard Output
func PrintMapToShell(params models.SystemsManagerParameters) {
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

// DownloadParametersByPath retrieves the parameter from the AWS System Manager Parameter Store.
func (pargolo *Pargolo) DownloadParametersByPath(path string, recursive bool) {
	params, err := pargolo.repo.GetParametersByPath(path)
	if err != nil {
		println(err.Error())
	}

	records := [][]string{}

	for key, value := range params {
		if strings.Contains(value.Value, "/common/") && recursive {
			common, err := pargolo.repo.GetParameterByName(value.Value)
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

// GetParametersByValue scrape the entire parameter store searching for all keys with a specific value
func (pargolo *Pargolo) GetParametersByValue(paramValue string) (params models.SystemsManagerParameters, err error) {
	if len(allparams) == 0 {
		allparams, err = pargolo.repo.GetParametersByPath("/")
		if err != nil {
			println(err.Error())
		}
	}

	params = make(map[string]*models.SystemsManagerParameter)

	for _, value := range allparams {
		if value.Value == paramValue {
			params[value.Name] = &models.SystemsManagerParameter{Name: value.Name, Type: value.Type, Value: value.Value}
		}
	}
	if len(params) == 0 {
		return params, errors.New("can't find any parameter with value " + paramValue)
	}
	return params, nil

}
