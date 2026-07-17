# Build/install targets for lazymux. Run from this directory
# (`make -C build <target>` or `cd build && make <target>`).
#
#   make build        build ./bin/lazymux
#   make dev           build ./bin/lazymux-dev, sandboxed to ~/lazymux-dev
#   make install       go-install lazymux to $GOBIN
#   make install-dev   go-install lazymux-dev to $GOBIN
#   make clean         remove ./bin

MODULE    := github.com/bkenks/lazymux
ROOT      := $(CURDIR)/..
BIN_DIR   := $(CURDIR)/bin
GOBIN     := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN     := $(shell go env GOPATH)/bin
endif
VERSION   := $(shell git -C $(ROOT) describe --tags --always --dirty 2>/dev/null || echo dev)

DEV_DIR_NAME := lazymux-dev

.PHONY: build dev install install-dev clean

build:
	mkdir -p $(BIN_DIR)
	cd $(ROOT) && go build -ldflags "-X main.buildVersion=$(VERSION)" -o $(BIN_DIR)/lazymux .

dev:
	mkdir -p $(BIN_DIR)
	cd $(ROOT) && go build -ldflags "-X main.buildVersion=$(VERSION)-dev -X $(MODULE)/internal/config.dirName=$(DEV_DIR_NAME)" -o $(BIN_DIR)/lazymux-dev .

install: build
	install -m 0755 $(BIN_DIR)/lazymux $(GOBIN)/lazymux

install-dev: dev
	install -m 0755 $(BIN_DIR)/lazymux-dev $(GOBIN)/lazymux-dev

clean:
	rm -rf $(BIN_DIR)
