package conf

import (
	"flag"
)

const (
	defaultDiscoveryIntervalSec = 600
)

type PrometurboArgs struct {
	DiscoveryIntervalSec *int
}

func NewPrometurboArgs(fs *flag.FlagSet) *PrometurboArgs {
	p := &PrometurboArgs{}

	p.DiscoveryIntervalSec = fs.Int("discovery-interval-sec", defaultDiscoveryIntervalSec, "The discovery interval in seconds")

	return p
}
