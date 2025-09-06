PLATFORM=linux/amd64
TAG=main
OUTPUT=type=docker,dest=${HOME}/tmp/whalehire_server.tar
GOCACHE=${HOME}/.cache/go-build
GOMODCACHE=${HOME}/go/pkg/mod
REGISTRY=ptonlix/whalehire
VERSION=dev-${shell git rev-parse HEAD}
BUILD_TIME=${shell date -u +"%Y-%m-%dT%H:%M:%SZ"}
GIT_COMMIT=${shell git rev-parse HEAD}

# make build PLATFORM= TAG= OUTPUT= GOCACHE=
image: swag wire
	docker buildx build \
	  -f build/Dockerfile \
	  --build-arg GOCACHE=${GOCACHE} \
	  --build-arg GOMODCACHE=${GOMODCACHE} \
	  --build-arg REPO_COMMIT=$(shell git rev-parse HEAD) \
	  --build-arg VERSION=${VERSION} \
	  --build-arg BUILD_TIME=${BUILD_TIME} \
	  --build-arg GIT_COMMIT=${GIT_COMMIT} \
	  --platform ${PLATFORM} \
	  --tag ${REGISTRY}/backend:${TAG} \
	  --output ${OUTPUT} \
	  .

wire:
	wire cmd/wire.go cmd/main.go

swag:
	swag fmt -d internal && swag init --pd -g cmd/main.go -ot "json"

migrate_sql:
	migrate create -ext sql -dir migration -seq ${SEQ}