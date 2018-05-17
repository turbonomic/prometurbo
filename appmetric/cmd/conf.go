package main

import (
	"encoding/json"
	"github.com/golang/glog"
	"io/ioutil"
)

type metricConf struct {
	Address string `json:"targetAddress,omitempty"`
	Port    string `json:"metricPort,omitempty"`
}

type wrapConf struct {
	MConf *metricConf `json:"prometurboTargetConfig,omitempty"`
}

func readConfig(fname string) (*metricConf, error) {
	glog.V(2).Infof("Reading config file: %v", fname)
	content, err := ioutil.ReadFile(fname)
	if err != nil {
		glog.Errorf("Failed to read config file(%v): %v", fname, err)
		return nil, err
	}
	glog.V(3).Infof("Config content: %v", string(content))

	var config wrapConf
	err = json.Unmarshal(content, &config)

	if err != nil {
		glog.Errorf("Unmarshall error :%v", err)
		return nil, err
	}
	glog.V(3).Infof("Configure results: %+v", config.MConf)

	return config.MConf, nil
}
