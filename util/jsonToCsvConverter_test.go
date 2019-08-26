package util

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"testing"
)


func TestOneVariableOneLevel(t *testing.T) {
	//input := `{"Environment": ""}`
	inputPath := "json/testOneVariableOneLevel.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJsonToCsvConverter().Convert(file, true)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, `Environment,string,`, actualCsv[0])
}

func TestOneVariableOneLevelWithValue(t *testing.T) {
	//input := `{"Environment": ""}`
	inputPath := "json/testOneVariableOneLevelWithValue.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJsonToCsvConverter().Convert(file, false)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, `Environment,string,5`, actualCsv[0])
}

func TestOneVariableOneLevelWithValue_ButOnlyEmpty_ShouldReturnNothing(t *testing.T) {
	//input := `{"Environment": ""}`
	inputPath := "json/testOneVariableOneLevelWithValue.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJsonToCsvConverter().Convert(file, true)
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

	actualCsv, err := NewJsonToCsvConverter().Convert(file, true)
	if err != nil {
		t.Fail()
	}
	expected := `Sentry/SentryDSN,string,`

	assert.Equal(t, expected , actualCsv[0])
}

func TestOneVariableTwoLevelAll(t *testing.T) {
	//input := `{"Sentry": {"SentryDSN": ""}}`
	inputPath := "json/testOneVariableTwoLevel.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJsonToCsvConverter().Convert(file, false)
	if err != nil {
		t.Fail()
	}

	actual := strings.Join(actualCsv,"|")

	assert.Equal(t, true, strings.Contains(actual, `Sentry/SentryDSN,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Sentry/ClientID,string,e7a096db-f113-9ca2-b091-2b3f7f4420af`))
}

func TestIfInvalidJsonShouldReturnError(t *testing.T) {
	input := `{"Sentry": {"SentryDSN": ""}`
	// check if file is json
	file, _ := ioutil.ReadFile(input)

	_, err := NewJsonToCsvConverter().Convert(file, true)
	if err != nil {
		assert.Equal(t,"unexpected end of JSON input", err.Error())
	}
}

func TestOneVariableThreeLevel(t *testing.T) {
	//input := `{"Sentry": {"SentryDSN1": {"SentryDSN2": ""}}}`
	inputPath := "json/testOneVariableThreeLevel.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJsonToCsvConverter().Convert(file, true)
	if err != nil {
		t.Fail()
	}
	expected := `Sentry/SentryDSN1/SentryDSN2,string,`

	assert.Equal(t, expected, actualCsv[0])
}

func TestMultilineOneVariableOneLevel(t *testing.T) {
	inputPath := "json/testMultilineOneVariableOneLevel.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJsonToCsvConverter().Convert(file, true)
	if err != nil {
		t.Fail()
	}
	actual := strings.Join(actualCsv,"|")

	assert.Equal(t, true, strings.Contains(actual, `OAuth2/SecretKey,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/ClientSecret,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Endpoint,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Environment,string,`))
	assert.Equal(t, true, strings.Contains(actual, `WebApp/Port,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Log/RollingFile/File,string,`))
}

func TestMultilineOneVariableSameLevel(t *testing.T) {
	//input := `{"Environment": ""}`
	inputPath := "json/testMultilineOneVariableSameLevel.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJsonToCsvConverter().Convert(file, true)
	if err != nil {
		t.Fail()
	}
	actual := strings.Join(actualCsv,"|")
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/SecretKey,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/ClientSecret,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Endpoint,string,`))
}

func TestRealCase(t *testing.T) {
	inputPath := "json/testRealCase.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJsonToCsvConverter().Convert(file, true)
	if err != nil {
		t.Fail()
	}
	actual := strings.Join(actualCsv,"|")

	assert.Equal(t, true, strings.Contains(actual, `Environment,string,`))
	assert.Equal(t, true, strings.Contains(actual, `AppClientId,string,`))
	assert.Equal(t, true, strings.Contains(actual, `WebApp/Port,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Sentry/SentryDSN,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Log/RollingFile/File,string,`))
	assert.Equal(t, true, strings.Contains(actual, `ConsoleCredentials/Username,string,`))
	assert.Equal(t, true, strings.Contains(actual, `ConsoleCredentials/Password,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Redis/Endpoint,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Redis/Port,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Redis/Database,string,`))
	assert.Equal(t, true, strings.Contains(actual, `AccountsRedis/Endpoint,string,`))
	assert.Equal(t, true, strings.Contains(actual, `AccountsRedis/Port,string,`))
	assert.Equal(t, true, strings.Contains(actual, `AccountsRedis/Database,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/SecretKey,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/ClientSecret,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Endpoint,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Port,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Database,string,`))
	assert.Equal(t, true, strings.Contains(actual, `ServiceEndpoints/Endpoint1/Endpoint,string,`))
	assert.Equal(t, true, strings.Contains(actual, `ServiceEndpoints/Endpoint2/Endpoint,string,`))
}


func TestRealCaseNotOnlyEmpty(t *testing.T) {
	inputPath := "json/testRealCase.json"
	// check if file is json
	file, _ := ioutil.ReadFile(inputPath)

	actualCsv, err := NewJsonToCsvConverter().Convert(file, false)
	if err != nil {
		t.Fail()
	}
	actual := strings.Join(actualCsv,"|")
	assert.Equal(t, true, strings.Contains(actual, `Log/AppenderType,string,RollingFile`))
	assert.Equal(t, true, strings.Contains(actual, `Log/RollingFile/MaximumFileSizeMB,string,50`))

	assert.Equal(t, true, strings.Contains(actual, `Environment,string,`))
	assert.Equal(t, true, strings.Contains(actual, `AppClientId,string,`))
	assert.Equal(t, true, strings.Contains(actual, `WebApp/Port,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Sentry/SentryDSN,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Log/RollingFile/File,string,`))
	assert.Equal(t, true, strings.Contains(actual, `ConsoleCredentials/Username,string,`))
	assert.Equal(t, true, strings.Contains(actual, `ConsoleCredentials/Password,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Redis/Endpoint,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Redis/Port,string,`))
	assert.Equal(t, true, strings.Contains(actual, `Redis/Database,string,`))
	assert.Equal(t, true, strings.Contains(actual, `AccountsRedis/Endpoint,string,`))
	assert.Equal(t, true, strings.Contains(actual, `AccountsRedis/Port,string,`))
	assert.Equal(t, true, strings.Contains(actual, `AccountsRedis/Database,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/SecretKey,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/ClientSecret,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Endpoint,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Port,string,`))
	assert.Equal(t, true, strings.Contains(actual, `OAuth2/CacheRedis/Database,string,`))
	assert.Equal(t, true, strings.Contains(actual, `ServiceEndpoints/Endpoint1/Endpoint,string,`))
	assert.Equal(t, true, strings.Contains(actual, `ServiceEndpoints/Endpoint2/Endpoint,string,`))
}
