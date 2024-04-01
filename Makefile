.PHONY: tidy build up down image vendor test coverage

MONGODB_VERSION		:= 6.0-ubi8
REDIS_VERSION       := 7.2.0-v9
BRANCH=$(shell bash ./scripts/branch.sh)
export BRANCH

SCHEMA_ROOT			:= schema
REST_SCHEMA_ROOT	:= $(SCHEMA_ROOT)/rest

TINY_URL_SCHEMA		:= $(REST_SCHEMA_ROOT)/v0/url-svc.yaml
TINY_URL_ROOT_V0	:= types/api/rest/v0

TINY_URL_SEVER		:= $(TINY_URL_ROOT_V0)/server.gen.go
TINY_URL_MODEL		:= $(TINY_URL_ROOT_V0)/model.gen.go
TINY_URL_SPEC		:= $(TINY_URL_ROOT_V0)/spec.gen.go

OAPICODEGEN		:= $(GOPATH)/bin/oapi-codegen

GENERATE_MODELS := ${OAPICODEGEN} -generate types,skip-prune
GENERATE_SERVER := ${OAPICODEGEN} -generate server,skip-prune
GENERATE_SPEC	:= ${OAPICODEGEN} -generate spec,skip-prune

GENERATE_LIST	:= $(TINY_URL_SEVER) $(TINY_URL_MODEL) $(TINY_URL_SPEC)

$(TINY_URL_ROOT_V0)/model.gen.go: $(TINY_URL_SCHEMA)
	${GENERATE_MODELS} -package v0 $< > $@

$(TINY_URL_ROOT_V0)/server.gen.go: $(TINY_URL_SCHEMA)
	${GENERATE_SERVER} -package v0 $< > $@

$(TINY_URL_ROOT_V0)/spec.gen.go: $(TINY_URL_SCHEMA)
	${GENERATE_SPEC} -package v0 $< > $@

generate: ${GENERATE_LIST}

tidy:
	go mod tidy

build:
	go build .

up: vendor image
	docker compose -f docker/docker-compose.yaml up -d

down:
	docker compose -f docker/docker-compose.yaml down -v

image:
	docker build -f docker/Dockerfile -t tiny-url-svc:${BRANCH} .

vendor:
	go mod vendor

test:
	go test ./...

coverage:
	go test -cover ./... -coverprofile=c.out
	go tool cover -html="c.out"