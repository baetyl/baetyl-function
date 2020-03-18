package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/baetyl/baetyl-function/common"
	baetyl "github.com/baetyl/baetyl-function/proto"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func Test_FunctionInstance(t *testing.T) {
	tests := []struct {
		name         string
		_exec        string
		functionName string
		codePath     string
		address      string
		runFile      string
	}{
		{
			name:         "test python3 runtime",
			_exec:        "python3",
			functionName: "python3-sayhi",
			codePath:     path.Join([]string{"testdata", "python3"}...),
			address:      "127.0.0.1:50040",
			runFile:      path.Join([]string{"python36", "runtime.py"}...),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_exec, err := exec.LookPath(tt._exec)
			if err != nil {
				t.Skip("need " + tt._exec)
			}

			env := os.Environ()
			env = append(env, fmt.Sprintf("%s=%s", "SERVICE_NAME", tt.functionName))
			env = append(env, fmt.Sprintf("%s=%s", "SERVICE_CODE", tt.codePath))
			env = append(env, fmt.Sprintf("%s=%s", "SERVICE_ADDRESS", tt.address))

			p, err := os.StartProcess(
				_exec,
				[]string{"python3", tt.runFile},
				&os.ProcAttr{
					Env: env,
					Files: []*os.File{
						os.Stdin,
						os.Stdout,
						os.Stderr,
					},
				},
			)
			assert.NoError(t, err)

			cli, err := newMockFcClient(tt.address)
			assert.NoError(t, err)
			defer cli.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			messages := generateTestCase()

			resp0, err := cli.Call(ctx, &messages[0])
			assert.NoError(t, err)
			assert.NotEmpty(t, resp0)
			code0, err0 := strconv.Atoi(resp0.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err0)
			assert.Equal(t, code0, http.StatusOK)

			resp1, err1 := cli.Call(ctx, &messages[1])
			assert.NoError(t, err1)
			assert.NotEmpty(t, resp1)
			code1, err1 := strconv.Atoi(resp1.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err1)
			assert.Equal(t, code1, http.StatusBadGateway)

			resp2, err2 := cli.Call(ctx, &messages[2])
			assert.NoError(t, err2)
			assert.NotEmpty(t, resp2)
			code2, err2 := strconv.Atoi(resp2.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err2)
			assert.Equal(t, code2, http.StatusBadGateway)

			resp3, err3 := cli.Call(ctx, &messages[3])
			assert.NoError(t, err3)
			assert.NotEmpty(t, resp3)
			code3, err3 := strconv.Atoi(resp3.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err3)
			assert.Equal(t, code3, http.StatusBadGateway)

			resp4, err4 := cli.Call(ctx, &messages[4])
			assert.NoError(t, err4)
			assert.NotEmpty(t, resp4)
			code4, err4 := strconv.Atoi(resp4.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err4)
			assert.Equal(t, code4, http.StatusBadGateway)

			resp5, err5 := cli.Call(ctx, &messages[5])
			assert.NoError(t, err5)
			assert.NotEmpty(t, resp5)
			code5, err5 := strconv.Atoi(resp5.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err5)
			assert.Equal(t, code5, http.StatusBadGateway)

			resp6, err6 := cli.Call(ctx, &messages[6])
			assert.NoError(t, err6)
			assert.NotEmpty(t, resp6)
			code6, err6 := strconv.Atoi(resp6.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err6)
			assert.Equal(t, code6, http.StatusOK)
			headers := getHeader(resp6.Metadata["headers"])
			assert.Equal(t, headers["Content-Type"], "text/plain")

			resp7, err7 := cli.Call(ctx, &messages[7])
			assert.NoError(t, err7)
			assert.NotEmpty(t, resp7)
			code7, err7 := strconv.Atoi(resp7.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err7)
			assert.Equal(t, code7, http.StatusOK)

			resp8, err8 := cli.Call(ctx, &messages[8])
			assert.NoError(t, err8)
			assert.NotEmpty(t, resp8)
			code8, err8 := strconv.Atoi(resp8.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err8)
			assert.Equal(t, code8, http.StatusOK)

			resp9, err9 := cli.Call(ctx, &messages[9])
			assert.NoError(t, err9)
			assert.NotEmpty(t, resp9)
			code9, err9 := strconv.Atoi(resp9.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err9)
			assert.Equal(t, code9, http.StatusOK)

			resp10, err10 := cli.Call(ctx, &messages[10])
			assert.NoError(t, err10)
			assert.NotEmpty(t, resp10)
			code10, err10 := strconv.Atoi(resp8.Metadata[common.HTTPStatusCode])
			assert.NoError(t, err10)
			assert.Equal(t, code10, http.StatusOK)
			headers = getHeader(resp10.Metadata["headers"])
			assert.Equal(t, headers["Content-Type"], "application/json")

			err = p.Signal(syscall.SIGTERM)
			assert.NoError(t, err)
			p.Wait()
		})
	}
}

func generateTestCase() []baetyl.Message {
	metedata0 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata0["headers"] = generateHeader()
	message0 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "index",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata0,
	}

	metedata1 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata1["headers"] = generateHeader()
	message1 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "create1",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata1,
	}

	metedata2 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata2["headers"] = generateHeader()
	message2 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "create2",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata2,
	}

	metedata3 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata3["headers"] = generateHeader()
	message3 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "create3",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata3,
	}

	metedata4 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata4["headers"] = generateHeader()
	message4 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "create4",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata4,
	}

	metedata5 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata5["headers"] = generateHeader()
	message5 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "create5",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata5,
	}

	metedata6 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata6["headers"] = generateHeader()
	message6 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "create6",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata6,
	}

	metedata7 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata7["headers"] = generateHeader()
	message7 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "create7",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata7,
	}

	metedata8 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata8["headers"] = generateHeader()
	message8 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "create8",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata8,
	}

	metedata9 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata9["headers"] = generateHeader()
	message9 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "create9",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata9,
	}

	metedata10 := map[string]string{
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
		"invokeId":              uuid.Generate().String(),
	}
	metedata10["headers"] = generateHeader()
	message10 := baetyl.Message{
		Name:     "baetyl-function",
		Method:   "create10",
		Type:     "HTTP",
		Payload:  []byte("baetyl test"),
		Metadata: metedata9,
	}

	return []baetyl.Message{message0, message1, message2, message3,
		message4, message5, message6, message7, message8, message9,
		message10}
}

type fcClient struct {
	conn *grpc.ClientConn
	baetyl.FunctionClient
}

func newMockFcClient(address string) (*fcClient, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, err
	}
	client := baetyl.NewFunctionClient(conn)
	return &fcClient{
		conn:           conn,
		FunctionClient: client,
	}, nil
}

func (fc *fcClient) Close() {
	if fc.conn != nil {
		fc.conn.Close()
	}
}

func generateHeader() string {
	m := map[string]string{
		"Content-Type":     "application/json",
		"Content-Encoding": "gzip",
		"Connection":       "keep-alive",
	}

	var headers []string
	for key, value := range m {
		headers = append(headers, fmt.Sprintf("%s%s%s", key, common.HeaderEquals, value))
	}
	return strings.Join(headers, common.HeaderDelim)
}

func getHeader(val string) map[string]string {
	if val == "" {
		return nil
	}
	headers := map[string]string{}
	items := strings.Split(val, common.HeaderDelim)
	for _, h := range items {
		kv := strings.Split(h, common.HeaderEquals)
		headers[kv[0]] = kv[1]
	}
	return headers
}
