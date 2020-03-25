package main

import (
	"fmt"
	"testing"

	"github.com/baetyl/baetyl-go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestServer(t *testing.T) {
	cfg := mockConfig(t)
	api := NewAPI(cfg)
	assert.NotEmpty(t, api)
	defer api.Close()

	req := fasthttp.AcquireRequest()
	req.SetRequestURI("http://localhost" + cfg.Server.Address)

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 404)

	url2 := fmt.Sprintf("%s%s%s", "http://localhost", cfg.Server.Address, "/test")
	req.SetRequestURI(url2)
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 404)

	url3 := fmt.Sprintf("%s%s%s", "http://localhost", cfg.Server.Address, "/book/create/detail")
	req.SetRequestURI(url3)
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 404)
}

func mockConfig(t *testing.T) Config {
	var cfg Config
	err := utils.UnmarshalYAML(nil, &cfg)
	assert.NoError(t, err)
	return cfg
}
