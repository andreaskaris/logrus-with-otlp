FROM golang:1.21 AS build-stage

WORKDIR /app
COPY . .
RUN make clean && make build

# Deploy the application binary into a lean image
FROM alpine:latest AS main
WORKDIR /
COPY --from=build-stage /app/_output/logrus-with-otlp /usr/local/bin/logrus-with-otlp
