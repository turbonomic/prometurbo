package server

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/provider"
	"net/http"
	"os"
	"strings"

	"github.com/turbonomic/prometurbo/appmetric/pkg/util"
)

type MetricServer struct {
	port int
	ip   string
	host string

	provider *provider.Provider
}

const (
	metricPath        = "/metrics"
	appMetricPath     = "/pod/metrics"
	serviceMetricPath = "/service/metrics"
	fakeMetricPath    = "/fake/metrics"
)

func NewMetricServer(port int, provider *provider.Provider) *MetricServer {
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
		port:       port,
		ip:         ip,
		host:       host,
		provider:   provider,
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

	if strings.EqualFold(path, fakeMetricPath) {
		s.handleFakeMetric(w, r)
		return
	}

	//if strings.EqualFold(path, "/health") {
	//	s.handleHealth(w, r)
	//}

	s.handleWelcome(path, w, r)
	return
}
