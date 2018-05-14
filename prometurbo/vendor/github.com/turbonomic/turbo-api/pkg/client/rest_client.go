package client

import (
	"net/http"
	"net/url"
)

type RESTClient struct {
	client *http.Client

	baseURL *url.URL
	apiPath string

	basicAuth *BasicAuthentication
}

func NewRESTClient(client *http.Client, baseURL *url.URL, apiPath string) *RESTClient {
	return &RESTClient{
		client:  client,
		baseURL: baseURL,
		apiPath: apiPath,
	}
}

func (c *RESTClient) BasicAuthentication(ba *BasicAuthentication) *RESTClient {
	c.basicAuth = ba
	return c
}

// Built request based on http verb and authentication.
func (c *RESTClient) Verb(verb string) *Request {
	request := NewRequest(c.client, verb, c.baseURL, c.apiPath)
	// Set basic authentication if there it is not null.
	if c.basicAuth != nil {
		request.BasicAuthentication(c.basicAuth)
	}
	return request
}

func (c *RESTClient) Get() *Request {
	return c.Verb("GET")
}

func (c *RESTClient) Post() *Request {
	return c.Verb("POST")
}

func (c *RESTClient) Delete() *Request {
	return c.Verb("DELETE")
}
