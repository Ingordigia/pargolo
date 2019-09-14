package repositories

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/diegoavanzini/pargolo/models"
	"os"
	"time"
)

// CreateSession returns a new AWS session
func CreateSession(Profile, inputRegion *string) (sess *session.Session, err error) {

	awsRegion := aws.String("eu-west-1") // example "eu-west-1"   EU (Ireland)
	if inputRegion != nil && *inputRegion != "" {
		awsRegion = aws.String(*inputRegion) // example "eu-west-1"   EU (Ireland)
	}
	if Profile != nil && *Profile != "" {
		os.Setenv("AWS_PROFILE", *Profile)
	}

	sess, err = session.NewSession(&aws.Config{
		Region: awsRegion,
	})

	return sess, err
}

func NewRepository(endpoint, profile, awsRegion *string) (IPargoloRepository, error) {

	repo := &ssm_repository{}

	if endpoint != nil {
		repo.Endpoint = *endpoint
	}

	sess, err := CreateSession(profile, awsRegion)
	if err != nil {
		return nil, err
	}

	repo.awsClient = ssm.New(sess)

	return repo, nil
}

type ssm_repository struct {
	Endpoint  string
	awsClient *ssm.SSM
}

// GetParameterByName deletes a parameter on parameter store
func (r *ssm_repository) GetParameterByName(paramName string) (param models.SystemsManagerParameter, err error) {

	var output *ssm.GetParameterOutput

	input := &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	}
	output, err = r.awsClient.GetParameter(input)
	if err != nil {
		return param, err
	}
	param = models.SystemsManagerParameter{Name: *output.Parameter.Name, Type: *output.Parameter.Type, Value: *output.Parameter.Value}

	return param, nil
}

// DeleteParameter deletes a parameter on parameter store
func (r *ssm_repository) DeleteParameter(paramName string) (err error) {
	input := &ssm.DeleteParameterInput{
		Name: aws.String(paramName),
	}
	_, err = r.awsClient.DeleteParameter(input)
	if err != nil {
		return err
	}

	return nil
}

// GetParametersByPath retrieves the parameter from the AWS System Manager Parameter Store starting from the initial path recursively.
func (r *ssm_repository) GetParametersByPath(path string) (params models.SystemsManagerParameters, err error) {
	params = make(map[string]*models.SystemsManagerParameter)

	var output *ssm.GetParametersByPathOutput
	var nextToken *string
	for output == nil || nextToken != nil {
		input := &ssm.GetParametersByPathInput{
			MaxResults:     aws.Int64(10),
			Path:           aws.String(path),
			Recursive:      aws.Bool(true),
			WithDecryption: aws.Bool(true),
			NextToken:      nextToken,
		}
		output, err = r.awsClient.GetParametersByPath(input)
		if err != nil {
			return nil, err
		}
		nextToken = output.NextToken
		for _, par := range output.Parameters {
			params[*par.Name] = &models.SystemsManagerParameter{Name: *par.Name, Type: *par.Type, Value: *par.Value}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return params, nil
}

// SetParameter sets a parameter on parameter store, paramType can be one of these: String, StringList, SecureString
func (r *ssm_repository) SetParameter(paramName string, paramType string, paramValue string, overwrite bool) (err error) {
	input := &ssm.PutParameterInput{
		Name:      aws.String(paramName),
		Type:      aws.String(paramType),
		Value:     aws.String(paramValue),
		Overwrite: aws.Bool(overwrite),
	}
	_, err = r.awsClient.PutParameter(input)
	if err != nil {
		return err
	}

	return nil
}
