ifeq ($(origin VERSION), undefined)
    VERSION := $(shell git describe --tags --always)
endif
ifeq ($(origin BUILD_DATE), undefined)
    BUILD_DATE := $(shell date -u)
endif
ifeq ($(origin GOOS), undefined)
    GOOS := $(shell go env GOOS)
endif
ifeq ($(origin GOARCH), undefined)
    GOARCH := $(shell go env GOARCH)
endif

CONTAINER = dkoshkin/invoices-validator
PKG = github.com/dkoshkin/invoices-validator
CMD ?= cli
BUILDER_CHECKSUM = $$(shasum go.sum | awk '{ print $$1 }' | cut -c1-6)

.PHONY: build-container
build-container:
	docker build                                \
	    --build-arg VERSION="$(VERSION)"        \
		--build-arg BUILD_DATE="$(BUILD_DATE)"  \
		-f build/docker/Dockerfile -t $(CONTAINER) .

.PHONY: builder
builder:
	docker build                                \
	    --target builder_base                   \
	    -f build/docker/Dockerfile -t invoices-validator-base:$(BUILDER_CHECKSUM) .

.PHONY: build-binaries
build-binaries:
	@$(MAKE) GOOS=darwin GOARCH=amd64 build-binary
	@$(MAKE) GOOS=linux GOARCH=amd64 build-binary

.PHONY: build-binary
build-binary:
	@docker run                             \
		--rm                                \
		-u root:root                        \
		-v "$(shell pwd)":/src/$(PKG)       \
		-w /src/$(PKG)                      \
		-e CGO_ENABLED=0                    \
		-e GOOS=$(GOOS)                     \
		-e GOARCH=$(GOARCH)                 \
		-e CMD=$(CMD)                       \
		-e VERSION=$(VERSION)               \
		-e BUILD_DATE="$(BUILD_DATE)"       \
		invoices-validator-base:$(BUILDER_CHECKSUM)  \
		make build-binary-local

.PHONY: build-binary-local
build-binary-local:
	go build \
		-a -ldflags "-X main.version=$(VERSION) -X 'main.buildDate=$(BUILD_DATE)'" \
		-o bin/invoices-validator-$(CMD)-$(GOOS)-$(GOARCH) cmd/$(CMD)/main.go

.PHONY: build-all
build-all: build-container build-binaries

.PHONY: lambda-zip
lambda-zip:
	rm -f bin/invoices-validator-handler.zip
	@$(MAKE) GOOS=linux GOARCH=amd64 CMD=lambda build-binary
	mv bin/invoices-validator-lambda-linux-amd64 bin/invoices-validator.handler
	cd bin/ && zip invoices-validator-handler.zip invoices-validator.handler
	rm -f bin/invoices-validator.handler

.PHONY: lambda-publish
lambda-publish: lambda-zip
	aws lambda update-function-code \
	--function-name invoices-validator \
	--zip-file fileb://bin/invoices-validator-handler.zip

.PHONY: test
test:
	@docker run                             \
		--rm                                \
		-u root:root                        \
		-v "$(shell pwd)":/src/$(PKG)       \
		-w /src/$(PKG)                      \
		invoices-validator-base:$(BUILDER_CHECKSUM)  \
		make test-local

.PHONY: test-local
test-local:
	go test -v ./cmd/... ./pkg/...

.PHONY: push
push:
	docker push $(CONTAINER):latest

.PHONY: tag
tag:
	docker tag $(CONTAINER) $(CONTAINER):$(VERSION)

.PHONY: tag-and-push
tag-and-push: tag
	docker push $(CONTAINER):$(VERSION)