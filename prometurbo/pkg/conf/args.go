package conf

import (
	"flag"

	"github.com/turbonomic/prometurbo/prometurbo/appmetric"
)

const (
	defaultDiscoveryIntervalSec = 600
	defaultKeepStandalone       = false
	defaultCreateProxyVM        = false
)

type PrometurboArgs struct {
	DiscoveryIntervalSec *int
	KeepStandalone       *bool
	CreateProxyVM        *bool
	AppmetricArgs        appmetric.Args
}

func NewPrometurboArgs(fs *flag.FlagSet) *PrometurboArgs {
	p := &PrometurboArgs{}

	p.DiscoveryIntervalSec = fs.Int("discovery-interval-sec", defaultDiscoveryIntervalSec, "The discovery interval in seconds")
	p.KeepStandalone = fs.Bool("keepStandalone", defaultKeepStandalone, "Do we keep non-stitched entities")
	p.CreateProxyVM = fs.Bool("createProxyVM", defaultCreateProxyVM, "Do we create a proxy VM for applications")

	fs.Var(&p.AppmetricArgs.PrometheusHosts, "promUrl", "the addresses of prometheus servers. Supply at least one.")
	fs.StringVar(&p.AppmetricArgs.ConfigFileName, "config", "", "path of the config file")

	return p
}
