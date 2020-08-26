package function

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	context2 "github.com/baetyl/baetyl-go/v2/context"
	baetyl "github.com/baetyl/baetyl-go/v2/faas"
	baetylhttp "github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/docker/distribution/uuid"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

var edgeNamespace = "baetyl-edge"

type API struct {
	cfg          *Config
	svr          *baetylhttp.Server
	manager      Manager
	endpoints    []Endpoint
	functionHost string
	serviceHost  string
	log          *log.Logger
}

type Endpoint struct {
	Methods []string
	Route   string
	Handler func(c *routing.Context) error
}

func NewAPI(cfg Config, ctx context2.Context) (*API, error) {
	cert := ctx.SystemConfig().Certificate
	m, err := NewManager(cfg.Client, cert)
	if err != nil {
		return nil, err
	}

	api := &API{
		cfg:          &cfg,
		manager:      m,
		functionHost: fmt.Sprintf("%s:%s", cfg.Server.Host.Function, ctx.FunctionHttpPort()),
		serviceHost:  fmt.Sprintf("%s:%s", cfg.Server.Host.Service, ctx.FunctionHttpPort()),
		log:          log.With(log.Any("function", "api")),
	}
	api.endpoints = append(api.endpoints, api.proxyEndpoints()...)

	handler := api.useRouter()
	cfg.Server.ServerConfig.Address = ":" + ctx.FunctionHttpPort()
	cfg.Server.ServerConfig.Certificate = cert
	api.svr = baetylhttp.NewServer(cfg.Server.ServerConfig, handler)
	api.svr.Start()
	return api, nil
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
		case a.functionHost:
			return a.onFunctionMessage(c)
		case a.serviceHost:
			return a.onServiceMessage(c)
		}
	}
	respondError(c, 404, "ERR_NO_ROUTE", "no route")
	return nil
}

func (a *API) onServiceMessage(c *routing.Context) error {
	uri := c.Request.URI()
	serviceName := c.Param("service")
	uri.SetHost(fmt.Sprintf("%s.%s", serviceName, edgeNamespace))
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

	address := fmt.Sprintf("%s.%s:%d", serviceName, edgeNamespace, a.cfg.Client.Grpc.Port)
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
