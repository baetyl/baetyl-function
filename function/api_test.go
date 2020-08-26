package function

import (
	"os"
	"path"
	"testing"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/stretchr/testify/assert"
)

func TestServer0(t *testing.T) {
	cmd, err := os.Getwd()
	assert.NoError(t, err)

	certPath := path.Join(cmd, "temp")
	initCert(t, certPath)
	defer os.RemoveAll(path.Join(cmd, "temp"))

	cert := utils.Certificate{
		CA:   path.Join(certPath, "ca.pem"),
		Key:  path.Join(certPath, "clientKey.pem"),
		Cert: path.Join(certPath, "clientCrt.pem"),
	}

	var cfg Config
	utils.UnmarshalYAML(nil, &cfg)
	cfg.Server.Address = ":50060"

	ctx := &mockContext{
		cert: cert,
		ns:   "test",
		port: "50060",
	}

	api, err := NewAPI(cfg, ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, api)
	defer api.Close()

	api2, err := NewAPI(cfg, ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, api2)
	defer api2.Close()
}

func TestServer1(t *testing.T) {
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
	utils.UnmarshalYAML(nil, &cfg)
	cfg.Server.Address = ":50050"

	ctx := &mockContext{
		cert: serverCert,
		ns:   "test",
		port: "50050",
	}

	api, err := NewAPI(cfg, ctx)
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

	resp, err := client.GetURL("https://localhost:50050")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 404)

	resp, err = client.GetURL("https://localhost:50050/test")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 404)

	resp, err = client.GetURL("https://localhost:50050/book/create/detail")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 500)
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
