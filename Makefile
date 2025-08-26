APP = s3-iam-cosi-driver

# Define high-level version
BASE_VERSION := 2.0.0

# Derive commit hash
GIT_COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Combine high-level version and commit hash
VERSION := $(BASE_VERSION)-$(GIT_COMMIT_HASH)

URL ?= icr.io
NAMESPACE ?= cosi-research
IMAGE_TAG_BASE ?= $(URL)/$(NAMESPACE)/$(APP)
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)

CRI := $(shell command -v docker 2> /dev/null || command -v podman 2> /dev/null || echo "none")

# Detect the host OS and architecture
HOST_OS := $(shell uname -s | tr A-Z a-z)
HOST_ARCH := $(shell uname -m)

# Map uname architectures to Go architectures
ifeq ($(HOST_ARCH), x86_64)
    GOARCH := amd64
else ifeq ($(HOST_ARCH), arm64)
    GOARCH := arm64
else
    GOARCH := $(HOST_ARCH)
endif

DEPLOY_ENV ?= prod

show-image-tag: ## Show the image tag
	@echo $(IMG)

# Define a target for checking the container runtime
check-runtime:
ifeq ($(CRI),none)
	$(error "No container runtime found. Please install Docker or Podman.")
else
	@echo "Using $(CRI) as the container runtime."
endif

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: deps
deps:  ## Fetch dependencies
	go mod tidy
	go mod vendor

# Build target
.PHONY: build
build: fmt vet deps  ## Build the Linux/amd64 binary for Kubernetes
	@echo "Building for linux/amd64..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-X main.version=$(VERSION) -extldflags "-static"' -o ./bin/$(APP) ./cmd/main.go

.PHONY: patch-crd
patch-crd: ## Patch CRD to allow nested parameters
	kubectl patch crd bucketaccessclasses.objectstorage.k8s.io \
	  --type=json \
	  -p "[{\
	    \"op\": \"replace\",\
	    \"path\": \"/spec/versions/0/schema/openAPIV3Schema/properties/parameters\",\
	    \"value\": {\
	      \"type\": \"object\",\
	      \"x-kubernetes-preserve-unknown-fields\": true,\
	      \"additionalProperties\": true,\
	      \"properties\": {\
	        \"defaultPolicy\": {\
	          \"type\": \"object\",\
	          \"x-kubernetes-preserve-unknown-fields\": true,\
	          \"additionalProperties\": true,\
	          \"properties\": {\
	            \"allow\": {\
	              \"type\": \"object\",\
	              \"x-kubernetes-preserve-unknown-fields\": true,\
	              \"additionalProperties\": true\
	            }\
	          }\
	        },\
	        \"requestPolicy\": {\
	          \"type\": \"object\",\
	          \"x-kubernetes-preserve-unknown-fields\": true,\
	          \"additionalProperties\": true,\
	          \"properties\": {\
	            \"allow\": {\
	              \"type\": \"object\",\
	              \"x-kubernetes-preserve-unknown-fields\": true,\
	              \"additionalProperties\": true\
	            }\
	          }\
	        }\
	      }\
	    }\
	  }]"

.PHONY: deploy
deploy: patch-crd ## Deploy (DEPLOY_ENV=dev for dev overlay, default is prod)
ifeq ($(DEPLOY_ENV),dev)
	kubectl apply -k overlays/dev
else
	kubectl apply -k resources/
endif

.PHONY: undeploy
undeploy: ## Undeploy (DEPLOY_ENV=dev for dev overlay, default is prod)
ifeq ($(DEPLOY_ENV),dev)
	kubectl delete -k overlays/dev
else
	kubectl delete -k resources/
endif

.PHONY: hot-reload
hot-reload: build ## Build, copy, and start driver in dev pod
	@POD=$$(kubectl get pod -l mode=dev -o jsonpath='{.items[0].metadata.name}'); \
	echo "Using pod: $$POD"; \
	kubectl cp ./bin/$(APP) $$POD:/s3-iam-cosi-driver -c $(APP); \
	echo "Starting driver..."; \
	kubectl exec -it $$POD -c $(APP) -- /s3-iam-cosi-driver

.PHONY: docker-build
docker-build: deps check-runtime  ## Build the Docker image
	${CRI} build --no-cache --build-arg TARGETOS=linux --build-arg TARGETARCH=amd64 -t ${IMG} .
	${CRI} tag ${IMG} ${IMAGE_TAG_BASE}:latest  # Tag the image as 'latest'

.PHONY: docker-push
docker-push: check-runtime ## Push docker image with the manager.
	${CRI} push ${IMG}  # Push the versioned tag (e.g., with hash or specific tag)
	${CRI} push ${IMAGE_TAG_BASE}:latest  # Push the 'latest' tag

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
