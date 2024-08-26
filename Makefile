IMAGE ?= quay.io/akaris/logrus-with-otlp

.PHONY: clean
clean:
	rm -f _output/*

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -o _output/logrus-with-otlp

.PHONY: container
container:
	podman build -t $(IMAGE) .
