# Pargolo

Pargolo is a command line tool written in golang that helps you simplify AWS Parameter Store management.
With Pargolo you can download, upload, search (by-key or by-value) for a specific subset of variables and initialize variables for a now project.

## Install from source

Download and install go: https://golang.org/doc/install

Download code in your gopath:
```
$ git clone https://github.com/ingordigia/pargolo
```
Change directory:
```
$ cd $GOPATH/github.com/ingordigia/pargolo
```
Build:
```bash
$ go build
```

### Usage

#### AWS Authentication

AWS authentication can be done using an AWS profile.
Refer to AWS documentation for credential profile configuration: https://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html

```bash
$ ./pargolo.exe

--- searchbypath ---
  -output string
        (optional) Output CSV file
  -path string
        (required) prefix path to download
  -profile string
        (optional) AWS profile

--- searchbyvalue ---
  -filter string
        (optional) Filters the results by path
  -output string
        (optional) Output CSV file
  -profile string
        (optional) AWS profile
  -value string
        (required) The Value to search

--- upload ---
  -input string
        (required) Input CSV file
  -overwrite
        (optional) Overwrite the value if the key already exists
  -profile string
        (optional) AWS profile

--- export ---
  -domain string
        (required) The project domain
  -env string
        (required) The source environment
  -profile string
        (optional) AWS profile
  -project string
        (required) The project name

--- validate ---
  -env string
        (required) The target environment
  -input string
        (required) Input CSV file
  -profile string
        (optional) AWS profile

--- initialize ---
  -domain string
        (required) The project domain
  -env string
        (required) The source environment
  -input string
        (required) Input CSV file
  -profile string
        (optional) AWS profile
  -project string
        (required) The project name
```

#### Download parameters with "pargolo searchbypath"

With `pargolo searchbypath` you can print all parameters, with a specific prefix in their path, from AWS parameter store:
```sh
$ ./pargolo.exe searchbypath -path /my/prefix/path
```
or write them in a CSV file with the `-output` flag
```sh
$ ./pargolo.exe searchbypath -output localcsvname -path /my/prefix/path
```

#### Upload parameters from a local CSV with "pargolo upload"

When You need to upload a batch of parameters form a local CSV to the AWS parameter store you can use `pargolo upload` command.

```sh
$ ./pargolo.exe upload -input inputcsv -profile awsprofile
```
The default behavior, in case of an already existing key in the AWS parameter store, is to preserve it, but you can change this behavior with the `-overwrite` flag.
```sh
$ ./pargolo.exe upload -input inputcsv -overwrite true -profile awsprofile
```

#### Search parameters by value with "pargolo searchbyvalue"

Sometimes You just need to find all parameters with a specific value, in this case you can use `pargolo scrape` command.

```sh
$ ./pargolo searchbyvalue -value foobar -filter /path/to/search -profile awsprofile
```
`-output` is an additional optional flag that let you export the result in a CSV file.

#### Create a CSV file containing all project parameters with "pargolo export"

When you need to promote parameters from an environment to another you can use `pargolo export` command to download all project related parameters.

```sh
$ ./pargolo export -env envname -domain domainname -project projectname -profile awsprofile
```

#### Validate a CSV file containing all project parameters with "pargolo validate"

Before you upload a batch of parameters from a CSV, you can check the validity of you data against some basic rules with `pargolo validate`.

```sh
$ ./pargolo validate -env envname -input inputcsv -profile targetenvawsprofile
```
For each parameter found in your CSV file, pargolo will print one of the following possible results:

|Output|Description|
| --- | --- |
|MISSING -> CREATE|Project parameter or Common parameter is missing, uploading this CSV a new parameter will be created.|
|MISSING -> DUPLICATE|Common parameter is missing, but pargolo found another with the same value for you.|
|PRESENT -> MAINTAIN|Project parameter or Common parameter already exists, the upload will do nothing.|
|PRESENT -> DESTRUCTIVE|Common parameter already exists with a different value, if you upload this CSV with the overwrite option you can do serious damage.|
|PRESENT -> OVERWRITE|Project parameter already exists with a different value, when you upload this CSV you can choose if retain the present value or overwrite it.|

#### Create a CSV template starting from a JSON configuration file with "pargolo initialize"

When a new project starts is pretty annoying to create a whole new CSV in order to upload it with "pargolo upload", just let pargolo do it for you

```sh
$ ./pargolo initialize -env envname -domain domainname -project projectname -input .\config.json
```
