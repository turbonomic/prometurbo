package client

import (
	"encoding/json"
	"fmt"

	"github.com/turbonomic/turbo-api/pkg/api"
)

func parseAPIErrorDTO(errorMessage string) (*api.APIErrorDTO, error) {
	var dto api.APIErrorDTO
	err := json.Unmarshal([]byte(errorMessage), &dto)
	if err != nil {
		return nil, fmt.Errorf("Unmarshall error: %s", err)
	}
	return &dto, nil
}
