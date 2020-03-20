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

	baetyl "github.com/baetyl/baetyl-go/faas"
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
			address:      "127.0.0.1:51200",
			runFile:      path.Join([]string{"python36", "runtime.py"}...),
		},
		{
			name:         "test node10 runtime",
			_exec:        "node",
			functionName: "node10-sayhi",
			codePath:     path.Join([]string{"testdata", "node10"}...),
			address:      "127.0.0.1:51201",
			runFile:      path.Join([]string{"node10", "runtime.js"}...),
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

			messages := generateTestCase()

			ctx0, cancel0 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel0()
			resp0, err := cli.Call(ctx0, &messages[0])
			assert.NoError(t, err)
			assert.NotEmpty(t, resp0)
			code0, err0 := strconv.Atoi(resp0.Metadata[HTTPStatusCode])
			assert.NoError(t, err0)
			assert.Equal(t, code0, http.StatusOK)

			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel1()
			resp1, err1 := cli.Call(ctx1, &messages[1])
			assert.NoError(t, err1)
			assert.NotEmpty(t, resp1)
			code1, err1 := strconv.Atoi(resp1.Metadata[HTTPStatusCode])
			assert.NoError(t, err1)
			assert.Equal(t, code1, http.StatusBadGateway)

			ctx2, cancel2 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel2()
			resp2, err2 := cli.Call(ctx2, &messages[2])
			assert.NoError(t, err2)
			assert.NotEmpty(t, resp2)
			code2, err2 := strconv.Atoi(resp2.Metadata[HTTPStatusCode])
			assert.NoError(t, err2)
			assert.Equal(t, code2, http.StatusBadGateway)

			ctx3, cancel3 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel3()
			resp3, err3 := cli.Call(ctx3, &messages[3])
			assert.NoError(t, err3)
			assert.NotEmpty(t, resp3)
			code3, err3 := strconv.Atoi(resp3.Metadata[HTTPStatusCode])
			assert.NoError(t, err3)
			assert.Equal(t, code3, http.StatusBadGateway)

			ctx4, cancel4 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel4()
			resp4, err4 := cli.Call(ctx4, &messages[4])
			assert.NoError(t, err4)
			assert.NotEmpty(t, resp4)
			code4, err4 := strconv.Atoi(resp4.Metadata[HTTPStatusCode])
			assert.NoError(t, err4)
			assert.Equal(t, code4, http.StatusBadGateway)

			ctx5, cancel5 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel5()
			resp5, err5 := cli.Call(ctx5, &messages[5])
			assert.NoError(t, err5)
			assert.NotEmpty(t, resp5)
			code5, err5 := strconv.Atoi(resp5.Metadata[HTTPStatusCode])
			assert.NoError(t, err5)
			assert.Equal(t, code5, http.StatusBadGateway)

			ctx6, cancel6 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel6()
			resp6, err6 := cli.Call(ctx6, &messages[6])
			assert.NoError(t, err6)
			assert.NotEmpty(t, resp6)
			code6, err6 := strconv.Atoi(resp6.Metadata[HTTPStatusCode])
			assert.NoError(t, err6)
			assert.Equal(t, code6, http.StatusOK)
			headers := getHeader(resp6.Metadata["headers"])
			assert.Equal(t, headers["Content-Type"], "text/plain")

			ctx7, cancel7 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel7()
			resp7, err7 := cli.Call(ctx7, &messages[7])
			assert.NoError(t, err7)
			assert.NotEmpty(t, resp7)
			code7, err7 := strconv.Atoi(resp7.Metadata[HTTPStatusCode])
			assert.NoError(t, err7)
			assert.Equal(t, code7, http.StatusOK)

			ctx8, cancel8 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel8()
			resp8, err8 := cli.Call(ctx8, &messages[8])
			assert.NoError(t, err8)
			assert.NotEmpty(t, resp8)
			code8, err8 := strconv.Atoi(resp8.Metadata[HTTPStatusCode])
			assert.NoError(t, err8)
			assert.Equal(t, code8, http.StatusOK)

			ctx9, cancel9 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel9()
			resp9, err9 := cli.Call(ctx9, &messages[9])
			assert.NoError(t, err9)
			assert.NotEmpty(t, resp9)
			code9, err9 := strconv.Atoi(resp9.Metadata[HTTPStatusCode])
			assert.NoError(t, err9)
			assert.Equal(t, code9, http.StatusOK)

			ctx10, cancel10 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel10()
			resp10, err10 := cli.Call(ctx10, &messages[10])
			assert.NoError(t, err10)
			assert.NotEmpty(t, resp10)
			code10, err10 := strconv.Atoi(resp10.Metadata[HTTPStatusCode])
			assert.NoError(t, err10)
			assert.Equal(t, code10, http.StatusOK)
			headers = getHeader(resp10.Metadata["headers"])
			assert.Equal(t, headers["Content-Type"], "application/json")

			ctx11, cancel11 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel11()
			resp11, err11 := cli.Call(ctx11, &messages[11])
			assert.NoError(t, err11)
			assert.NotEmpty(t, resp11)
			code11, err11 := strconv.Atoi(resp11.Metadata[HTTPStatusCode])
			assert.NoError(t, err11)
			assert.Equal(t, code11, http.StatusBadGateway)

			ctx12, cancel12 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel12()
			resp12, err12 := cli.Call(ctx12, &messages[12])
			assert.NoError(t, err12)
			assert.NotEmpty(t, resp12)
			code12, err12 := strconv.Atoi(resp12.Metadata[HTTPStatusCode])
			assert.NoError(t, err12)
			assert.Equal(t, code12, http.StatusOK)

			ctx13, cancel13 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel13()
			resp13, err13 := cli.Call(ctx13, &messages[13])
			assert.NoError(t, err13)
			assert.NotEmpty(t, resp13)
			code13, err13 := strconv.Atoi(resp13.Metadata[HTTPStatusCode])
			assert.NoError(t, err13)
			assert.Equal(t, code13, http.StatusOK)

			ctx14, cancel14 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel14()
			resp14, err14 := cli.Call(ctx14, &messages[14])
			assert.NoError(t, err14)
			assert.NotEmpty(t, resp14)
			code14, err14 := strconv.Atoi(resp14.Metadata[HTTPStatusCode])
			assert.NoError(t, err14)
			assert.Equal(t, code14, http.StatusInternalServerError)
			assert.Contains(t, string(resp14.Payload), "test custom error")

			ctx15, cancel15 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel15()
			resp15, err15 := cli.Call(ctx15, &messages[15])
			assert.NoError(t, err15)
			assert.NotEmpty(t, resp15)
			code15, err15 := strconv.Atoi(resp15.Metadata[HTTPStatusCode])
			assert.NoError(t, err15)
			assert.Equal(t, code15, http.StatusInternalServerError)
			assert.Contains(t, string(resp15.Payload), "test custom error")

			ctx16, cancel16 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel16()
			resp16, err16 := cli.Call(ctx16, &messages[16])
			assert.NoError(t, err16)
			assert.NotEmpty(t, resp16)
			code16, err16 := strconv.Atoi(resp16.Metadata[HTTPStatusCode])
			assert.NoError(t, err16)
			assert.Equal(t, code16, http.StatusInternalServerError)

			err = p.Signal(syscall.SIGTERM)
			assert.NoError(t, err)
			p.Wait()
		})
	}
}

func generateTestCase() []baetyl.Message {
	metedata0 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "index",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata0["headers"] = generateHeader()
	message0 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata0,
	}

	metedata1 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create1",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata1["headers"] = generateHeader()
	message1 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata1,
	}

	metedata2 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create2",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata2["headers"] = generateHeader()
	message2 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata2,
	}

	metedata3 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create3",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata3["headers"] = generateHeader()
	message3 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata3,
	}

	metedata4 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create4",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata4["headers"] = generateHeader()
	message4 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata4,
	}

	metedata5 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create5",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata5["headers"] = generateHeader()
	message5 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata5,
	}

	metedata6 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create6",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata6["headers"] = generateHeader()
	message6 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata6,
	}

	metedata7 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create7",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata7["headers"] = generateHeader()
	message7 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata7,
	}

	metedata8 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create8",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata8["headers"] = generateHeader()
	message8 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata8,
	}

	metedata9 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create9",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata9["headers"] = generateHeader()
	message9 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata9,
	}

	metedata10 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create10",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata10["headers"] = generateHeader()
	message10 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata10,
	}

	metedata11 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create11",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata11["headers"] = generateHeader()
	message11 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata11,
	}

	metedata12 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create12",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata12["headers"] = generateHeader()
	message12 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata12,
	}

	metedata13 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create13",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata13["headers"] = generateHeader()
	message13 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata13,
	}

	metedata14 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create14",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata14["headers"] = generateHeader()
	message14 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata14,
	}

	metedata15 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create15",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata15["headers"] = generateHeader()
	message15 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata15,
	}

	metedata16 := map[string]string{
		"type":                  "HTTP",
		"name":                  "baetyl-function",
		"method":                "create16",
		"path":                  "test",
		"httpMethod":            strings.ToUpper(http.MethodGet),
		"isBase64Encoded":       "false",
		"queryStringParameters": "name=baetyl&cxv=cxv",
	}
	metedata16["headers"] = generateHeader()
	message16 := baetyl.Message{
		ID:       12345,
		Payload:  []byte("baetyl test"),
		Metadata: metedata16,
	}

	return []baetyl.Message{message0, message1, message2, message3,
		message4, message5, message6, message7, message8, message9,
		message10, message11, message12, message13, message14, message15,
		message16}
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
		headers = append(headers, fmt.Sprintf("%s%s%s", key, HeaderEquals, value))
	}
	return strings.Join(headers, HeaderDelim)
}

func getHeader(val string) map[string]string {
	if val == "" {
		return nil
	}
	headers := map[string]string{}
	items := strings.Split(val, HeaderDelim)
	for _, h := range items {
		kv := strings.Split(h, HeaderEquals)
		headers[kv[0]] = kv[1]
	}
	return headers
}
