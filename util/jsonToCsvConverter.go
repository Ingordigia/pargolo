package util

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type converter interface {
	Convert(inputJson []byte, onlyEmpty bool) ([]string, error)
}

type jsonToCsvConverter struct{}

func NewJsonToCsvConverter() converter {
	return &jsonToCsvConverter{}
}

func (c *jsonToCsvConverter) Convert(inputJson []byte, onlyEmpty bool) ([]string, error) {
	var ret []string
	jsonMap := map[string]interface{}{}
	err := json.Unmarshal(inputJson, &jsonMap)
	if err != nil {
		return ret, err
	}
	ret = c.getRows(jsonMap, onlyEmpty)
	return ret, nil
}

func (c *jsonToCsvConverter) getRows(jsonMap map[string]interface{}, onlyEmpty bool) []string {
	var ret []string
	for key, value := range jsonMap {
		finalKey := key
		kind := reflect.TypeOf(value).Kind()
		if kind != reflect.Map {
			if !onlyEmpty || fmt.Sprintf("%v", value) == "" {
				finalKey = fmt.Sprintf("%s,string,%v", key, value)
				ret = append(ret, finalKey)
			}
		} else {
			for _, partialKey := range c.getRows(value.(map[string]interface{}), onlyEmpty) {
				finalKey = fmt.Sprintf("%s/%s", key, partialKey)
				ret = append(ret, finalKey)
			}
		}
	}
	return ret
}
