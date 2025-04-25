
PREFIX = /usr
export GOCACHE := /tmp/gocache
GOPKG_PREFIX = linglong-tools

GOCMD=go

GOBUILD=$(GOCMD) build -mod vendor $(GO_BUILD_FLAGS)
GOCLEAN=$(GOCMD) clean

.PHONY: all

all: build

build: out/linyaps-tools test

fmt:
	@echo ">Formatting code..."
	$(GOCMD) fmt ./...

out/linyaps-tools:
	@echo ">Compiling linyaps-tools..."
	$(GOBUILD) -o out/linyaps-tools

test:
	@echo ">Run test..."
	$(GOCMD) test -v ./...


install:
	@echo ">Installing..."
	install -Dm755 out/linyaps-tools ${DESTDIR}${PREFIX}/bin/linyaps-tools

clean:
	@echo ">Cleaning up..."
	@rm -rf out
	@${GOCLEAN}
