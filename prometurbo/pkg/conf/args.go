package conf

import (
	"flag"
)

const (
	defaultDiscoveryIntervalSec = 600
	defaultKeepStandalone = false
	defaultCreateProxyVM = false
)

type PrometurboArgs struct {
	DiscoveryIntervalSec *int
	KeepStandalone *bool
	CreateProxyVM *bool
}

func NewPrometurboArgs(fs *flag.FlagSet) *PrometurboArgs {
	p := &PrometurboArgs{}

	p.DiscoveryIntervalSec = fs.Int("discovery-interval-sec", defaultDiscoveryIntervalSec, "The discovery interval in seconds")
	p.KeepStandalone = fs.Bool("keepStandalone", defaultKeepStandalone, "Do we keep non-stitched entities")
	p.CreateProxyVM = fs.Bool("createProxyVM", defaultCreateProxyVM, "Do we create a proxy VM for applications")

	return p
}
