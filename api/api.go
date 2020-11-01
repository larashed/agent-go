package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/larashed/agent-go/config"
)

type Api interface {
	SendServerMetrics(data string) (*Response, error)
	SendEnvironmentMetrics(data string) (*Response, error)
	//SendDeployment(data string) (*Response, error)
}

type Client struct {
	url       string
	apiKey    string
	apiSecret string
	env       string
	hostname  string
	client    http.Client
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewClient(url, env, key, secret, hostname string) *Client {
	return &Client{
		url:       url,
		env:       env,
		hostname:  hostname,
		apiKey:    key,
		apiSecret: secret,
		client: http.Client{
			Timeout: time.Second * 10, // Maximum of 2 secs
		},
	}
}

func (c *Client) SendServerMetrics(data string) (*Response, error) {
	return c.doRequest("POST", "agent/server/metrics", data)
}

func (c *Client) SendEnvironmentMetrics(data string) (*Response, error) {
	return c.doRequest("POST", "agent/environment/metrics", data)
}

func (c *Client) doRequest(method, url string, data string) (*Response, error) {
	req, err := http.NewRequest(
		method,
		strings.TrimRight(c.url, "/")+"/v1/"+url,
		bytes.NewBuffer([]byte(data)),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Larashed/Agent v1.0")
	req.Header.Set("Larashed-Environment", c.env)
	req.Header.Set("Larashed-Hostname", c.hostname)
	req.Header.Set("Larashed-Agent-Version", config.GitTag)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.apiKey, c.apiSecret)

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
