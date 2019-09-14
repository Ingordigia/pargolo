package tests

import (
	"github.com/ory/dockertest"
	"github.com/pkg/errors"
)

type TestUtils struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func (this *TestUtils) SetupSuite() error {
	var err error

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	this.pool, err = dockertest.NewPool("")
	if err != nil {
		err = errors.Wrap(err, "Could not create docker pool")
		return err
	}

	// pulls an image, creates a container based on it and runs it
	opt := &dockertest.RunOptions{
		Repository: "localstack/localstack",
		Tag:        "latest",
		Name:       "localstack_demo",
		Cmd:        []string{},
		//PortBindings: map[dc.Port][]dc.PortBinding{"1433/tcp": []dc.PortBinding{{HostPort: "1433"}}},
		Env: []string{"SERVICES=ssm", "DEBUG=1", "DATA_DIR=/tmp/localstack/data"},
	}

	this.resource, err = this.pool.RunWithOptions(opt)
	if err != nil {
		err = errors.Wrap(err, "Could not start docker resource")
		return err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := this.pool.Retry(func() error {
		return nil
	}); err != nil {
		err = errors.Wrap(err, "Could not access docker resource")
		return err
	}
	return nil
}
func (this *TestUtils) TearDownSuite() error {
	// You can't defer this because os.Exit doesn't care for defer
	if err := this.pool.Purge(this.resource); err != nil {
		err = errors.Wrap(err, "Could not purge docker resource")
		return err
	}
	return nil
}

func (this *TestUtils) SetupTest(testcase string) error {

	return nil
}
func (this *TestUtils) TearDownTest(testcase string) error {

	return nil
}
func (this *TestUtils) InitTestData(testcase string, others ...string) error {

	return nil
}
