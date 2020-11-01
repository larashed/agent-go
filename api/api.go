package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/larashed/agent-go/config"
)

// Api client interface
type Api interface { //nolint:golint
	SendServerMetrics(data string) (*Response, error)
	SendEnvironmentMetrics(data string) (*Response, error)
	//SendDeployment(data string) (*Response, error)
}

// Client holds the API Client
type Client struct {
	config *config.Config
	client http.Client
}

// Response object structure
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewClient creates a new instance of `Client`
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		client: http.Client{
			Timeout: time.Second * 10, // Maximum of 2 secs
		},
	}
}

// SendServerMetrics sends collected server metrics to our API
func (c *Client) SendServerMetrics(data string) (*Response, error) {
	return c.doRequest("POST", "agent/server/metrics", data)
}

// SendEnvironmentMetrics sends collected app metrics to our API
func (c *Client) SendEnvironmentMetrics(data string) (*Response, error) {
	return c.doRequest("POST", "agent/environment/metrics", data)
}

func (c *Client) doRequest(method, url string, data string) (*Response, error) {
	req, err := http.NewRequest(
		method,
		strings.TrimRight(c.config.ApiUrl, "/")+"/v1/"+url,
		bytes.NewBuffer([]byte(data)),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Larashed/Agent v1.0")
	req.Header.Set("Larashed-Environment", c.config.AppEnvironment)
	req.Header.Set("Larashed-Hostname", c.config.Hostname)
	req.Header.Set("Larashed-GoAgent-Version", config.GitTag)
	req.Header.Set("Larashed-In-Docker", strconv.FormatBool(c.config.InDocker))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.config.AppId, c.config.AppKey)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	response := &Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New(response.Message)
	}

	return response, nil
}
