package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/baetyl/baetyl-function/common"
	baetyl "github.com/baetyl/baetyl-function/proto"
	routing "github.com/qiangxue/fasthttp-routing"
	"google.golang.org/grpc"
)

type API struct {
	endpoints             []Endpoint
	connectionCreatorFn func(address string, recreateIfExists bool) (*grpc.ClientConn, error)
}

// Endpoint is a collection of route information for an Dapr API
type Endpoint struct {
	Methods []string
	Route   string
	Handler func(c *routing.Context) error
}

func NewAPI(m *Manager) *API {
	api := &API{
		connectionCreatorFn: m.GetGRPCConnection,
	}
	api.endpoints = append(api.endpoints, api.constructFunctionEndpoints()...)
}

func (a *API) Close() {
	// 需要实现成接口
}

func (a *API) constructFunctionEndpoints() []Endpoint {
	return []Endpoint{
		{
			Methods: []string{common.Get, common.Post, common.Delete, common.Put},
			Route:   "baetyl-function/<service>/*",
			Handler: a.onFunctionMessage,
		},
	}
}

func (a *API) onFunctionMessage(c *routing.Context) error {
	serviceName := c.Param("service")
	path := string(c.Path())
	// TODO: 只有一级
	method := path[strings.Index(path, serviceName + "/")+7:]
	body := c.PostBody()
	verb := strings.ToUpper(string(c.Method()))
	queryString := string(c.QueryArgs().QueryString())

	metedata := map[string]string{common.HTTPVerb: verb, common.QueryString: queryString}
	a.setHeaders(c, metedata)
	message := baetyl.MessageRequest{
		Name:                 serviceName,
		Method:               method,
		Type:                 "HTTP",
		Payload:              body,
		Metadata:             metedata,
	}

	address := "xxx"
	conn, err := a.connectionCreatorFn(address, false)
	if err != nil {
		msg := NewErrorResponse("ERR_FUNCTION_CALL", err.Error())
		respondWithError(c.RequestCtx, 500, msg)
	}

	// TODO: 超时时间要设置到配置文件中
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	client := baetyl.NewFunctionClient(conn)
	resp, err := client.Call(ctx, &message)
	if err != nil {
		msg := NewErrorResponse("ERR_FUNCTION_CALL", err.Error())
		respondWithError(c.RequestCtx, 500, msg)
	} else {
		statusCode := GetStatusCodeFromMetadata(resp.Metadata)
		a.setHeadersOnRequest(resp.Metadata, c)
		respond(c.RequestCtx, statusCode, resp.Payload)
	}
	return nil
}

func (a *API) setHeaders(c *routing.Context, metadata map[string]string) {
	var headers []string
	c.RequestCtx.Request.Header.VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)

		headers = append(headers, fmt.Sprintf("%s&__header_equals__&%s", k, v))
	})
	if len(headers) > 0 {
		metadata["headers"] = strings.Join(headers, "&__header_delim__&")
	}
}

type FunctionMessage struct {
	connectionCreatorFn func(address string, recreateIfExists bool) (*grpc.ClientConn, error) {
}

func (a *API) setHeadersOnRequest(metadata map[string]string, c *routing.Context) {
	if metadata == nil {
		return
	}
	if val, ok := metadata["headers"]; ok {
		headers := strings.Split(val, "&__header_delim__&")
		for _, h := range headers {
			kv := strings.Split(h, "&__header_equals__&")
			c.RequestCtx.Response.Header.Set(kv[0], kv[1])
		}
	}
}

// GetStatusCodeFromMetadata extracts the http status code from the metadata if it exists
func GetStatusCodeFromMetadata(metadata map[string]string) int {
	code := metadata[common.HTTPStatusCode]
	if code != "" {
		statusCode, err := strconv.Atoi(code)
		if err == nil {
			return statusCode
		}
	}

	return 200
}