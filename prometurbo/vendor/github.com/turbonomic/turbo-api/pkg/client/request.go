package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/turbonomic/turbo-api/pkg/api"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Request struct {
	client HTTPClient

	verb    string
	baseURL *url.URL

	pathPrefix string
	params     url.Values

	basicAuth *BasicAuthentication

	resource     api.ResourceType
	resourceName string

	data    io.Reader
	headers map[string]string

	err error
}

type Result struct {
	statusCode int
	status     string
	body       string
	err        error
}

func NewRequest(client HTTPClient, verb string, baseURL *url.URL, apiPath string) *Request {
	if len(apiPath) != 0 && !strings.HasPrefix(apiPath, "/") {
		apiPath = path.Join("/", apiPath)
	}
	return &Request{
		client:     client,
		verb:       verb,
		baseURL:    baseURL,
		pathPrefix: apiPath,
	}
}

func (r *Request) BasicAuthentication(basicAuth *BasicAuthentication) *Request {
	r.basicAuth = basicAuth
	return r
}

// Set the kind of the api resource that the request is made to.
func (r *Request) Resource(resource api.ResourceType) *Request {
	if r.err != nil {
		return r
	}

	if r.resource != "" {
		r.err = errors.New("Resource has already been set. Cannot be changed!")
		return r
	}
	r.resource = resource
	return r
}

func (r *Request) Name(resourceName string) *Request {
	if r.err != nil {
		return r
	}

	if r.resourceName != "" {
		r.err = fmt.Errorf("Resource name has already been set to %s. Cannot be changed!", r.resourceName)
		return r
	}
	if len(resourceName) == 0 {
		r.err = errors.New("Resource name cannot be empty.")
		return r
	}
	r.resourceName = resourceName
	return r
}

// Set parameters for the request.
func (r *Request) Param(paramName, value string) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}
	r.params[paramName] = append(r.params[paramName], value)

	return r
}

func (r *Request) Header(key, value string) *Request {
	if r.err != nil {
		return r
	}
	if r.headers == nil {
		r.headers = map[string]string{}
	}
	r.headers[key] = value
	return r
}

func (r *Request) Data(data []byte) *Request {
	if r.err != nil {
		return r
	}
	r.data = bytes.NewBuffer(data)
	return r
}

// URL returns the current working URL.
func (r *Request) URL() *url.URL {
	p := r.pathPrefix

	if len(r.resource) != 0 {
		p = path.Join(p, strings.ToLower(string(r.resource)))
	}

	if len(r.resourceName) != 0 {
		p = path.Join(p, r.resourceName)
	}

	finalURL := &url.URL{}
	if r.baseURL != nil {
		*finalURL = *r.baseURL
	}
	finalURL.Path = p

	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	finalURL.RawQuery = query.Encode()

	return finalURL
}

func (r *Request) Do() (Result, error) {
	var result Result
	err := r.request(func(resp *http.Response) {
		result = parseHTTPResponse(resp)
	})
	if err != nil {
		return Result{}, err
	}
	if result.err != nil {
		return Result{}, err
	}
	return result, nil
}

// Perform the actual http request.
// fn is the function to parse http Response.
func (r *Request) request(fn func(*http.Response)) error {
	if r.err != nil {
		return r.err
	}

	requestURL := r.URL().String()
	req, err := http.NewRequest(r.verb, requestURL, r.data)
	if err != nil {
		return err
	}
	if r.headers != nil {
		for key, value := range r.headers {
			req.Header.Set(key, value)
		}
	}
	// Set basic authentication for the request if there is any.
	if r.basicAuth != nil {
		req.SetBasicAuth(r.basicAuth.username, r.basicAuth.password)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	fn(resp)
	return nil
}

func parseHTTPResponse(resp *http.Response) Result {
	if resp == nil {
		return Result{
			err: errors.New("response sent in is nil"),
		}
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Result{
			err: fmt.Errorf("Error reading response body: %++v", err),
		}
	}

	return Result{
		statusCode: resp.StatusCode,
		status:     resp.Status,
		body:       string(content),
		err:        nil,
	}
}
