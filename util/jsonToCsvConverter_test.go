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
	assert.Equal(t, `Environment`, actualCsv[0])
}

func TestOneVariableOneLevelWithValue_ButOnlyEmpty_ShouldReturnNothing(t *testing.T) {
	//input := `{"Environment": ""}`
	inputPath := "json/testOneVariableOneLevelWithValue.json"
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
	expected := `Sentry/SentryDSN`

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
	expected := `Sentry/SentryDSN1/SentryDSN2`

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

	assert.Equal(t, true, strings.Contains(actual, `OAuth2/SecretKey`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/ClientSecret`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Endpoint`))
	assert.Equal(t, true, strings.Contains(actual, `Environment`))
	assert.Equal(t, true, strings.Contains(actual, `WebApp/Port`))
	assert.Equal(t, true, strings.Contains(actual, `Log/RollingFile/File`))
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
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/SecretKey`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/ClientSecret`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Endpoint`))
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

	assert.Equal(t, true, strings.Contains(actual, `Environment`))
	assert.Equal(t, true, strings.Contains(actual, `AppClientId`))
	assert.Equal(t, true, strings.Contains(actual, `WebApp/Port`))
	assert.Equal(t, true, strings.Contains(actual, `Sentry/SentryDSN`))
	assert.Equal(t, true, strings.Contains(actual, `Log/RollingFile/File`))
	assert.Equal(t, true, strings.Contains(actual, `ConsoleCredentials/Username`))
	assert.Equal(t, true, strings.Contains(actual, `ConsoleCredentials/Password`))
	assert.Equal(t, true, strings.Contains(actual, `Redis/Endpoint`))
	assert.Equal(t, true, strings.Contains(actual, `Redis/Port`))
	assert.Equal(t, true, strings.Contains(actual, `Redis/Database`))
	assert.Equal(t, true, strings.Contains(actual, `AccountsRedis/Endpoint`))
	assert.Equal(t, true, strings.Contains(actual, `AccountsRedis/Port`))
	assert.Equal(t, true, strings.Contains(actual, `AccountsRedis/Database`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/SecretKey`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/ClientSecret`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Endpoint`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Port`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Database`))
	assert.Equal(t, true, strings.Contains(actual, `ServiceEndpoints/Endpoint1/Endpoint`))
	assert.Equal(t, true, strings.Contains(actual, `ServiceEndpoints/Endpoint2/Endpoint`))
}
