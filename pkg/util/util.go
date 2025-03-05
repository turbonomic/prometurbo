package util

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"
)

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

func ParseIP(addr string, default_port int) (string, string, error) {
	addr = strings.TrimSpace(addr)
	if len(addr) < 2 {
		return "", "", fmt.Errorf("illegal addr[%v]", addr)
	}

	items := strings.Split(addr, ":")
	if len(items) >= 2 {
		return items[0], items[1], nil
	}
	return items[0], fmt.Sprintf("%v", default_port), nil
}

func GetName(name, namespace string) string {
	if name == "" {
		return ""
	}
	if namespace == "" {
		return name
	}
	return fmt.Sprintf("%s-%s", name, namespace)
}

func GetDisplay(name, namespace string) string {
	if name == "" {
		return ""
	}
	if namespace == "" {
		return name
	}
	return fmt.Sprintf("%s [%s]", name, namespace)
}

// Utility function to return a pointer to the function argument.
// Useful for creating in-line primitive pointers, e.g. AsPtr(123), AsPtr(true)
func AsPtr[T any](t T) *T {
	return &t
}

func CloneMap[K comparable, V any](m map[K]V) map[K]V {
	r := map[K]V{}
	for k, v := range m {
		r[k] = v
	}
	return r
}
