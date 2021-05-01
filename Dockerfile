FROM golang:1.16.3 AS build
WORKDIR /src
ADD . /src
#RUN export GOOS=linux
#RUN export GOARCH=amd64
#RUN CGO_ENABLED=0 go build -v -ldflags '-extldflags "-static"'

RUN CGO_ENABLED=0 go build -o humio-plugin -v -ldflags '-extldflags "-static"'
#FROM alpine:latest as certs
#RUN apk --update add ca-certificates

FROM jaegertracing/all-in-one:latest
ENV SPAN_STORAGE_TYPE="grpc-plugin"
ENV GRPC_STORAGE_PLUGIN_BINARY="/go/bin/humio-plugin"
#COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /src/humio-plugin /go/bin