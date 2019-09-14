package domains

import (
	"github.com/diegoavanzini/pargolo/models"
)

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
	DownloadParametersByPath(path string)
	// GetParametersByValue scrape the entire parameter store searching for all keys with a specific value
	GetParametersByValue(paramValue string) (params models.SystemsManagerParameters, err error)

	GetParameterByName(paramName string) (param models.SystemsManagerParameter, err error)
	SetParameter(paramName string, paramType string, paramValue string, overwrite bool) (err error)
	DeleteParameter(paramName string) (err error)
	GetParametersByPath(path string) (params models.SystemsManagerParameters, err error)
}
