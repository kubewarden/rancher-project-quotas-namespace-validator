SOURCE_FILES := $(shell find . -type f -name '*.go')

CONTAINER_IMAGE = tinygo:patched

# We cannot use the official tinygo container image until
# this issue is closed: https://github.com/tinygo-org/tinygo/issues/3501
build-container:
	DOCKER_BUILDKIT=1 docker build . -t $(CONTAINER_IMAGE)

policy.wasm: $(SOURCE_FILES) go.mod go.sum types_easyjson.go
	docker run \
		--rm \
		-e GOFLAGS="-buildvcs=false" \
		-v ${PWD}:/src \
		-w /src \
		$(CONTAINER_IMAGE) \
		tinygo build -o policy.wasm -target=wasi -no-debug .

annotated-policy.wasm: policy.wasm metadata.yml
	kwctl annotate -m metadata.yml -o annotated-policy.wasm policy.wasm

.PHONY: generate-easyjson
types_easyjson.go: types.go
	docker run \
		--rm \
		-v ${PWD}:/src \
		-w /src \
		golang:1.17-alpine ./hack/generate-easyjson.sh

.PHONY: test
test: types_easyjson.go
	go test -v

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run


.PHONY: e2e-tests
e2e-tests: annotated-policy.wasm
	bats e2e.bats

.PHONY: clean
clean:
	go clean
	rm -f policy.wasm annotated-policy.wasm
