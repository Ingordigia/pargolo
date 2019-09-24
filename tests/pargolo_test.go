package tests

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	ssm_aws "github.com/aws/aws-sdk-go/service/ssm"
	"github.com/diegoavanzini/pargolo/domains"
	"github.com/diegoavanzini/pargolo/repositories"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PargoloTestSuite struct {
	suite     suite.Suite
	testutils TestUtils
	name      string
}

func (spts *PargoloTestSuite) SetupSuite() {
	spts.name = "PargoloTestSuite"
	fmt.Println(fmt.Sprintf("STARTING TEST SUITE: %s", spts.name))
	spts.testutils = TestUtils{}
	err := spts.testutils.SetupSuite()
	if err != nil {
		spts.suite.FailNow("Suite Setup error", err.Error())
	}
}

func (spts *PargoloTestSuite) TearDownSuite() {
	err := spts.testutils.TearDownSuite()
	if err != nil {
		spts.suite.FailNow("Suite TearDown error", err.Error())
	}
}

func (spts *PargoloTestSuite) BeforeTest(suiteName, testName string) {
	fmt.Println(fmt.Sprintf("EXECUTING - Suite: %s, Test: %s", suiteName, testName))
	err := spts.testutils.SetupTest(fmt.Sprintf("%s_%s", spts.name, testName))
	if err != nil {
		spts.suite.FailNow("Test Setup error", err.Error())
	}
}

func (spts *PargoloTestSuite) AfterTest(suiteName, testName string) {
	//fmt.Println(fmt.Sprintf("FINISHED - Suite: %s, Test: %s", suiteName, testName))
	err := spts.testutils.TearDownTest(fmt.Sprintf("%s_%s", spts.name, testName))
	if err != nil {
		spts.suite.FailNow("Test TearDown error", err.Error())
	}
}

func (spts *PargoloTestSuite) SetT(t *testing.T) {
	spts.suite.SetT(t)
}

func (spts *PargoloTestSuite) T() *testing.T {
	return spts.suite.T()
}

func Test_PargoloTestSuite(t *testing.T) {
	suite.Run(t, new(PargoloTestSuite))
}

/* PUT YOUR TESTS HERE */

func (spts *PargoloTestSuite) Test_WhenParameterIsValid_ShouldSetTheSpecifiedParameter() {

	// ARRANGE
	parameterName := "/Test/IAD/helloWorld"
	parameterValue := "My1stParameter"

	localEndpoint := "localhost:32783"
	repo, err := repositories.NewRepository(&localEndpoint, nil, nil)
	if err != nil {
		spts.suite.FailNow("NewPargolo error", err.Error())
		return
	}
	pargolo, err := domains.NewPargolo(repo)
	if err != nil {
		spts.suite.FailNow("NewPargolo error", err.Error())
		return
	}

	awsClient, err := SetupAwsClient()
	if err != nil {
		spts.suite.FailNow("SetupAwsClient error", err.Error())
		return
	}
	awsGetParameterInput := &ssm_aws.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(true),
	}

	// ACT
	err = pargolo.SetParameter(parameterName, "String", parameterValue, true)
	if err != nil {
		spts.suite.FailNow("SetParameter error: %s", err.Error())
		return
	}

	// ASSERT
	actual, err := awsClient.GetParameter(awsGetParameterInput)
	if err != nil {
		spts.suite.FailNow("GetParameter error", err.Error())
		return
	}
	if actual == nil {
		spts.suite.FailNow("error: returned parameter is Empty")
	}
	if *actual.Parameter.Value != parameterValue {
		spts.suite.FailNow("error: returned parameter value is %s, expected %s", *actual.Parameter.Value, parameterValue)
	}

}

func SetupAwsClient() (*ssm_aws.SSM, error) {

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-1"), // example "eu-west-1"   EU (Ireland)
	})
	if err != nil {
		return nil, err
	}
	awsClient := ssm_aws.New(sess)
	return awsClient, nil
}
