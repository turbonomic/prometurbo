package client

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/turbonomic/turbo-api/pkg/api"
)

type Client struct {
	*RESTClient
	SessionCookie *http.Cookie
}

const (
	SessionCookie string = "JSESSIONID"
)

// Create a Turbo API Client based on basic authentication.
func NewAPIClientWithBA(c *Config) (*Client, error) {
	if c.basicAuth == nil {
		return nil, errors.New("Basic authentication is not set")
	}
	client := http.DefaultClient
	// If use https, disable the security check.
	if c.serverAddress.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}
	restClient := NewRESTClient(client, c.serverAddress, c.apiPath).BasicAuthentication(c.basicAuth)
	return &Client{restClient, nil}, nil
}

// Login to the Turbo API server
func (c *Client) Login() (*Result, error) {
	var data []byte
	data = []byte(fmt.Sprintf("username=%s&password=%s", c.basicAuth.username, c.basicAuth.password))
	request := c.Post().Resource("login").
		Header("Content-Type", "application/x-www-form-urlencoded").
		Data(data)

	response, err := request.Do()
	if err != nil {
		return nil, fmt.Errorf("Failed to login  %s: %s", c.baseURL, err)
	}
	if response.statusCode != 200 {
		return nil, buildResponseError("Turbo server login", response.status, response.body)
	}

	// Save the session cookie
	sessionCookie, ok := response.cookies[SessionCookie]
	if ok {
		c.SessionCookie = sessionCookie
		glog.V(4).Infof("Session Cookie = %s:%s\n", c.SessionCookie.Name, c.SessionCookie.Value)
	} else {
		return nil, buildResponseError("Invalid session cookie", response.status, fmt.Sprintf("%s", response.cookies))
	}
	return &response, nil
}

// Discover a target using API
func (c *Client) DiscoverTarget(uuid string) (*Result, error) {
	response, err := c.Post().Resource(api.Resource_Type_Target).Name(uuid).Do()
	if err != nil {
		return nil, fmt.Errorf("Failed to discover target %s: %s", uuid, err)
	}
	if response.statusCode != 200 {
		return nil, buildResponseError("target discovery", response.status, response.body)
	}
	return &response, nil
}

//Add a target using API
func (c *Client) AddTarget(target *api.Target) (*Result, error) {
	if c.SessionCookie == nil {
		return nil, fmt.Errorf("Null login session cookie\n")
	}

	// Find if the target exists - this is a workaround since in current XL server,
	// duplicate targets can be added
	targetExists, _ := c.FindTarget(target)
	if targetExists {
		return nil, fmt.Errorf("Target %v exists", target)
	}

	glog.V(2).Infof("***************** [AddTarget] %++v\n", target)

	// Create the rest api request
	targetData, err := json.Marshal(target)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshall target instance: %s", err)
	}

	request := c.Post().Resource(api.Resource_Type_Target).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Header("Cookie", fmt.Sprintf("%s=%s", c.SessionCookie.Name, c.SessionCookie.Value)).
		Data(targetData)

	glog.V(4).Infof("[AddTarget] Request %++v\n", request)

	response, err := request.Do()
	if err != nil {
		return nil, fmt.Errorf("Failed to add target: %s", err)
	}
	if response.statusCode != 200 {
		return nil, buildResponseError("target addition", response.status, response.body)
	}

	return &response, nil
}

//Find if the given target exists within the Turbo server
func (c *Client) FindTarget(target *api.Target) (bool, error) {
	// Get a list of targets from the Turbo server
	request := c.Get().Resource(api.Resource_Type_Target).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Header("Cookie", fmt.Sprintf("%s=%s", c.SessionCookie.Name, c.SessionCookie.Value))

	glog.V(4).Infof("[FindTarget] Request %++v\n", request)

	response, err := request.Do()

	if err != nil {
		fmt.Printf("Failed to execute find target request: %s", err)
		return false, fmt.Errorf("Failed to execute find target request: %s", err)
	}

	if response.statusCode != 200 {
		return false, buildResponseError("find target", response.status, response.body)
	}

	// The target identifier for the given target
	targetId := getTargetId(target)

	// Parse the response - list of targets
	var targetList []interface{}
	json.Unmarshal([]byte(response.body), &targetList)

	// Iterate over the list of targets to look for the given target
	// by comparing the category, target type and identifier fields
	for _, tgt := range targetList {
		targetInstance, isMap := tgt.(map[string]interface{})
		if !isMap {
			continue
		}
		var category, targetType, tgtId string
		category, _ = targetInstance["category"].(string)
		targetType, _ = targetInstance["type"].(string)
		inputFields, ok := targetInstance["inputFields"].([]interface{})
		if ok {
			// array of InputFields
			for _, value := range inputFields {
				// an instance of InputField
				inputFieldMap, isMap := value.(map[string]interface{})
				// Get the target identifier value from the
				// input field named 'targetIdentifier'
				if isMap {
					field, isNameField := inputFieldMap["name"].(string)
					//
					if isNameField && field == "targetIdentifier" {
						tgtId, _ = inputFieldMap["value"].(string)
					}
				}
			}
		}
		glog.V(4).Infof("%s::%s::%s\n", category, targetType, tgtId)
		if target.Category == category && target.Type == targetType && tgtId == targetId {
			return true, nil
		}
	}
	return false, nil
}

// Get the target identifier for the given target
func getTargetId(target *api.Target) string {
	for _, inputField := range target.InputFields {
		field := inputField.Name
		if field == "targetIdentifier" {
			tgtId := inputField.Value
			return tgtId
		}
	}
	return ""
}

func buildResponseError(requestDesc string, status string, content string) error {
	errorMsg := fmt.Sprintf("unsuccessful %s response: %s.", requestDesc, status)
	errorDTO, err := parseAPIErrorDTO(content)
	if err == nil && errorDTO.Message != "" {
		// Add error message only if we can parse result content to errorDTO.
		errorMsg = errorMsg + fmt.Sprintf(" %s.", errorDTO.Message)
	}
	return fmt.Errorf("%s", errorMsg)
}
