package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/provider"

	"github.com/turbonomic/prometurbo/appmetric/pkg/util"
)

type MetricServer struct {
	port int
	ip   string
	host string

	providerFactory ProviderCreator
}

type ProviderCreator interface {
	NewProvider(promHost string) (*provider.Provider, error)
}

const (
	metricPath = "/metrics"
)

func NewMetricServer(port int, providerFactory ProviderCreator) *MetricServer {
	ip, err := util.ExternalIP()
	if err != nil {
		glog.Errorf("Failed to get server IP: %v", err)
		ip = "localhost"
	}

	host, err := os.Hostname()
	if err != nil {
		glog.Errorf("Failed to get hostname: %v", err)
		host = "localhost"
	}

	return &MetricServer{
		port:            port,
		ip:              ip,
		host:            host,
		providerFactory: providerFactory,
	}
}

func (s *MetricServer) Run() {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s,
	}

	glog.V(2).Infof("HTTP server listens on: %v:%v", s.ip, s.port)
	panic(server.ListenAndServe())
}

func (s *MetricServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	glog.V(2).Infof("Begin to handle path: %v", path)

	if strings.EqualFold(path, "/favicon.ico") {
		s.faviconHandler(w, r)
		return
	}

	if strings.EqualFold(path, metricPath) {
		s.handleMetric(w, r)
		return
	}

	s.handleWelcome(path, w, r)
	return
}
