package config

import (
	"fmt"
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// BusinessApplicationConf defines a list of BusinessApplication
type BusinessApplicationConf struct {
	BusinessApplications []BusinessApplication `yaml:"businessApplications"`
}

// BusinessApplication defines a business application and its associated business transactions and services
type BusinessApplication struct {
	Name         string        `yaml:"name"`         // Name of the business application
	From         string        `yaml:"from"`         // Discovering source, i.e., the target URL
	Transactions []Transaction `yaml:"transactions"` // A list of optional business transactions
	Services     []string      `yaml:"services"`     // A list of services that the business application depends on
}

// Transaction defines a business transaction
type Transaction struct {
	Name     string   `yaml:"name"`     // The name of the business transaction
	Path     string   `yaml:"path"`     // The request path of the business transaction
	DependOn []string `yaml:"dependOn"` // A list of services that the business transaction depends on
}

// A configuration map by source and name of business applications
type BusinessAppConfBySource map[string]BusinessAppConfByName
type BusinessAppConfByName map[string]*BusinessApplication

func NewBusinessApplicationConfig(path string) (BusinessAppConfBySource, error) {
	glog.Infof("Read business application configuration from %s", path)
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %v: %v", path, err)
	}
	glog.Infof("%v", string(file))

	var bizAppConf BusinessApplicationConf

	if err := yaml.UnmarshalStrict(file, &bizAppConf); err != nil {
		return nil, fmt.Errorf("failed to unmarshall file %v: %v", path, err)
	}

	if len(bizAppConf.BusinessApplications) < 1 {
		return nil, fmt.Errorf("no business application is configured in file %v", path)
	}

	var bizAppConfBySource = BusinessAppConfBySource{}
	for _, bizApp := range bizAppConf.BusinessApplications {
		if err := validate(bizApp); err != nil {
			return nil, err
		}
		// Name and Source combination uniquely identifies a business application.
		// There cannot be two configured business applications with the same name and source.
		bizAppConfByName, ok := bizAppConfBySource[bizApp.From]
		if !ok {
			bizAppConfByName = make(map[string]*BusinessApplication)
			bizAppConfBySource[bizApp.From] = bizAppConfByName
		} else if _, found := bizAppConfByName[bizApp.Name]; found {
			return nil, fmt.Errorf("business app %v has alread been defined for source %v",
				bizApp.Name, bizApp.From)
		}
		bizAppCopy := bizApp
		bizAppConfByName[bizApp.Name] = &bizAppCopy
	}
	return bizAppConfBySource, nil
}

func validate(bizApp BusinessApplication) error {
	if bizApp.Name == "" {
		return fmt.Errorf("missing business application name")
	}
	if bizApp.From == "" {
		return fmt.Errorf("missing discovering source for business application %v", bizApp.Name)
	}
	if len(bizApp.Services) < 1 {
		return fmt.Errorf("no service is configured for business application %v", bizApp.Name)
	}
	for _, transaction := range bizApp.Transactions {
		if transaction.Path == "" {
			return fmt.Errorf("one or more transaction paths are empty for business application %v",
				bizApp.Name)
		}
	}
	return nil
}
