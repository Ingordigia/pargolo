package util

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOneVariableOneLevel(t *testing.T) {
	//input := `{"Environment": ""}`
	inputPath := "json/testOneVariableOneLevel.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJSONToCsvConverter().Convert(file)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, 0, len(actualCsv))
}

func TestOneVariableTwoLevel(t *testing.T) {
	//input := `{"Sentry": {"SentryDSN": ""}}`
	inputPath := "json/testOneVariableTwoLevel.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJSONToCsvConverter().Convert(file)
	if err != nil {
		t.Fail()
	}
	expected := `sentry/sentrydsn`

	assert.Equal(t, expected, actualCsv[0])
}

func TestIfInvalidJsonShouldReturnError(t *testing.T) {
	input := `{"Sentry": {"SentryDSN": ""}`
	// check if file is json
	file, _ := ioutil.ReadFile(input)

	_, err := NewJSONToCsvConverter().Convert(file)
	if err != nil {
		assert.Equal(t, "unexpected end of JSON input", err.Error())
	}
}

func TestOneVariableThreeLevel(t *testing.T) {
	//input := `{"Sentry": {"SentryDSN1": {"SentryDSN2": ""}}}`
	inputPath := "json/testOneVariableThreeLevel.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJSONToCsvConverter().Convert(file)
	if err != nil {
		t.Fail()
	}
	expected := `sentry/sentrydsn1/sentrydsn2`

	assert.Equal(t, expected, actualCsv[0])
}

func TestMultilineOneVariableOneLevel(t *testing.T) {
	inputPath := "json/testMultilineOneVariableOneLevel.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJSONToCsvConverter().Convert(file)
	if err != nil {
		t.Fail()
	}
	actual := strings.Join(actualCsv, "|")

	assert.Equal(t, true, strings.Contains(actual, `oauth2/secretkey`))
	assert.Equal(t, true, strings.Contains(actual, `oauth2/clientsecret`))
	assert.Equal(t, true, strings.Contains(actual, `oauth2/cacheredis/endpoint`))
	assert.Equal(t, false, strings.Contains(actual, `environment`))
	assert.Equal(t, true, strings.Contains(actual, `webapp/port`))
	assert.Equal(t, true, strings.Contains(actual, `log/rollingfile/file`))
}

func TestMultilineOneVariableSameLevel(t *testing.T) {
	//input := `{"Environment": ""}`
	inputPath := "json/testMultilineOneVariableSameLevel.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJSONToCsvConverter().Convert(file)
	if err != nil {
		t.Fail()
	}
	actual := strings.Join(actualCsv, "|")
	assert.Equal(t, true, strings.Contains(actual, `oauth2/secretkey`))
	assert.Equal(t, true, strings.Contains(actual, `oauth2/clientsecret`))
	assert.Equal(t, true, strings.Contains(actual, `oauth2/cacheredis/endpoint`))
}

func TestRealCase(t *testing.T) {
	inputPath := "json/testRealCase.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJSONToCsvConverter().Convert(file)
	if err != nil {
		t.Fail()
	}
	actual := strings.Join(actualCsv, "|")

	assert.Equal(t, false, strings.Contains(actual, `environment`))
	assert.Equal(t, true, strings.Contains(actual, `appclientid`))
	assert.Equal(t, true, strings.Contains(actual, `webapp/port`))
	assert.Equal(t, true, strings.Contains(actual, `sentry/sentrydsn`))
	assert.Equal(t, true, strings.Contains(actual, `log/rollingfile/file`))
	assert.Equal(t, true, strings.Contains(actual, `consolecredentials/username`))
	assert.Equal(t, true, strings.Contains(actual, `consolecredentials/password`))
	assert.Equal(t, true, strings.Contains(actual, `redis/endpoint`))
	assert.Equal(t, true, strings.Contains(actual, `redis/port`))
	assert.Equal(t, true, strings.Contains(actual, `redis/database`))
	assert.Equal(t, true, strings.Contains(actual, `accountsredis/endpoint`))
	assert.Equal(t, true, strings.Contains(actual, `accountsredis/port`))
	assert.Equal(t, true, strings.Contains(actual, `accountsredis/database`))
	assert.Equal(t, true, strings.Contains(actual, `oauth2/secretkey`))
	assert.Equal(t, true, strings.Contains(actual, `oauth2/clientsecret`))
	assert.Equal(t, true, strings.Contains(actual, `oauth2/cacheredis/endpoint`))
	assert.Equal(t, true, strings.Contains(actual, `oauth2/cacheredis/port`))
	assert.Equal(t, true, strings.Contains(actual, `oauth2/cacheredis/database`))
	assert.Equal(t, true, strings.Contains(actual, `serviceendpoints/endpoint1/endpoint`))
	assert.Equal(t, true, strings.Contains(actual, `serviceendpoints/endpoint2/endpoint`))
}
