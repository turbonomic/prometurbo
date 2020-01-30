package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/service"
)

const (
	LocalDebugConfPath = "configs/prometurbo-config.json"
	DefaultConfPath    = "/etc/prometurbo/turbo.config"
	defaultEndpoint    = "http://localhost:8081/metrics"
)

type PrometurboConf struct {
	Communicator           *service.TurboCommunicationConfig `json:"communicationConfig,omitempty"`
	TargetConf             *PrometurboTargetConf             `json:"prometurboTargetConfig,omitempty"`
	MetricExporterEndpoint string                            `json:"metricExporterEndpoint,omitempty"`
	// Appended to the end of a probe name when registering with the platform. Useful when you need
	// multiple prometheus probe instances with affinity for discovering specific targets.
	TargetTypeSuffix *string `json:"ProbeNameSuffix,omitempty"`
}

type PrometurboTargetConf struct {
	Address string `json:"targetAddress,omitempty"`
	Scope   string `json:"scope,omitempty"`
}

func NewPrometurboConf(configFilePath string) (*PrometurboConf, error) {

	glog.Infof("Read configuration from %s", configFilePath)
	config, err := readConfig(configFilePath)

	if err != nil {
		return nil, err
	}

	if config.MetricExporterEndpoint == "" {
		config.MetricExporterEndpoint = defaultEndpoint
	}

	if config.Communicator == nil {
		return nil, fmt.Errorf("unable to read the turbo communication config from %s", configFilePath)
	}

	return config, nil
}

func readConfig(path string) (*PrometurboConf, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Errorf("File error: %v\n", err)
		return nil, err
	}
	glog.Infoln(string(file))

	var config PrometurboConf
	err = json.Unmarshal(file, &config)

	if err != nil {
		glog.Errorf("Unmarshall error :%v\n", err)
		return nil, err
	}
	glog.Infof("Results: %+v\n", config)

	return &config, nil
}
