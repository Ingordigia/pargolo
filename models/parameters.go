package models

// SystemsManagerParameter defines an AWS Systems Manager Parameter
type SystemsManagerParameter struct {
	Name  string
	Type  string
	Value string
}


// SystemsManagerParameters is  a map of parameter names and SystemsManagerParameter objects
type SystemsManagerParameters map[string]*SystemsManagerParameter

