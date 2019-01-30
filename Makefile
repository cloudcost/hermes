SHELL := /bin/bash
DATE := $(shell date +%Y%m%d-%H:%M:%S)
HASH := $(shell git rev-parse HEAD)
GOVERSION := $(shell go version)
LDFLAGS := -X 'main.date=${DATE}' -X 'main.hash=${HASH}' -X 'main.goversion=${GOVERSION}'

install:
	-rm ${GOPATH}/bin/mercurius
	go install -ldflags "${LDFLAGS}" github.com/itsubaki/hermes/cmd/hermes

test:
	mkdir -p /var/tmp/hermes/{awsprice,costexp,reserved}
	GOCACHE=off go test -cover $(shell go list ./... | grep -v /vendor/ | grep -v /build/) -v -run TestSerialize* -timeout 20m
	go test -cover $(shell go list ./... | grep -v /vendor/ | grep -v /build/) -v
