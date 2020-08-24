package function

import (
	"os"
	"path"
	"testing"

	"github.com/baetyl/baetyl-go/v2/http"
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

	api, err := NewAPI(cfg, "test-ns", cert)
	assert.NoError(t, err)
	assert.NotEmpty(t, api)
	defer api.Close()

	api2, err := NewAPI(cfg, "test-ns", cert)
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

	api, err := NewAPI(cfg, "test-ns", serverCert)
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

func mockConfig1() Config {
	var cfg Config
	utils.UnmarshalYAML(nil, &cfg)
	cfg.Server.Address = ":50050"
	return cfg
}
