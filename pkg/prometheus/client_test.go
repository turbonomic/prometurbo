package prometheus

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewClient(t *testing.T) {
	hosts := []struct {
		input          string
		expectedOutput string
	}{
		{"127.0.0.1", "http://127.0.0.1/api/v1/query"},
		{"http://127.0.0.1", "http://127.0.0.1/api/v1/query"},
		{"https://127.0.0.1", "https://127.0.0.1/api/v1/query"},
		{"https://127.0.0.1:1900", "https://127.0.0.1:1900/api/v1/query"},
		{"https://localhost", "https://localhost/api/v1/query"},
		{"https://localhost:8080", "https://localhost:8080/api/v1/query"},
		{"https://x.y.z/api/v2/promql/query", "https://x.y.z/api/v2/promql/query"},
	}

	for _, host := range hosts {
		testToken := "ABC"
		client, err := NewRestClient(host.input, testToken)
		if err != nil {
			t.Errorf("Failed to create rest client: %v", err)
		}
		assert.Equal(t, host.expectedOutput, client.host)
		assert.Equal(t, client.bearerToken, testToken)
	}
}
