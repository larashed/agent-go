package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type Api interface {
	SendServerMetrics(data interface{}) error
	SendApplicationRecords(data string) error
	SendDeployment(data interface{}) error
}

type Client struct {
	url       string
	apiKey    string
	apiSecret string
	env       string
	client    http.Client
}

type Response struct {
	Success bool `json:"success"`
}

func NewClient(url, env, key, secret string) *Client {
	return &Client{
		url:       url,
		env:       env,
		apiKey:    key,
		apiSecret: secret,
		client: http.Client{
			Timeout: time.Second * 5, // Maximum of 2 secs
		},
	}
}

func (c *Client) SendServerMetrics(data interface{}) error {
	return c.doRequest("POST", "agent/server", data)
}

func (c *Client) SendApplicationRecords(data string) error {
	return c.doRequest("POST", "agent/application", data)
}

func (c *Client) SendDeployment(data interface{}) error {
	return c.doRequest("POST", "agent/deployment", data)
}

func (c *Client) doRequest(method, url string, data interface{}) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, c.url+"/v1/"+url, bytes.NewBuffer(j))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Larashed/Agent v1.0")
	req.Header.Set("Larashed-Environment", c.env)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	response := &Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return nil
}
