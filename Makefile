# Copyright © 2020-present ichenq@outlook.com All rights reserved.
# Distributed under the terms and conditions of the BSD License.
# See accompanying files LICENSE.

PWD = $(shell pwd)
GOBIN = $(PWD)/bin

GO_PKG_LIST := $(shell go list ./...)

# load .env if exist
ifneq (,$(wildcard ./.env))
    include .env
    export
endif


# docker-compose up -d
test:
	go test -v ./x/...
	go test -v ./codec
	go test -v ./codes
	go test -v ./debug
	go test -v ./packet
	go test -v ./sched
	go test -v ./qlog
	go test -v ./qnet
	go test -v ./discovery

clean:
	go clean

.PHONY: clean build test all
