.PHONY: build release-build unit-test integration-test generate push-only push

IMAGE_REGISTRY:=docker.io
IMAGE_TAG:=latest

IMAGE:=planetscale/vitess-operator

IMAGE_NAME:=$(IMAGE_REGISTRY)/$(IMAGE)

OPERATOR_SDK_VERSION:=v0.10.0

# Enable Go modules
export GO111MODULE=on

# Regular operator-sdk build is good for development because it does the actual
# build outside Docker, so it uses your cached modules.
build:
	operator-sdk-$(OPERATOR_SDK_VERSION) build $(IMAGE_NAME):$(IMAGE_TAG) --image-build-args '--no-cache'

# Release build is slow but self-contained (doesn't depend on anything in your
# local machine). We use this for automated builds that we publish.
release-build:
	docker build -f build/Dockerfile.release -t $(IMAGE_NAME):$(IMAGE_TAG) .

unit-test:
	pkgs="$$(go list ./... | grep -v '/test/integration/')" && \
		go test -i $${pkgs} && \
		go test $${pkgs}

integration-test:
	tools/get-kube-binaries.sh
	go test -i ./test/integration/...
	PATH="$(PWD)/tools/_bin:$(PATH)" go test -v -timeout 5m ./test/integration/... -args --logtostderr -v=6

generate:
	operator-sdk-$(OPERATOR_SDK_VERSION) generate k8s
	operator-sdk-$(OPERATOR_SDK_VERSION) generate openapi
	go run github.com/ahmetb/gen-crd-api-reference-docs -api-dir ./pkg/apis -config docs/api/config.json -template-dir docs/api/template -out-file docs/api/index.html

push-only: DATE=$(shell date -I)
push-only: GITHASH=$(shell git rev-parse HEAD)
push-only:
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):$(DATE)-$(GITHASH)
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_NAME):$(DATE)-$(GITHASH)

push: build
push: push-only
