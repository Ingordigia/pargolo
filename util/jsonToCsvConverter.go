package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

var keysToIgnore = map[string]bool{
	"Environment":  true,
	"AWSAccessKey": true,
	"AWSSecretKey": true,
}

type converter interface {
	Convert(inputJSON []byte) ([]string, error)
}

type jsonToCsvConverter struct{}

// NewJSONToCsvConverter create a new converter
func NewJSONToCsvConverter() converter {
	return &jsonToCsvConverter{}
}

// Convert return a slice of strings with converted data
func (c *jsonToCsvConverter) Convert(inputJSON []byte) ([]string, error) {
	var ret []string
	jsonMap := map[string]interface{}{}
	err := json.Unmarshal(inputJSON, &jsonMap)
	if err != nil {
		return ret, err
	}
	ret = c.getRows(jsonMap)
	return ret, nil
}

func (c *jsonToCsvConverter) getRows(jsonMap map[string]interface{}) []string {
	var ret []string
	for key, value := range jsonMap {
		finalKey := key
		kind := reflect.TypeOf(value).Kind()
		if kind != reflect.Map {
			if fmt.Sprintf("%v", value) == "" && !keysToIgnore[finalKey] {
				ret = append(ret, strings.ToLower(finalKey))
			}
		} else {
			for _, partialKey := range c.getRows(value.(map[string]interface{})) {
				finalKey = fmt.Sprintf("%s/%s", key, partialKey)
				ret = append(ret, strings.ToLower(finalKey))
			}
		}
	}
	return ret
}
