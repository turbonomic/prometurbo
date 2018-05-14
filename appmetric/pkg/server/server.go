package server

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"os"
	"strings"

	"github.com/turbonomic/prometurbo/appmetric/pkg/alligator"
	"github.com/turbonomic/prometurbo/appmetric/pkg/util"
)

type MetricServer struct {
	port int
	ip   string
	host string

	appClient  *alligator.Alligator
	vappClient *alligator.Alligator
}

const (
	appMetricPath     = "/pod/metrics"
	serviceMetricPath = "/service/metrics"
	fakeMetricPath    = "/fake/metrics"
)

func NewMetricServer(port int, appClient, vappclient *alligator.Alligator) *MetricServer {
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
	glog.V(2).Infof("Will server on %s:%d", ip, port)

	return &MetricServer{
		port:       port,
		ip:         ip,
		host:       host,
		appClient:  appClient,
		vappClient: vappclient,
	}
}

func (s *MetricServer) Run() {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s,
	}

	glog.V(1).Infof("HTTP server listens on: %s", server.Addr)
	panic(server.ListenAndServe())
}

func (s *MetricServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	glog.V(2).Infof("Begin to handle path: %v", path)

	if strings.EqualFold(path, "/favicon.ico") {
		s.faviconHandler(w, r)
		return
	}

	if strings.EqualFold(path, appMetricPath) {
		s.handleAppMetric(w, r)
		return
	}

	if strings.EqualFold(path, serviceMetricPath) {
		s.handleServiceMetric(w, r)
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
