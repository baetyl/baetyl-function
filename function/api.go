package function

import (
	"context"
	"net/http"
	"strings"

	context2 "github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	baetyl "github.com/baetyl/baetyl-go/v2/faas"
	baetylhttp "github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/docker/distribution/uuid"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/baetyl/baetyl-function/v2/resolve"
)

type API struct {
	cfg       *Config
	svr       *baetylhttp.Server
	manager   Manager
	endpoints []Endpoint
	resolver  resolve.Resolver
	log       *log.Logger
}

type Endpoint struct {
	Methods []string
	Route   string
	Handler func(c *routing.Context) error
}

func NewAPI(cfg Config, ctx context2.Context, resolver resolve.Resolver) (*API, error) {
	cert := ctx.SystemConfig().Certificate
	m, err := NewManager(cert)
	if err != nil {
		return nil, err
	}

	api := &API{
		cfg:      &cfg,
		manager:  m,
		resolver: resolver,
		log:      log.With(log.Any("function", "api")),
	}
	api.endpoints = append(api.endpoints, api.proxyEndpoints()...)

	handler := api.useRouter()
	cfg.Server.Address = ":" + ctx.FunctionHttpPort()
	cfg.Server.Certificate = cert
	api.svr = baetylhttp.NewServer(cfg.Server, handler)
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
	if a.resolver != nil {
		a.resolver.Close()
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
			Handler: a.onFunctionMessage,
		},
		{
			Methods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
			Route:   "/<service>/<function>",
			Handler: a.onFunctionMessage,
		},
	}
}

func (a *API) onFunctionMessage(c *routing.Context) error {
	serviceName := c.Param("service")
	functionName := c.Param("function")
	body := c.PostBody()

	a.log.Info("proxy received a request", log.Any("service", serviceName), log.Any("function", functionName))

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

	address, err := a.resolver.Resolve(serviceName)
	if err != nil {
		a.log.Debug("resolve service's address failed", log.Error(err))
		respondError(c, 404, "ERR_ADDRESS_RESOLVE", err.Error())
		return nil
	}

	conn, err := a.manager.GetGRPCConnection(address, false)
	if err != nil {
		a.log.Debug("get grpc conn failed", log.Error(err))
		respondError(c, 500, "ERR_GET_GRPC_CONN", err.Error())
		return nil
	}

	for i := 0; i < a.cfg.Client.Grpc.Retries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Client.Grpc.Timeout)
		defer cancel()

		client := baetyl.NewFunctionClient(conn)
		resp, err := client.Call(ctx, &message)
		if err == nil {
			a.log.Debug("call function successfully", log.Any("service", serviceName), log.Any("function", functionName))
			respond(c, http.StatusOK, resp.Payload)
			return nil
		}

		code := status.Code(err)
		if code == codes.Unavailable || code == codes.Unauthenticated {
			a.log.Debug("function service is unavailable or unauthenticated with retry", log.Any("retry", i+1), log.Error(err))
			address, err = a.resolver.Resolve(serviceName)
			if err != nil {
				a.log.Debug("resolve service's address failed with retry", log.Any("retry", i+1), log.Error(err))
				respondError(c, 404, "ERR_ADDRESS_RESOLVE", err.Error())
				return nil
			}

			conn, err = a.manager.GetGRPCConnection(address, false)
			if err != nil {
				a.log.Debug("get grpc conn failed with retry", log.Any("retry", i+1), log.Error(err))
				respondError(c, 500, "ERR_GET_GRPC_CONN", err.Error())
				return nil
			}
			continue
		}

		a.log.Debug("call function failed", log.Any("service", serviceName), log.Any("function", functionName), log.Error(err))
		respondError(c, 500, "ERR_FUNCTION_CALL", err.Error())
		return nil
	}

	err = errors.Errorf("failed to invoke target %s after %v retries", address, a.cfg.Client.Grpc.Retries)
	a.log.Debug("call function failed", log.Any("service", serviceName), log.Any("function", functionName), log.Error(err))
	respondError(c, 500, "ERR_FUNCTION_CALL", err.Error())
	return nil
}
