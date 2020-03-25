package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	baetyl "github.com/baetyl/baetyl-go/faas"
	"github.com/baetyl/baetyl-go/utils"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
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

func TestServer2(t *testing.T) {
	cfg := mockConfig2()
	api := NewAPI(cfg)
	assert.NotEmpty(t, api)
	defer api.Close()
	mockHttp(t)

	// wait http server start
	time.Sleep(time.Second)

	req := fasthttp.AcquireRequest()
	url := "http://127.0.0.1:50051"
	req.SetRequestURI(url)

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 404)

	url2 := fmt.Sprintf("%s/%s", "http://127.0.0.1:50051", "127.0.0.1:8523/angthing")
	req.SetRequestURI(url2)
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 200)
	assert.Equal(t, "Hello One!", string(resp.Body()))

	url3 := fmt.Sprintf("%s/%s", "http://127.0.0.1:50051", "127.0.0.1:8523/anything/anything")
	req.SetRequestURI(url3)
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 200)
	assert.Equal(t, "Hello Two!", string(resp.Body()))

	url4 := fmt.Sprintf("%s/%s", "http://127.0.0.1:50051", "127.0.0.1:8523/anything/anything/anything")
	req.SetRequestURI(url4)
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 200)
	assert.Equal(t, "Hello Three!", string(resp.Body()))
}

func TestServer3(t *testing.T) {
	cfg := mockConfig3()
	api := NewAPI(cfg)
	assert.NotEmpty(t, api)
	defer api.Close()
	mockGrpc(t)

	// wait http server start
	time.Sleep(time.Second)

	req := fasthttp.AcquireRequest()
	url := "http://0.0.0.0:50052"
	req.SetRequestURI(url)

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 404)

	url2 := fmt.Sprintf("%s/%s", "http://0.0.0.0:50052", "127.0.0.1")
	req.SetRequestURI(url2)
	req.SetBody([]byte("Hello Grpc"))
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 200)
	assert.Equal(t, "Hello Grpc", string(resp.Body()))

	url3 := fmt.Sprintf("%s/%s", "http://0.0.0.0:50052", "127.0.0.1/angthing")
	req.SetRequestURI(url3)
	req.SetBody([]byte("Hello Grpc 2"))
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 200)
	assert.Equal(t, "Hello Grpc 2", string(resp.Body()))

	url4 := fmt.Sprintf("%s/%s", "http://0.0.0.0:50052", "127.0.0.1/angthing")
	req.SetRequestURI(url4)
	req.SetBody([]byte("Hello err"))
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 500)

	url5 := fmt.Sprintf("%s/%s", "http://0.0.0.0:50052", "127.0.0.1/angthing")
	req.SetRequestURI(url5)
	req.SetBody([]byte("json"))
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 200)
	assert.Equal(t, string(resp.Header.ContentType()), "application/json")
}

func mockConfig1() Config {
	var cfg Config
	utils.UnmarshalYAML(nil, &cfg)
	cfg.Server.Address = ":50050"
	return cfg
}

func mockConfig2() Config {
	var cfg Config
	utils.UnmarshalYAML(nil, &cfg)
	cfg.Server.Address = ":50051"
	cfg.Server.Host.Service = "127.0.0.1:50051"
	cfg.Client.Grpc.Port = 50010
	return cfg
}

func mockConfig3() Config {
	var cfg Config
	utils.UnmarshalYAML(nil, &cfg)
	cfg.Server.Address = ":50052"
	cfg.Server.Host.Function = "0.0.0.0:50052"
	cfg.Client.Grpc.Port = 50010
	return cfg
}

type mockGrpcServer struct{}

func (m *mockGrpcServer) Call(ctx context.Context, msg *baetyl.Message) (*baetyl.Message, error) {
	body := string(msg.Payload)
	if body == "Hello err" {
		return nil, errors.New("err")
	} else if body == "json" {
		o := map[string]string{
			"name": "baetyl",
		}
		s, _ := json.Marshal(o)
		msg.Payload = s
	}
	return msg, nil
}

func mockGrpc(t *testing.T) {
	lis, err := net.Listen("tcp", ":50010")
	assert.NoError(t, err)
	s := grpc.NewServer()
	baetyl.RegisterFunctionServer(s, new(mockGrpcServer))
	go func() {
		err = s.Serve(lis)
		assert.NoError(t, err)
	}()
}

func mockHttp(t *testing.T) {
	router := routing.New()
	router.Get("/<service>/<anything>/<anything>", func(c *routing.Context) error {
		fmt.Fprintf(c, "Hello Three!")
		return nil
	})
	router.Get("/<service>/<anything>", func(c *routing.Context) error {
		fmt.Fprintf(c, "Hello Two!")
		return nil
	})
	router.Get("/<service>", func(c *routing.Context) error {
		fmt.Fprintf(c, "Hello One!")
		return nil
	})

	go func() {
		err := fasthttp.ListenAndServe(":8523", router.HandleRequest)
		assert.NoError(t, err)
	}()
}
