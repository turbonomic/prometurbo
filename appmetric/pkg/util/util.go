package util

import (
	"fmt"
	"github.com/golang/glog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func TimeTrack(start time.Time, name string) time.Duration {
	elapsed := time.Since(start)
	glog.V(2).Infof("%s took %s", name, elapsed)
	return elapsed
}

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("are you connected to the network?")
}

func GetClientIP(r *http.Request) string {
	return r.RemoteAddr
}

func GetOriginalClientInfo(r *http.Request) string {
	orig := r.Header.Get("X-Forwarded-For")
	glog.V(3).Infof("request from %v, %v", r.RemoteAddr, orig)

	if len(orig) > 0 {
		ips := strings.Split(orig, ", ")
		return ips[0]
	}

	return ""
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
