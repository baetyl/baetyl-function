package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
	"testing"
	"time"

	baetyl "github.com/baetyl/baetyl-go/v2/faas"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Test_FunctionInstance(t *testing.T) {
	cmd, err := os.Getwd()
	assert.NoError(t, err)

	certPath := path.Join(cmd, "var/lib/baetyl/system/certs/")
	initCert(t, certPath)
	defer os.RemoveAll(path.Join(cmd, "var"))

	tests := []struct {
		name         string
		_exec        string
		functionName string
		codePath     string
		confFile     string
		port         string
		runFile      string
	}{
		{
			name:         "test python3 runtime",
			_exec:        "python3",
			functionName: "python3-sayhi",
			codePath:     path.Join([]string{cmd, "..", "testdata", "python3", "code"}...),
			confFile:     path.Join([]string{cmd, "..", "testdata", "python3", "config", "service.yml"}...),
			port:         "51200",
			runFile:      path.Join([]string{"..", "python36", "runtime.py"}...),
		},
		{
			name:         "test node10 runtime",
			_exec:        "node",
			functionName: "node10-sayhi",
			codePath:     path.Join([]string{cmd, "..", "testdata", "node10", "code"}...),
			confFile:     path.Join([]string{cmd, "..", "testdata", "node10", "config", "service.yml"}...),
			port:         "51201",
			runFile:      path.Join([]string{"..", "node10", "runtime.js"}...),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_exec, err := exec.LookPath(tt._exec)
			if err != nil {
				t.Skip("need " + tt._exec)
			}

			env := os.Environ()
			env = append(env, fmt.Sprintf("%s=%s", "BAETYL_SERVICE_NAME", tt.functionName))
			env = append(env, fmt.Sprintf("%s=%s", "BAETYL_CODE_PATH", tt.codePath))
			env = append(env, fmt.Sprintf("%s=%s", "BAETYL_CONF_FILE", tt.confFile))
			env = append(env, fmt.Sprintf("%s=%s", "BAETYL_SERVICE_DYNAMIC_PORT", tt.port))
			env = append(env, fmt.Sprintf("%s=%s", "BAETYL_RUN_MODE", "native"))

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

			cli, err := newMockFcClient("127.0.0.1:"+tt.port, utils.Certificate{
				CA:   path.Join(certPath, "ca.pem"),
				Key:  path.Join(certPath, "clientKey.pem"),
				Cert: path.Join(certPath, "clientCrt.pem"),
			})
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

			ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
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

			ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
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

			ctx3, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
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

			ctx4, cancel4 := context.WithTimeout(context.Background(), 5*time.Second)
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

			ctx5, cancel5 := context.WithTimeout(context.Background(), 5*time.Second)
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

			ctx6, cancel6 := context.WithTimeout(context.Background(), 5*time.Second)
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

			ctx7, cancel7 := context.WithTimeout(context.Background(), 5*time.Second)
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

func newMockFcClient(address string, cert utils.Certificate) (*fcClient, error) {
	tlsConfig, err := utils.NewTLSConfigClient(cert)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithBlock(),
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

const (
	ca = `-----BEGIN CERTIFICATE-----
MIICfzCCAiSgAwIBAgIIFizowlvYkxgwCgYIKoZIzj0EAwIwgaUxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMRkwFwYDVQQHExBIYWlkaWFuIERpc3RyaWN0
MRUwEwYDVQQJEwxCYWlkdSBDYW1wdXMxDzANBgNVBBETBjEwMDA5MzEeMBwGA1UE
ChMVTGludXggRm91bmRhdGlvbiBFZGdlMQ8wDQYDVQQLEwZCQUVUWUwxEDAOBgNV
BAMTB3Jvb3QuY2EwIBcNMjAwODIwMDcxODA5WhgPMjA3MDA4MDgwNzE4MDlaMIGl
MQswCQYDVQQGEwJDTjEQMA4GA1UECBMHQmVpamluZzEZMBcGA1UEBxMQSGFpZGlh
biBEaXN0cmljdDEVMBMGA1UECRMMQmFpZHUgQ2FtcHVzMQ8wDQYDVQQREwYxMDAw
OTMxHjAcBgNVBAoTFUxpbnV4IEZvdW5kYXRpb24gRWRnZTEPMA0GA1UECxMGQkFF
VFlMMRAwDgYDVQQDEwdyb290LmNhMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
3GSIw55wTQIaVWSD2fePbIts9pToj9OtiyG0/1zlvkht1Go2yCGc0xwaoR0YdW1H
Fi1jpzMfmvJhppQaz5F6F6M6MDgwDgYDVR0PAQH/BAQDAgGGMA8GA1UdEwEB/wQF
MAMBAf8wFQYDVR0RBA4wDIcEAAAAAIcEfwAAATAKBggqhkjOPQQDAgNJADBGAiEA
qaeTS1oKts1XiC6eWkuK0n6TH45yWJvC3/NU6PqpBSYCIQDIHGDb3OL+4OsUitvb
svDCT14MNf0cgIeg7gO+D0Xvqg==
-----END CERTIFICATE-----
`
	serverCrt = `-----BEGIN CERTIFICATE-----
MIICojCCAkigAwIBAgIIFizowlwTTkAwCgYIKoZIzj0EAwIwgaUxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMRkwFwYDVQQHExBIYWlkaWFuIERpc3RyaWN0
MRUwEwYDVQQJEwxCYWlkdSBDYW1wdXMxDzANBgNVBBETBjEwMDA5MzEeMBwGA1UE
ChMVTGludXggRm91bmRhdGlvbiBFZGdlMQ8wDQYDVQQLEwZCQUVUWUwxEDAOBgNV
BAMTB3Jvb3QuY2EwHhcNMjAwODIwMDcxODA5WhcNNDAwODE1MDcxODA5WjCBpDEL
MAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxGTAXBgNVBAcTEEhhaWRpYW4g
RGlzdHJpY3QxFTATBgNVBAkTDEJhaWR1IENhbXB1czEPMA0GA1UEERMGMTAwMDkz
MR4wHAYDVQQKExVMaW51eCBGb3VuZGF0aW9uIEVkZ2UxDzANBgNVBAsTBkJBRVRZ
TDEPMA0GA1UEAxMGc2VydmVyMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQXCQ
TGn4+frJYOumFk8gs8BIbgduEuiHonhYdJTFGIPLiOqPQoIvDmICod7W0oIzYYXw
TF4NfadliSryoXx9IaNhMF8wDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsG
AQUFBwMCBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMCAGA1UdEQQZMBeCCWxvY2Fs
aG9zdIcEAAAAAIcEfwAAATAKBggqhkjOPQQDAgNIADBFAiB5vz8+oob7SkN54uf7
RErbE4tWT5AHtgBgIs3A+TjnyQIhAPvnL8W1dq4qdkVr0eiH5He0xNHdsQc6eWxS
RcKyjhh1
-----END CERTIFICATE-----
`
	serverKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIPcuJ/l+/s8PXvAN5M6VNBZrKD4HDW6n6Y4rQCYinF5doAoGCCqGSM49
AwEHoUQDQgAEQXCQTGn4+frJYOumFk8gs8BIbgduEuiHonhYdJTFGIPLiOqPQoIv
DmICod7W0oIzYYXwTF4NfadliSryoXx9IQ==
-----END EC PRIVATE KEY-----
`
	clientCrt = `-----BEGIN CERTIFICATE-----
MIIClzCCAj2gAwIBAgIIFizowlwFTFAwCgYIKoZIzj0EAwIwgaUxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMRkwFwYDVQQHExBIYWlkaWFuIERpc3RyaWN0
MRUwEwYDVQQJEwxCYWlkdSBDYW1wdXMxDzANBgNVBBETBjEwMDA5MzEeMBwGA1UE
ChMVTGludXggRm91bmRhdGlvbiBFZGdlMQ8wDQYDVQQLEwZCQUVUWUwxEDAOBgNV
BAMTB3Jvb3QuY2EwHhcNMjAwODIwMDcxODA5WhcNNDAwODE1MDcxODA5WjCBpDEL
MAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxGTAXBgNVBAcTEEhhaWRpYW4g
RGlzdHJpY3QxFTATBgNVBAkTDEJhaWR1IENhbXB1czEPMA0GA1UEERMGMTAwMDkz
MR4wHAYDVQQKExVMaW51eCBGb3VuZGF0aW9uIEVkZ2UxDzANBgNVBAsTBkJBRVRZ
TDEPMA0GA1UEAxMGY2xpZW50MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAErp2K
LVIYfqeCzJlR/gteIZyN7i1/ckXuuXNO1i2GGu/bFdkoj1ST1ypj1FuY/WpdmwSQ
HBIVm42s1vf0Gnc7yKNWMFQwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsG
AQUFBwMCBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMBUGA1UdEQQOMAyHBAAAAACH
BH8AAAEwCgYIKoZIzj0EAwIDSAAwRQIgOVxan95heTLe3c20iUvPJmX1EPMvfg6J
5GeWeK2cA8QCIQCfO6xOoQj386u+7XD4K4srGdFj77f9tfWt/M6ryIscdA==
-----END CERTIFICATE-----
`
	clientKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIAwqz3MtAG2O5xaKy2RXWLxcOpKYKWilCs89L27+6it1oAoGCCqGSM49
AwEHoUQDQgAErp2KLVIYfqeCzJlR/gteIZyN7i1/ckXuuXNO1i2GGu/bFdkoj1ST
1ypj1FuY/WpdmwSQHBIVm42s1vf0Gnc7yA==
-----END EC PRIVATE KEY-----
`
)

func initCert(t *testing.T, certDir string) {
	err := os.MkdirAll(certDir, 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(path.Join(certDir, "ca.pem"), []byte(ca), os.ModePerm)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(certDir, "crt.pem"), []byte(serverCrt), os.ModePerm)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(certDir, "key.pem"), []byte(serverKey), os.ModePerm)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(certDir, "clientCrt.pem"), []byte(clientCrt), os.ModePerm)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(certDir, "clientKey.pem"), []byte(clientKey), os.ModePerm)
	assert.NoError(t, err)
}
