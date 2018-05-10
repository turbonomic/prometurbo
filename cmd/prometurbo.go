package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg"
)

func main() {
	flag.Parse()

	// The default is to log to both of stderr and file
	// These arguments can be overloaded from the command-line args
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "/var/log")
	defer glog.Flush()

	glog.Info("Starting prometurbo...")

	s := pkg.P8sTAPService{}
	s.Start()
}
