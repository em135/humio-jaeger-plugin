FROM golang:1.16.3 AS build
WORKDIR /src
ADD . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o humio-plugin -v -ldflags '-extldflags "-static"'

FROM jaegertracing/jaeger-query:latest
ENV SPAN_STORAGE_TYPE="grpc-plugin" \
    GRPC_STORAGE_PLUGIN_BINARY="/go/bin/humio-plugin"
COPY --from=build /src/humio-plugin /go/bin
