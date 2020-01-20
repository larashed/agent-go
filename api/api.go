package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Api struct {
	url       string
	apiKey    string
	apiSecret string
	env       string
	client    http.Client
}

type Response struct {
	Success bool `json:"success"`
}

func NewApi(url, env, key, secret string) *Api {
	return &Api{
		url:       url,
		env:       env,
		apiKey:    key,
		apiSecret: secret,
		client: http.Client{
			Timeout: time.Second * 5, // Maximum of 2 secs
		},
	}
}

func (c *Api) SendServerMetrics(data interface{}) error {
	return c.doRequest("POST", "agent/server", data)
}

func (c *Api) SendRecords(data interface{}) error {
	return c.doRequest("POST", "agent/application", data)
}

func (c *Api) SendDeployment(data interface{}) error {
	return c.doRequest("POST", "agent/deployment", data)
}

func (c *Api) doRequest(method, url string, data interface{}) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, c.url+"/v1/"+url, bytes.NewBuffer(j))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "larashed/agent v1")
	req.Header.Set("Larashed-Environment", c.env)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, getErr := c.client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
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
