#!/bin/bash
export GOOS=linux
export GOARCH=amd64
CGO_ENABLED=0 go build -v -ldflags '-extldflags "-static"'

docker stop jaeger
docker rm jaeger

docker build --no-cache -t  em135/humio-jaeger-plugin:latest .
docker run -it -d \
  --name jaeger \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 9411:9411 \
  -p 14250:14250 \
  em135/humio-jaeger-plugin:latest

sleep 2
docker logs jaeger