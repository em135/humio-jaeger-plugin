FROM golang:1.16.3 AS build
WORKDIR /src
ADD . /src
RUN export GOOS=linux
RUN export GOARCH=amd64
RUN CGO_ENABLED=0 go build -o humio-plugin -v -ldflags '-extldflags "-static"'

FROM jaegertracing/all-in-one:latest
ENV SPAN_STORAGE_TYPE="grpc-plugin"
ENV GRPC_STORAGE_PLUGIN_BINARY="/go/bin/humio-plugin"
COPY --from=build /src/humio-plugin /go/bin