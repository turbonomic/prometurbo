package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"

	"github.com/turbonomic/prometurbo/pkg/provider"
	"github.com/turbonomic/prometurbo/pkg/topology"
	"github.com/turbonomic/prometurbo/pkg/util"
	"github.com/turbonomic/prometurbo/pkg/worker"
)

type Server struct {
	port       int
	ip         string
	host       string
	provider   provider.MetricProvider
	topology   *topology.BusinessTopology
	dispatcher *worker.Dispatcher
}

const (
	metricPath = "/metrics"
)

func NewServer(port int) *Server {
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

	return &Server{
		port: port,
		ip:   ip,
		host: host,
	}
}

func (s *Server) MetricProvider(provider provider.MetricProvider) *Server {
	s.provider = provider
	return s
}

func (s *Server) Topology(topology *topology.BusinessTopology) *Server {
	s.topology = topology
	return s
}

func (s *Server) Dispatcher(dispatcher *worker.Dispatcher) *Server {
	s.dispatcher = dispatcher
	return s
}

func (s *Server) Run() {
	// Launch dispatcher to dispatch discovery tasks
	s.dispatcher.Start()
	// Start the http server to process discovery request
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s,
	}
	glog.V(2).Infof("HTTP server listens on: %v:%v", s.ip, s.port)
	panic(server.ListenAndServe())
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	glog.V(4).Infof("Begin to handle path: %v", path)

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
