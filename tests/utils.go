package tests

import (
	"fmt"
	"io/ioutil"
	"strings"

	"bitbucket.org/mailupteam/be-statistics-report-builder/configuration"
	"github.com/pkg/errors"

	"database/sql"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/ory/dockertest"
	dc "github.com/ory/dockertest/docker"
)

type TestUtils struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func (this *TestUtils) SetupSuite() error {
	var err error

	configuration.InitConfigurationManager("../config.json")

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	this.pool, err = dockertest.NewPool("")
	if err != nil {
		err = errors.Wrap(err, "Could not create docker pool")
		return err
	}

	// pulls an image, creates a container based on it and runs it
	opt := &dockertest.RunOptions{
		Repository:   "microsoft/mssql-server-linux",
		Tag:          "latest",
		PortBindings: map[dc.Port][]dc.PortBinding{"1433/tcp": []dc.PortBinding{{HostPort: "1433"}}},
		Env:          []string{"ACCEPT_EULA=Y", "SA_PASSWORD=LocalSql_!"},
	}

	this.resource, err = this.pool.RunWithOptions(opt)
	if err != nil {
		err = errors.Wrap(err, "Could not start docker resource")
		return err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := this.pool.Retry(func() error {
		db, err := this.GetConnection("master")
		if err != nil {
			err = errors.Wrap(err, "Error connecting to dockered database")
			return err
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			err = errors.Wrap(err, "Error dockered database not responding")
			return err
		}
		return nil
	}); err != nil {
		err = errors.Wrap(err, "Could not access docker resource")
		return err
	}
	return nil
}
func (this *TestUtils) GetConnection(dbname string) (*sql.DB, error) {
	var err error

	var password = "LocalSql_!"
	var port = 1433
	var server = "localhost"
	var user = "sa"

	conn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&connection+timeout=30", user, password, server, port, dbname)

	db, err := sql.Open("mssql", conn)
	if err != nil {
		err = errors.Wrap(err, "Error connecting to dockered database")
		return nil, err
	}
	return db, nil
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
	err := this.createDb(testcase)
	if err != nil {
		err = errors.Wrap(err, "Test Setup error")
		return err
	}

	return nil
}
func (this *TestUtils) TearDownTest(testcase string) error {
	err := this.dropDb(testcase)
	if err != nil {
		err = errors.Wrap(err, "Test drop database error")
		return err
	}

	return nil
}
func (this *TestUtils) InitTestData(testcase string, others ...string) error {

	folder := strings.Replace(testcase, "_", "/", -1)

	db, err := this.GetConnection(testcase)
	if err != nil {
		return err
	}
	defer db.Close()

	b, err := ioutil.ReadFile(fmt.Sprintf("./sql/%s.sql", folder))
	if err != nil {
		err = errors.Wrap(err, "Test load data error")
		return err
	}
	str := fmt.Sprintf("USE [%s]\n", testcase)
	str = str + string(b)
	_, err = db.Exec(str)
	if err != nil {
		err = errors.Wrap(err, "Test init data error")
		return err
	}

	if others != nil {
		for _, other := range others {
			b, err := ioutil.ReadFile(fmt.Sprintf("./sql/%s.sql", other))
			if err != nil {
				err = errors.Wrap(err, "Test load data error")
				return err
			}
			str = fmt.Sprintf("USE [%s]\n", testcase)
			str = str + string(b)
			_, err = db.Exec(str)
			if err != nil {
				err = errors.Wrap(err, "Test init data error")
				return err
			}
		}
	}

	return nil
}

func (this *TestUtils) createDb(testcase string) error {
	db, err := this.GetConnection("master")
	if err != nil {
		return err
	}
	defer db.Close()

	db.Exec(fmt.Sprintf("CREATE DATABASE %s", testcase))
	b, err := ioutil.ReadFile("./sql/init_schema.sql")
	if err != nil {
		return err
	}
	str := fmt.Sprintf("USE [%s]\n", testcase)
	str = str + string(b)
	_, err = db.Exec(str)
	if err != nil {
		return err
	}
	return nil
}
func (this *TestUtils) dropDb(testcase string) error {
	db, err := this.GetConnection("master")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", testcase))
	if err != nil {
		return err
	}
	return nil
}
