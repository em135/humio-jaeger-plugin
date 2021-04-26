package plugin

import (
	"github.com/hashicorp/go-hclog"
	"net/http"
)

type HumioPlugin struct {
	spanReader       *spanReader
	spanWriter       *spanWriter
	dependencyReader *dependencyReader

	Logger hclog.Logger
	Client *http.Client
}

func NewHumioPlugin(logger hclog.Logger, token string) *HumioPlugin {
	client := http.DefaultClient
	rt := NewAddHeader(client.Transport, token)
	client.Transport = rt
	return &HumioPlugin{
		Logger: logger,
		Client: client,
	}
}

type AddHeader struct {
	rt    http.RoundTripper
	token string
}

func NewAddHeader(rt http.RoundTripper, token string) AddHeader {
	if rt == nil {
		rt = http.DefaultTransport
	}
	return AddHeader{rt, token}
}

func (ah AddHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer " + ah.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return ah.rt.RoundTrip(req)
}
