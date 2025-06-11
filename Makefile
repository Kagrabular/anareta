BINARY_NAME := manager
IMG := ghcr.io/Kagrabular/anareta-operator:0.1.0

all: generate manifests fmt vet test

run: generate fmt vet
	go run main.go

build: generate fmt vet
	go build -o bin/$(BINARY_NAME) main.go

docker-build: build
	docker build -t $(IMG) .

docker-push: docker-build
	docker push $(IMG)

generate:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.0 \
		object:headerFile="hack/boilerplate.go.txt" paths="./api/..."

manifests:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.0 \
		crd:crdVersions=v1 paths="./api/..." output:crd:dir=config/crd/bases

fmt:
	go fmt ./...

vet:
	go vet ./...

test: test-unit test-integration

test-unit:
	go test ./controllers -v

test-integration:
	go test ./test/integration -v

helm-package:
	helm package charts/anareta-operator -d artifacts/

clean:
	rm -rf bin artifacts

.PHONY: all run build docker-build docker-push generate manifests fmt vet test test-unit test-integration helm-package clean
