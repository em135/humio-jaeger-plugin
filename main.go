package main

import (
	"flag"
	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc/shared"
	"humio-jaeger-plugin/plugin"
	"os"
)

const (
	loggerName = "jaeger-humio"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Warn,
		Name:       loggerName,
		JSONFormat: true,
	})

	var configPath string
	flag.StringVar(&configPath, "config", "", "A path to the Humio plugin's configuration file")
	flag.Parse()

	token, tokenExists := os.LookupEnv("API_TOKEN")
	if !tokenExists {
		logger.Error("No API_TOKEN provided")
		os.Exit(0)
	}

	humioPlugin := plugin.NewHumioPlugin(logger, token)
	grpc.Serve(&shared.PluginServices{
		Store: humioPlugin,
	})
}
