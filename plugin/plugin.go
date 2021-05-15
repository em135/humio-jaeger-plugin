package plugin

import (
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/hashicorp/go-hclog"
)

type HumioPlugin struct {
	spanReader       *spanReader
	spanWriter       *spanWriter
	dependencyReader *dependencyReader

	logger hclog.Logger
	client *http.Client
	url    string
}

func NewHumioPlugin(logger hclog.Logger, token string) *HumioPlugin {
	client := http.DefaultClient
	rt := NewAddHeader(client.Transport, token)
	client.Transport = rt

	endpoint, err := url.Parse(os.Getenv("HUMIO_ENDPOINT"))
	if err != nil {
		logger.Error(err.Error())
	}

	repo := os.Getenv("HUMIO_REPOSITORY")
	endpoint.Path = path.Join(endpoint.Path, "api/v1/repositories/", repo, "query")

	return &HumioPlugin{
		logger: logger,
		client: client,
		url:    endpoint.String(),
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
	req.Header.Set("Authorization", "Bearer "+ah.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return ah.rt.RoundTrip(req)
}
