SOURCE_FILES := $(shell find . -type f -name '*.go')
VERSION := $(shell git describe | cut -c2-)

# We cannot use the official tinygo container image until
# this issue is closed: https://github.com/tinygo-org/tinygo/issues/3501
CONTAINER_IMAGE = tinygo/tinygo:0.30.0

# TODO: drop this once we can use the official tinygo container image
# see comment from above
build-container:
	DOCKER_BUILDKIT=1 docker build . -t $(CONTAINER_IMAGE)

policy.wasm: $(SOURCE_FILES) go.mod go.sum
	# TODO: remove the -opt=0 once tinygo ships with a more recent version of
	# wasm-opt
	docker run \
		--rm \
		-e GOFLAGS="-buildvcs=false" \
		-v ${PWD}:/src \
		-w /src \
		$(CONTAINER_IMAGE) \
		tinygo build -o policy-no-opt.wasm -opt=0 -target=wasi -no-debug .
	# Note: requires binaryen >= 116 to be installed
	wasm-opt -Os policy-no-opt.wasm -o policy.wasm

artifacthub-pkg.yml: metadata.yml go.mod
	$(warning If you are updating the artifacthub-pkg.yml file for a release, \
	  remember to set the VERSION variable with the proper value. \
	  To use the latest tag, use the following command:  \
	  make VERSION=$$(git describe --tags --abbrev=0 | cut -c2-) annotated-policy.wasm)
	kwctl scaffold artifacthub \
	  --metadata-path metadata.yml --version $(VERSION) \
	  --output artifacthub-pkg.yml

annotated-policy.wasm: policy.wasm metadata.yml
	kwctl annotate -m metadata.yml -u README.md -o annotated-policy.wasm policy.wasm

.PHONY: test
test:
	go test -v

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run

.PHONY: e2e-tests-env-setup
e2e-tests-env-setup:
	./hack/setup-e2e-env.sh

.PHONY: e2e-tests-env-destroy
e2e-tests-env-destroy:
	./hack/destroy-e2e-env.sh

.PHONY: e2e-tests
e2e-tests: annotated-policy.wasm
	bats e2e.bats

.PHONY: clean
clean:
	go clean
	rm -f policy.wasm annotated-policy.wasm
