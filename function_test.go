package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
	"testing"
	"time"

	baetyl "github.com/baetyl/baetyl-go/faas"
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
		confFile     string
		address      string
		runFile      string
	}{
		{
			name:         "test python3 runtime",
			_exec:        "python3",
			functionName: "python3-sayhi",
			codePath:     path.Join([]string{"testdata", "python3", "code"}...),
			confFile:     path.Join([]string{"testdata", "python3", "config", "service.yml"}...),
			address:      "127.0.0.1:51200",
			runFile:      path.Join([]string{"python36", "runtime.py"}...),
		},
		{
			name:         "test node10 runtime",
			_exec:        "node",
			functionName: "node10-sayhi",
			codePath:     path.Join([]string{"testdata", "node10", "code"}...),
			confFile:     path.Join([]string{"testdata", "node10", "config", "service.yml"}...),
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
			env = append(env, fmt.Sprintf("%s=%s", "SERVICE_CONF", tt.confFile))
			env = append(env, fmt.Sprintf("%s=%s", "SERVICE_ADDRESS", tt.address))

			p, err := os.StartProcess(
				_exec,
				[]string{tt._exec, tt.runFile},
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

			// round 1: test json payload
			msgID := uint64(1234)
			msgTimestamp := string(time.Now().Unix())
			invokeId := uuid.Generate().String()
			payload := []byte(`{"name":"baetyl"}`)
			msg := &baetyl.Message{
				ID: msgID,
				Metadata: map[string]string{
					"functionName":     "index",
					"messageQOS":       "1",
					"messageTopic":     "topic-test",
					"serviceName":      "test-functionName",
					"invokeid":         invokeId,
					"messageTimestamp": msgTimestamp,
				},
				Payload: payload,
			}

			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel1()
			resp1, err1 := cli.Call(ctx1, msg)
			assert.NoError(t, err1)
			assert.NotEmpty(t, resp1)

			dataArr := map[string]interface{}{}
			err = json.Unmarshal(resp1.Payload, &dataArr)
			assert.NoError(t, err)
			assert.Equal(t, len(dataArr), 7)
			assert.Equal(t, dataArr["messageQOS"], "1")
			assert.Equal(t, dataArr["messageTopic"], "topic-test")
			assert.Equal(t, dataArr["functionName"], "index")
			assert.Equal(t, dataArr["invokeid"], invokeId)
			assert.Equal(t, dataArr["messageTimestamp"], msgTimestamp)
			assert.Equal(t, dataArr["name"], "baetyl")
			assert.Equal(t, dataArr["Say"], "Hello Baetyl")

			// round 2: test binary payload
			msgID2 := uint64(1234)
			msgTimestamp2 := strconv.FormatInt(time.Now().Unix(), 10)
			invokeId2 := uuid.Generate().String()
			payload2 := []byte("Baetyl Project")
			msg2 := &baetyl.Message{
				ID: msgID2,
				Metadata: map[string]string{
					"functionName":     "index",
					"messageQOS":       "1",
					"messageTopic":     "topic-test",
					"serviceName":      "test-functionName",
					"invokeid":         invokeId2,
					"messageTimestamp": msgTimestamp2,
				},
				Payload: payload2,
			}

			ctx2, cancel2 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel2()
			resp2, err2 := cli.Call(ctx2, msg2)
			assert.NoError(t, err2)
			assert.NotEmpty(t, resp2)

			dataArr2 := map[string]interface{}{}
			err2 = json.Unmarshal(resp2.Payload, &dataArr2)
			assert.NoError(t, err2)
			assert.Equal(t, len(dataArr2), 7)
			assert.Equal(t, dataArr2["messageQOS"], "1")
			assert.Equal(t, dataArr2["messageTopic"], "topic-test")
			assert.Equal(t, dataArr2["functionName"], "index")
			assert.Equal(t, dataArr2["invokeid"], invokeId2)
			assert.Equal(t, dataArr2["messageTimestamp"], msgTimestamp2)
			assert.Equal(t, dataArr2["bytes"], string(payload2))
			assert.Equal(t, dataArr2["Say"], "Hello Baetyl")

			// round 3: test empty payload
			msgID3 := uint64(1234)
			msgTimestamp3 := strconv.FormatInt(time.Now().Unix(), 10)
			invokeId3 := uuid.Generate().String()
			payload3 := []byte("")
			msg3 := &baetyl.Message{
				ID: msgID3,
				Metadata: map[string]string{
					"functionName":     "index",
					"messageQOS":       "1",
					"messageTopic":     "topic-test",
					"serviceName":      "test-functionName",
					"invokeid":         invokeId3,
					"messageTimestamp": msgTimestamp3,
				},
				Payload: payload3,
			}

			ctx3, cancel3 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel3()
			resp3, err3 := cli.Call(ctx3, msg3)
			assert.NoError(t, err3)
			assert.NotEmpty(t, resp3)
			assert.NotEmpty(t, resp3.Payload)

			// round 4: test error
			msgID4 := uint64(1234)
			msgTimestamp4 := strconv.FormatInt(time.Now().Unix(), 10)
			invokeId4 := uuid.Generate().String()
			payload4 := []byte(`{"err":"Baetyl"}`)
			msg4 := &baetyl.Message{
				ID: msgID4,
				Metadata: map[string]string{
					"method":           "index",
					"messageQOS":       "1",
					"messageTopic":     "topic-test",
					"functionName":     "index",
					"invokeid":         invokeId4,
					"messageTimestamp": msgTimestamp4,
				},
				Payload: payload4,
			}

			ctx4, cancel4 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel4()
			_, err4 := cli.Call(ctx4, msg4)
			assert.Error(t, err4)

			// round 5: test function doesn't exist
			msgID5 := uint64(1234)
			msgTimestamp5 := strconv.FormatInt(time.Now().Unix(), 10)
			invokeId5 := uuid.Generate().String()
			payload5 := []byte(`{"err":"Baetyl"}`)
			msg5 := &baetyl.Message{
				ID: msgID5,
				Metadata: map[string]string{
					"functionName":     "unknown",
					"messageQOS":       "1",
					"messageTopic":     "topic-test",
					"serviceName":      "test-functionName",
					"invokeid":         invokeId5,
					"messageTimestamp": msgTimestamp5,
				},
				Payload: payload5,
			}

			ctx5, cancel5 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel5()
			_, err5 := cli.Call(ctx5, msg5)
			assert.Error(t, err5)

			// round 5: test method is null
			msgID6 := uint64(1234)
			msgTimestamp6 := strconv.FormatInt(time.Now().Unix(), 10)
			invokeId6 := uuid.Generate().String()
			payload6 := []byte("Hello Baetyl")
			msg6 := &baetyl.Message{
				ID: msgID6,
				Metadata: map[string]string{
					"functionName":     "",
					"messageQOS":       "1",
					"messageTopic":     "topic-test",
					"serviceName":      "test-functionName",
					"invokeid":         invokeId6,
					"messageTimestamp": msgTimestamp6,
				},
				Payload: payload6,
			}

			ctx6, cancel6 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel6()
			_, err6 := cli.Call(ctx6, msg6)
			assert.NoError(t, err6)

			msgID7 := uint64(1234)
			msgTimestamp7 := string(time.Now().Unix())
			invokeId7 := uuid.Generate().String()
			payload7 := []byte(`{"name":"baetyl"}`)
			msg7 := &baetyl.Message{
				ID: msgID7,
				Metadata: map[string]string{
					"functionName":     "project",
					"messageQOS":       "1",
					"messageTopic":     "topic-test",
					"servicenName":     "test-functionName",
					"invokeid":         invokeId7,
					"messageTimestamp": msgTimestamp7,
				},
				Payload: payload7,
			}

			ctx7, cancel7 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel7()
			resp7, err7 := cli.Call(ctx7, msg7)
			assert.NoError(t, err7)
			assert.NotEmpty(t, resp7)

			dataArr7 := map[string]interface{}{}
			err7 = json.Unmarshal(resp7.Payload, &dataArr7)
			assert.NoError(t, err7)
			assert.Equal(t, len(dataArr7), 7)
			assert.Equal(t, dataArr7["messageQOS"], "1")
			assert.Equal(t, dataArr7["messageTopic"], "topic-test")
			assert.Equal(t, dataArr7["functionName"], "project")
			assert.Equal(t, dataArr7["invokeid"], invokeId7)
			assert.Equal(t, dataArr7["messageTimestamp"], msgTimestamp7)
			assert.Equal(t, dataArr7["name"], "baetyl")
			assert.Equal(t, dataArr7["Say"], "Hello Python36")

			err = p.Signal(syscall.SIGTERM)
			assert.NoError(t, err)
			p.Wait()
		})
	}
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
