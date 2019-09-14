package repositories

import "github.com/diegoavanzini/pargolo/models"

type IPargoloRepository interface {
	GetParameterByName(paramName string) (param models.SystemsManagerParameter, err error)
	DeleteParameter(paramName string) (err error)
	GetParametersByPath(path string) (params models.SystemsManagerParameters, err error)
	SetParameter(paramName string, paramType string, paramValue string, overwrite bool) (err error)
}
