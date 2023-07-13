FILES = $(shell find . -type f -name '*.go')

default: help

help:                   ## Display this help message
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

init:                   ## Install development tools
	cd tools && go generate -x -tags=tools

build:                ## Build binaries
	go build -race -o bin/percona-everest-backend ./cmd

build-debug:                ## Build binaries
	go build -tags debug -race -o bin/percona-everest-backend-debug ./cmd

gen:                    ## Generate code
	go generate ./...
	make format

format:                 ## Format source code
	bin/gofumpt -l -w $(FILES)
	bin/goimports -local github.com/percona/percona-everest-backend -l -w $(FILES)
	bin/gci write --section Standard --section Default --section "Prefix(github.com/percona/percona-everest-backend)" $(FILES)

check:                  ## Run checks/linters for the whole project
	bin/go-consistent -pedantic ./...
	LOG_LEVEL=error bin/golangci-lint run

test:                   ## Run tests
	go test -race -timeout=10m ./...

test-cover:             ## Run tests and collect per-package coverage information
	go test -race -timeout=10m -count=1 -coverprofile=cover.out -covermode=atomic ./...

test-crosscover:        ## Run tests and collect cross-package coverage information
	go test -race -timeout=10m -count=1 -coverprofile=crosscover.out -covermode=atomic -p=1 -coverpkg=./... ./...

run: build            ## Run binary
	bin/percona-everest-backend

run-debug: build-debug            ## Run binary
	bin/percona-everest-backend-debug

local-env-up:                 ## Start development environment
	docker-compose up --detach --remove-orphans

local-env-down:               ## Stop development environment
	docker-compose down --volumes --remove-orphans

cert:                   ## Install dev TLS certificates
	mkcert -install
	mkcert -cert-file=dev-cert.pem -key-file=dev-key.pem percona-everest-backend percona-everest-backend.localhost 127.0.0.1

k8s: ## Create a local minikube cluster
	minikube start --nodes=3 --cpus=4 --memory=4g --apiserver-names host.docker.internal
	minikube addons disable storage-provisioner
	kubectl delete storageclass standard
	kubectl apply -f ./dev/kubevirt-hostpath-provisioner.yaml
