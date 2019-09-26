package main

import (
	"flag"
	"fmt"
	"github.com/diegoavanzini/pargolo/domains"
	"github.com/diegoavanzini/pargolo/repositories"
	"os"
)

// Parameters is  a map of parameter names and values
type Parameters map[string]string

var profile, path, output, input, value, env, domain, filter, project string
var overwrite, recursive bool

var searchbypath *flag.FlagSet
var upload *flag.FlagSet
var searchbyvalue *flag.FlagSet
var export *flag.FlagSet
var validate *flag.FlagSet
var initialize *flag.FlagSet

func main() {
	flag.Parse()

	if len(os.Args) < 2 {

		fmt.Printf("\n--- searchbypath ---\n")
		searchbypath.PrintDefaults()

		fmt.Printf("\n--- searchbyvalue ---\n")
		searchbyvalue.PrintDefaults()

		fmt.Printf("\n--- upload ---\n")
		upload.PrintDefaults()

		fmt.Printf("\n--- export ---\n")
		export.PrintDefaults()

		fmt.Printf("\n--- validate ---\n")
		validate.PrintDefaults()

		fmt.Printf("\n--- initialize ---\n")
		initialize.PrintDefaults()

		os.Exit(0)
	}

	switch os.Args[1] {
	case "searchbypath":
		searchbypath.Parse(os.Args[2:])
		if path == "" {
			searchbypath.PrintDefaults()
			os.Exit(1)
		}
		buildPargolo().DownloadParametersByPath(path, recursive)

	case "upload":
		upload.Parse(os.Args[2:])
		if input == "" {
			upload.PrintDefaults()
			os.Exit(1)
		}

		buildPargolo().UploadParametersFromCsv(input, overwrite)

	case "searchbyvalue":
		searchbyvalue.Parse(os.Args[2:])
		if value == "" {
			searchbyvalue.PrintDefaults()
			os.Exit(1)
		}

		buildPargolo().DownloadParametersByValue(value, filter)

	case "export":
		export.Parse(os.Args[2:])
		if env == "" || domain == "" || project == "" {
			export.PrintDefaults()
			os.Exit(1)
		}

		buildPargolo().ExportParameters(env, domain, project)

	case "validate":
		validate.Parse(os.Args[2:])
		if input == "" || env == "" {
			validate.PrintDefaults()
			os.Exit(1)
		}

		buildPargolo().ValidateParameters(input, env)

	case "initialize":
		initialize.Parse(os.Args[2:])
		if input == "" || env == "" || domain == "" || project == "" {
			initialize.PrintDefaults()
			os.Exit(1)
		}

		buildPargolo().InitializeParameters(input, env, domain, project)

	default:
		flag.PrintDefaults()
		os.Exit(1)

	}
}

func buildPargolo() domains.IPargolo {
	repo, err := repositories.NewRepository(nil, &profile, nil)
	if err != nil {
		os.Exit(-1)
	}
	pargolo, err := domains.NewPargolo(repo)
	if err != nil {
		os.Exit(-1)
	}
	return pargolo
}

func init() {
	searchbypath = flag.NewFlagSet("SearchByPath", flag.ExitOnError)
	searchbypath.StringVar(&domains.Profile, "profile", "", "(optional) AWS profile")
	searchbypath.StringVar(&path, "path", "", "(required) prefix path to download")
	searchbypath.StringVar(&output, "output", "", "(optional) Output CSV file")
	searchbypath.BoolVar(&recursive, "recursive", false, "(optional) Select if pargolo should recursively resolve parameters value")
	searchbyvalue = flag.NewFlagSet("SearchByValue", flag.ExitOnError)
	searchbyvalue.StringVar(&domains.Profile, "profile", "", "(optional) AWS profile")
	searchbyvalue.StringVar(&value, "value", "", "(required) The Value to search")
	searchbyvalue.StringVar(&filter, "filter", "", "(optional) Filters the results by path")
	searchbyvalue.StringVar(&output, "output", "", "(optional) Output CSV file")
	upload = flag.NewFlagSet("Upload", flag.ExitOnError)
	upload.StringVar(&domains.Profile, "profile", "", "(optional) AWS profile")
	upload.StringVar(&input, "input", "", "(required) Input CSV file")
	upload.BoolVar(&overwrite, "overwrite", false, "(optional) Overwrite the value if the key already exists")
	export = flag.NewFlagSet("Export", flag.ExitOnError)
	export.StringVar(&domains.Profile, "profile", "", "(optional) AWS profile")
	export.StringVar(&env, "env", "", "(required) The source environment")
	export.StringVar(&domain, "domain", "", "(required) The project domain")
	export.StringVar(&project, "project", "", "(required) The project name")
	validate = flag.NewFlagSet("Validate", flag.ExitOnError)
	validate.StringVar(&domains.Profile, "profile", "", "(optional) AWS profile")
	validate.StringVar(&input, "input", "", "(required) Input CSV file")
	validate.StringVar(&env, "env", "", "(required) The target environment")
	initialize = flag.NewFlagSet("Initialize", flag.ExitOnError)
	initialize.StringVar(&domains.Profile, "profile", "", "(optional) AWS profile")
	initialize.StringVar(&input, "input", "", "(required) Input JSON config file")
	initialize.StringVar(&env, "env", "", "(required) The source environment")
	initialize.StringVar(&domain, "domain", "", "(required) The project domain")
	initialize.StringVar(&project, "project", "", "(required) The project name")
}
