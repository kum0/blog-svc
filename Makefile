BUILD_DIR = build
SERVICES = gateway
DOCKERS_CLEANBUILD = $(addprefix docker_prod_,$(SERVICES))
DOCKERS_DEV = $(addprefix docker_dev_,$(SERVICES))
DOCKERS_DEBUG = $(addprefix docker_debug_,$(SERVICES))
STAGES = dev debug prod
COMPOSEUP = $(addsuffix -compose-up,$(STAGES))
COMPOSEDOWN = $(addsuffix -compose-down,$(STAGES))
CGO_ENABLED ?= 0
GOOS ?= linux
# GOOS ?= darwin
TAG_PREFIX = mircosvc

define compile_service
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -ldflags "-s -w" -o
	${BUILD_DIR}/$(TAG_PREFIX)-$(1) cmd/$(1)/main.go
endef

define compile_debug_service
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -gcflags "all=-N -l" -o
	${BUILD_DIR}/$(TAG_PREFIX)-$(1) cmd/$(1)/main.go
endef

define make_docker_cleanbuild
	docker build --no-cache --build-arg SVC_NAME=$(subst docker_prod_,,$(1)) --tag=kum0/$(TAG_PREFIX)-$(subst
	docker_prod_,,$(1)) -f deployments/docker/Dockerfile .
endef

define make_docker_dev
	docker build --build-arg SVC_NAME=$(subst docker_dev_,,$(1)) --tag=kum0/$(TAG_PREFIX)-$(subst docker_dev_,,$(1))
	 -f deployments/docker/Dockerfile.dev ./build
endef

define make_docker_debug
	docker build --build-arg SVC_NAME=$(subst docker_debug_,,$(1)) --tag=kum0/$(TAG_PREFIX)-debug-$(subst
	docker_debug_,,$(1)) -f deployments/docker/Dockerfile.debug ./build
endef

all: $(SERVICES)

.PHONY: all $(SERVICES) dockers dockers_dev dockers_debug

cleandocker: cleanghost
	# Stop all containers (if running)
	docker-compose -f deployments/docker/docker-compose.yaml stop
	# Remove containers
	docker ps -f name=gokitconsul -aq | xargs docker rm
	# Remove old images
	docker images -q kum0/$(TAG_PREFIX)-* | xargs docker rmi

# Clean ghost docker images
cleanghost:
	# Remove exited containers
	docker ps -f status=dead -f status=exited -aq | xargs docker rm -v
	# Remove unused images
	docker images -f dangling=true -q | xargs docker rmi
	# Remove unused volumes
	docker volume ls -f dangling=true -q | xargs docker volume rm

install:
	cp ${BUILD_DIR}/* $(GOBIN)

test:
	go test -v -race -tags test $(shell go list ./... | grep -v 'vendor\|cmd')

# compile proto
PD_SOURCES:=$(shell find ./pb -type d)
proto:
	@for var in $(PD_SOURCES); do \
		if [ -f "$$var/compile.sh" ]; then \
			cd $$var && ./compile.sh; \
			echo "complie $$var/$$(basename $$var).proto"; \
			cd $(PWD); \
		fi \
	done

# Regenerates OPA data from rego files
HAVE_GO_BINDATA := $(shell command -v go-bindata 2> /dev/null)
generate:
ifndef HAVE_GO_BINDATA
	@echo "requires 'go-bindata' (go get -u github.com/kevinburke/go-bindata/go-bindata)"
	@exit 1 # fail
else
	go generate ./...
endif

$(SERVICES):
	$(call compile_service,$(@))

$(DOCKERS_CLEANBUILD):
	$(call make_docker_cleanbuild,$(@))

$(DOCKERS_DEV):
	$(call compile_service,$(subst docker_dev_,,$(@)))
	$(call make_docker_dev,$(subst docker_dev_,,$(@)))

$(DOCKERS_DEBUG):
	$(call compile_debug_service,$(subst docker_debug_,,$(@)))
	$(call make_docker_debug,$(subst docker_debug_,,$(@)))

services: $(SERVICES)

dockers: $(DOCKERS_CLEANBUILD)

dockers_debug: $(DOCKERS_DEBUG)

dockers_dev: $(DOCKERS_DEV)

define make_docker_compose_up
	@if [ $(1) == prod ]; then \
		echo "docker-compose -f deployments/docker/docker-compose.yaml up -d"; \
		docker-compose -f deployments/docker/docker-compose.yaml up -d; \
	else \
		echo "docker-compose -f deployments/docker/docker-compose-$(1).yaml up -d"; \
		docker-compose -f deployments/docker/docker-compose-$(1).yaml up -d; \
	fi
endef

define make_docker_compose_down
	@if [ $(1) == prod ]; then \
		echo "docker-compose -f deployments/docker/docker-compose.yaml down"; \
		docker-compose -f deployments/docker/docker-compose.yaml down; \
	else \
		echo "docker-compose -f deployments/docker/docker-compose-$(1).yaml down"; \
		docker-compose -f deployments/docker/docker-compose-$(1).yaml down; \
	fi
endef

$(COMPOSEUP):
	$(call make_docker_compose_up,$(subst -compose-up,,$(@)))

$(COMPOSEDOWN):
	$(call make_docker_compose_down,$(subst -compose-down,,$(@)))

u:
	make dev-compose-up
#	make debug-compose-up
#	make prod-compose-up

d:
	make dev-compose-down
#	make debug-compose-down
#	make prod-compose-down

restart: d u
