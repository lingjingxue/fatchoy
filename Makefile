# Copyright © 2020-present ichenq@outlook.com All rights reserved.
# Distributed under the terms and conditions of the BSD License.
# See accompanying files LICENSE.

PWD = $(shell pwd)
GOBIN = $(PWD)/bin
GO ?= go

GO_PKG_LIST := $(shell go list ./...)

test:
	docker-compose up -d
	$(GO) test -v ./...
	docker-compose down

clean:
	$(GO) clean

.PHONY: clean build test all
