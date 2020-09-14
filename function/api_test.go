package function

import (
	"bytes"
	context2 "context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	baetyl "github.com/baetyl/baetyl-go/v2/faas"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	"github.com/baetyl/baetyl-go/v2/native"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v2"

	"github.com/baetyl/baetyl-function/v2/resolve"
)

func TestServerNativeNormal(t *testing.T) {
	logCfg := log.Config{
		Level:    "debug",
		Encoding: "json",
	}

	_, err := log.Init(logCfg)
	assert.NoError(t, err)

	cmd, err := os.Getwd()
	assert.NoError(t, err)

	certPath := path.Join(cmd, "temp")
	initCert(t, certPath)
	defer os.RemoveAll(path.Join(cmd, "temp"))

	serverCert := utils.Certificate{
		CA:   path.Join(certPath, "ca.pem"),
		Key:  path.Join(certPath, "key.pem"),
		Cert: path.Join(certPath, "crt.pem"),
	}

	var cfg Config
	err = utils.UnmarshalYAML(nil, &cfg)
	assert.NoError(t, err)
	cfg.Server.Address = ":50050"
	cfg.Server.ReadTimeout = 2 * time.Second

	ctx := &mockContext{
		cert: serverCert,
		ns:   "test",
		port: "50050",
	}

	ports, err := getFreePorts(3)

	mapping := native.ServiceMapping{
		Services: map[string]native.ServiceMappingInfo{
			"serviceA": {
				Ports: native.PortsInfo{
					Items: ports[:1],
				},
			},
			"serviceB": {
				Ports: native.PortsInfo{
					Items: ports[1:2],
				},
			},
		},
	}

	data, err := yaml.Marshal(mapping)
	assert.NoError(t, err)

	portMappingFile := path.Join(cmd, native.ServiceMappingFile)
	err = os.MkdirAll(path.Dir(portMappingFile), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(portMappingFile, data, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(path.Join(cmd, "var"))

	s0 := mockGrpc(t, ports[0], serverCert)
	defer s0.GracefulStop()

	s1 := mockGrpc(t, ports[1], serverCert)
	defer s1.GracefulStop()

	s2 := mockGrpc(t, ports[2], serverCert)
	defer s2.GracefulStop()

	resolver, err := resolve.NewNativeResolver(ctx)
	assert.NotNil(t, resolver)
	assert.NoError(t, err)

	api, err := NewAPI(cfg, ctx, resolver)
	assert.NoError(t, err)
	assert.NotEmpty(t, api)
	defer api.Close()

	cert := utils.Certificate{
		CA:   path.Join(certPath, "ca.pem"),
		Key:  path.Join(certPath, "clientKey.pem"),
		Cert: path.Join(certPath, "clientCrt.pem"),
	}

	tlsConfig, err := utils.NewTLSConfigClient(cert)
	assert.NoError(t, err)

	ops := http.NewClientOptions()
	ops.TLSConfig = tlsConfig
	client := http.NewClient(ops)

	resp, err := client.PostURL("https://localhost:50050", bytes.NewBuffer([]byte("payload")))
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 404)

	resp, err = client.PostURL("https://localhost:50050/serviceA", bytes.NewBuffer([]byte("payload")))
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	respData, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(respData), fmt.Sprintf("{\"port\":%d}", ports[0]))

	resp, err = client.PostURL("https://localhost:50050/serviceB", bytes.NewBuffer([]byte("payload")))
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	respData, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(respData), fmt.Sprintf("{\"port\":%d}", ports[1]))

	mapping.Services["serviceA"] = native.ServiceMappingInfo{
		Ports: native.PortsInfo{
			Items: ports[2:3],
		},
	}

	data, err = yaml.Marshal(mapping)
	assert.NoError(t, err)

	err = ioutil.WriteFile(portMappingFile, data, 0755)
	assert.NoError(t, err)

	time.Sleep(time.Microsecond * 500)

	resp, err = client.PostURL("https://localhost:50050/serviceA", bytes.NewBuffer([]byte("payload")))
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	respData, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(respData), fmt.Sprintf("{\"port\":%d}", ports[2]))

	fmt.Println("-----> end <-----")
}

type mockGrpcServer struct {
	port int
}

func (m *mockGrpcServer) Call(ctx context2.Context, msg *baetyl.Message) (*baetyl.Message, error) {
	body := string(msg.Payload)
	if body == "error" {
		return nil, errors.New("err")
	} else {
		o := map[string]int{
			"port": m.port,
		}
		s, _ := json.Marshal(o)
		msg.Payload = s
	}
	return msg, nil
}

func mockGrpc(t *testing.T, port int, cert utils.Certificate) *grpc.Server {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	assert.NoError(t, err)
	tlsCfg, err := utils.NewTLSConfigServer(cert)
	assert.NoError(t, err)
	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsCfg)))
	grpcServer := &mockGrpcServer{port: port}
	baetyl.RegisterFunctionServer(s, grpcServer)
	go func() {
		fmt.Printf("-----> grpc server is running at: %d with tls <-----\n", port)
		err = s.Serve(lis)
		assert.NoError(t, err)
	}()
	return s
}

func getFreePorts(n int) ([]int, error) {
	ports := make([]int, 0)
	for i := 0; i < n; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return nil, err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}
		l.Close()
		ports = append(ports, l.Addr().(*net.TCPAddr).Port)
	}
	return ports, nil
}

type mockContext struct {
	cert utils.Certificate
	ns   string
	port string
}

func (m *mockContext) NodeName() string {
	return ""
}

func (m *mockContext) AppName() string {
	return ""
}

func (m *mockContext) AppVersion() string {
	return ""
}

func (m *mockContext) ServiceName() string {
	return ""
}

func (m *mockContext) ConfFile() string {
	return ""
}

func (m *mockContext) RunMode() string {
	return ""
}

func (m *mockContext) BrokerHost() string {
	return ""
}

func (m *mockContext) BrokerPort() string {
	return ""
}

func (m *mockContext) FunctionHost() string {
	return ""
}

func (m *mockContext) FunctionHttpPort() string {
	return m.port
}

func (m *mockContext) EdgeNamespace() string {
	return ""
}

func (m *mockContext) EdgeSystemNamespace() string {
	return ""
}

func (m *mockContext) SystemConfig() *context.SystemConfig {
	return &context.SystemConfig{
		Certificate: m.cert,
	}
}

func (m *mockContext) Log() *log.Logger {
	return nil
}

func (m *mockContext) Wait() {
	return
}

func (m *mockContext) WaitChan() <-chan os.Signal {
	return nil
}

func (m *mockContext) Load(key interface{}) (value interface{}, ok bool) {
	return nil, false
}

func (m *mockContext) Store(key, value interface{}) {
	return
}

func (m *mockContext) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	return nil, false
}

func (m *mockContext) Delete(key interface{}) {
	return
}

func (m *mockContext) CheckSystemCert() error {
	return nil
}

func (m *mockContext) LoadCustomConfig(cfg interface{}, files ...string) error {
	return nil
}

func (m *mockContext) NewFunctionHttpClient() (*http.Client, error) {
	return nil, nil
}

func (m *mockContext) NewSystemBrokerClientConfig() (mqtt.ClientConfig, error) {
	return mqtt.ClientConfig{}, nil
}

func (m *mockContext) NewBrokerClient(mqtt.ClientConfig) (*mqtt.Client, error) {
	return nil, nil
}
