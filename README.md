# humio-jaeger-plugin
 This repository contains the implementation of a gRPC storaget plugin for Jaeger, with the ability to use Humio as a storage backend. Running Jaeger with this storage plugin required the following environment variables:

 - `API_TOKEN`: The API token assigned to your account on Humio
 - `HUMIO_ENDPOINT`: The endpoint at which your Humio instance is hosted, such as `https://cloud.humio.com/`
 - `HUMIO_REPOSITORY`: The name of the repository to pull data from
