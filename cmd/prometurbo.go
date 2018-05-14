package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/pkg"
	"github.com/turbonomic/prometurbo/pkg/conf"
	"os"
)

func main() {
	// The default is to log to both of stderr and file
	// These arguments can be overloaded from the command-line args
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "/var/log")
	defer glog.Flush()

	args := conf.NewPrometurboArgs(flag.CommandLine)
	flag.Parse()

	glog.Infof("GIT_COMMIT: %s", os.Getenv("GIT_COMMIT"))
	glog.Info("Starting Prometurbo...")
	glog.V(3).Infof("Arguments: %++v", args)
	s, err := pkg.NewP8sTAPService(args)

	if err != nil {
		glog.Fatal("Failed creating Prometurbo: %v", err)
	}

	s.Start()
}
