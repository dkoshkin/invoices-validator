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
		-e VERSION=$(VERSION)               \
		-e BUILD_DATE="$(BUILD_DATE)"       \
		invoices-validator-base:$(BUILDER_CHECKSUM)  \
		make build-binary-local

.PHONY: build-binary-local
build-binary-local:
	go build \
		-ldflags "-X main.version=$(VERSION) -X 'main.buildDate=$(BUILD_DATE)'" \
		-o bin/invoices-validator-$(GOOS)-$(GOARCH) cmd/main.go

.PHONY: build-all
build-all: build-container build-binaries

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