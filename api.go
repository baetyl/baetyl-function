package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	baetyl "github.com/baetyl/baetyl-go/faas"
	baetylhttp "github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/docker/distribution/uuid"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

const (
	UserNamespace = "baetyl-edge"
)

type API struct {
	log       *log.Logger
	cfg       *Config
	svr       *baetylhttp.Server
	manager   Manager
	endpoints []Endpoint
}

type Endpoint struct {
	Methods []string
	Route   string
	Handler func(c *routing.Context) error
}

func NewAPI(cfg Config) *API {
	m := NewManager(cfg.Client)
	api := &API{
		log:     log.With(log.Any("main", "api")),
		cfg:     &cfg,
		manager: m,
	}
	api.endpoints = append(api.endpoints, api.proxyEndpoints()...)

	handler := api.useRouter()
	api.svr = baetylhttp.NewServer(cfg.Server.ServerConfig, handler)
	api.svr.Start()
	return api
}

// Close closes api
func (a *API) Close() {
	if a.svr != nil {
		a.svr.Close()
	}
	if a.manager != nil {
		a.manager.Close()
	}
}

func (a *API) useRouter() fasthttp.RequestHandler {
	router := routing.New()

	for _, e := range a.endpoints {
		methods := strings.Join(e.Methods, ",")
		router.To(methods, e.Route, e.Handler)
	}

	return router.HandleRequest
}

func (a *API) proxyEndpoints() []Endpoint {
	return []Endpoint{
		{
			Methods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
			Route:   "/<service>",
			Handler: a.onHttpMessage,
		},
		{
			Methods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
			Route:   "/<service>/<function>",
			Handler: a.onHttpMessage,
		},
		{
			Methods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
			Route:   "/<service>/<function>/*",
			Handler: a.onServiceMessage,
		},
	}
}

func (a *API) onHttpMessage(c *routing.Context) error {
	service := c.Param("service")
	if service != "" {
		switch string(c.Host()) {
		case a.cfg.Server.Host.Function:
			return a.onFunctionMessage(c)
		case a.cfg.Server.Host.Service:
			return a.onServiceMessage(c)
		}
	}
	respondError(c, 404, "ERR_NO_ROUTE", "no route")
	return nil
}

func (a *API) onServiceMessage(c *routing.Context) error {
	uri := c.Request.URI()
	serviceName := c.Param("service")
	uri.SetHost(fmt.Sprintf("%s.%s", serviceName, UserNamespace))
	uri.SetPathBytes(uri.Path()[len(serviceName)+1:])

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	c.Request.CopyTo(req)
	client := a.manager.GetHttpClient()
	if err := client.Do(req, resp); err != nil {
		respondError(c, 500, "ERR_SERVICE_CALL", err.Error())
		return nil
	}

	resp.CopyTo(&c.Response)
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
	return nil
}

func (a *API) onFunctionMessage(c *routing.Context) error {
	serviceName := c.Param("service")
	functionName := c.Param("function")
	body := c.PostBody()

	invokeId := string(c.RequestCtx.Request.Header.Peek("invokeid"))
	if invokeId == "" {
		invokeId = uuid.Generate().String()
	}
	metedata := map[string]string{
		"serviceName":  serviceName,
		"functionName": functionName,
		"invokeId":     invokeId,
	}
	message := baetyl.Message{
		Payload:  body,
		Metadata: metedata,
	}

	address := fmt.Sprintf("%s.%s:%d", serviceName, UserNamespace, a.cfg.Client.Grpc.Port)
	conn, err := a.manager.GetGRPCConnection(address, false)
	if err != nil {
		respondError(c, 500, "ERR_FUNCTION_GRPC", err.Error())
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Client.Grpc.Timeout)
	defer cancel()

	client := baetyl.NewFunctionClient(conn)
	resp, err := client.Call(ctx, &message)
	if err != nil {
		respondError(c, 500, "ERR_FUNCTION_CALL", err.Error())
		return nil
	}
	respond(c, http.StatusOK, resp.Payload)
	return nil
}
