package prometheus

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	hosts := []string{"127.0.0.1",
		"http://127.0.0.1",
		"https://127.0.0.1",
		"https://127.0.0.1:1900",
		"https://localhost",
		"https://localhost:8080",
	}

	for _, host := range hosts {
		_, err := NewRestClient(host)
		if err != nil {
			t.Errorf("Failed to create rest client: %v", err)
		}
	}
}
