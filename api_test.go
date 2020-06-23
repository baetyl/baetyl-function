package main

import (
	"fmt"
	"testing"

	"github.com/baetyl/baetyl-go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestServer0(t *testing.T) {
	cfg := mockConfig1()
	api := NewAPI(cfg)
	assert.NotEmpty(t, api)
	defer api.Close()

	api2 := NewAPI(cfg)
	assert.NotEmpty(t, api2)
	defer api2.Close()
}

func TestServer1(t *testing.T) {
	cfg := mockConfig1()
	api := NewAPI(cfg)
	assert.NotEmpty(t, api)
	defer api.Close()

	req := fasthttp.AcquireRequest()
	req.SetRequestURI("http://localhost:50050")
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 404)

	url2 := fmt.Sprintf("%s%s", "http://localhost:50050", "/test")
	req.SetRequestURI(url2)
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 404)

	url3 := fmt.Sprintf("%s%s", "http://localhost:50050", "/book/create/detail")
	req.SetRequestURI(url3)
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 500)
}

func mockConfig1() Config {
	var cfg Config
	utils.UnmarshalYAML(nil, &cfg)
	cfg.Server.Address = ":50050"
	return cfg
}
