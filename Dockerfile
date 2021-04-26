#FROM golang:1.14 AS build
#WORKDIR /src
#ADD /humio /src/humio
#ADD /plugin /src/plugin
#ADD /go.mod /src
#ADD /go.sum /src
#ADD /main.go /src
#
#RUN export GOOS=linux
#RUN export GOARCH=amd64
#RUN CGO_ENABLED=0 go build -v -ldflags '-extldflags "-static"'
#
##FROM alpine:latest as certs
##RUN apk --update add ca-certificates

FROM jaegertracing/all-in-one:1.21.0
ENV SPAN_STORAGE_TYPE="grpc-plugin"
ENV GRPC_STORAGE_PLUGIN_BINARY="/go/bin/HumioJaegerStoragePlugin"
#COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
#COPY --from=build /src/HumioJaegerStoragePlugin /go/bin
COPY HumioJaegerStoragePlugin /go/bin



# Prøv at kun overføre nødvendige filer
# Prøv med certs
# Prøv med ny version