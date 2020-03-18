package main

import (
	"context"
	"fmt"
	"github.com/baetyl/baetyl-function/utils"
	"net/http"
	"strings"

	baetyl "github.com/baetyl/baetyl-function/proto"
	"github.com/baetyl/baetyl-go/log"
	"github.com/docker/distribution/uuid"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

// Config
type Config struct {
	Server ServerConfig `yaml:"server" json:"server"`
}

type API struct {
	log       *log.Logger
	cfg       *Config
	manager   Manager
	endpoints []Endpoint
}

type Endpoint struct {
	Methods []string
	Route   string
	Handler func(c *routing.Context) error
}

func NewAPI(cfg Config) (*API, error) {
	m := NewGRPCManager()
	api := &API{
		log:     log.With(log.Any("main", "api")),
		cfg:     &cfg,
		manager: m,
	}
	api.endpoints = append(api.endpoints, api.constructFunctionEndpoints()...)
	api.endpoints = append(api.endpoints, api.constructServiceEndpoints()...)

	handler := api.useRouter()
	go func() {
		// TODO: support tls
		if err := fasthttp.ListenAndServe(cfg.Server.Address, handler); err != nil {
			panic(err)
		}
	}()

	return api, nil
}

// Close closes api
func (a *API) Close() {
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

func (a *API) constructFunctionEndpoints() []Endpoint {
	return []Endpoint{
		{
			Methods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
			Route:   "/baetyl-function/<service>",
			Handler: a.onFunctionMessage,
		},
		{
			Methods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
			Route:   "/baetyl-function/<service>/<method>",
			Handler: a.onFunctionMessage,
		},
	}
}

func (a *API) constructServiceEndpoints() []Endpoint {
	return []Endpoint{
		{
			Methods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
			Route:   "/*",
			Handler: a.onServiceMessage,
		},
	}
}

func (a *API) onServiceMessage(c *routing.Context) error {
	// TODO: http proxy
	fmt.Fprintf(c, "Hello World!")
	return nil
}

func (a *API) onFunctionMessage(c *routing.Context) error {
	serviceName := c.Param("service")
	method := c.Param("method")
	body := c.PostBody()

	metedata := map[string]string{
		"path":                  string(c.Request.URI().Path()),
		"httpMethod":            strings.ToUpper(string(c.Method())),
		"isBase64Encoded":       "false",
		"queryStringParameters": string(c.QueryArgs().QueryString()),
		"invokeId":              uuid.Generate().String(),
	}
	utils.SetHeaders(c, metedata)
	message := baetyl.Message{
		Name:     serviceName,
		Method:   method,
		Type:     "HTTP",
		Payload:  body,
		Metadata: metedata,
	}

	address := utils.ResolveAddress(serviceName)
	conn, err := a.manager.GetGRPCConnection(address, false)
	if err != nil {
		msg := NewErrorResponse("ERR_FUNCTION_CALL", err.Error())
		respondWithError(c.RequestCtx, 500, msg)
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.Timeout)
	defer cancel()

	client := baetyl.NewFunctionClient(conn)
	resp, err := client.Call(ctx, &message)
	if err != nil {
		msg := NewErrorResponse("ERR_FUNCTION_CALL", err.Error())
		respondWithError(c.RequestCtx, 500, msg)
	} else {
		statusCode := utils.GetStatusCodeFromMetadata(resp.Metadata)
		utils.SetHeadersOnRequest(resp.Metadata, c)
		respond(c.RequestCtx, statusCode, resp.Payload)
	}
	return nil
}
