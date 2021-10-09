# Copyright © 2020-present ichenq@outlook.com All rights reserved.
# Distributed under the terms and conditions of the BSD License.
# See accompanying files LICENSE.

PWD = $(shell pwd)
GOBIN = $(PWD)/bin
GO ?= go

GO_PKG_LIST := $(shell go list ./...)

test:
	$(GO) test -v ./...

clean:
	$(GO) clean

.PHONY: clean build test all
