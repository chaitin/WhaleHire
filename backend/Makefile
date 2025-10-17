PLATFORM=linux/amd64
OUTPUT=type=docker,dest=${HOME}/tmp/whalehire_server.tar
GOCACHE=${HOME}/.cache/go-build
GOMODCACHE=${HOME}/go/pkg/mod
REGISTRY=portus.in.chaitin.net/whalehire
VERSION=v0.1.8
BUILD_TIME=${shell date -u +"%Y-%m-%dT%H:%M:%SZ"}
GIT_COMMIT=${shell git rev-parse HEAD}

# make build PLATFORM= TAG= OUTPUT= GOCACHE=
image: #swag wire
	docker buildx build \
	  -f build/Dockerfile \
	  --build-arg GOCACHE=${GOCACHE} \
	  --build-arg GOMODCACHE=${GOMODCACHE} \
	  --build-arg REPO_COMMIT=$(shell git rev-parse HEAD) \
	  --build-arg VERSION=${VERSION} \
	  --build-arg BUILD_TIME=${BUILD_TIME} \
	  --build-arg GIT_COMMIT=${GIT_COMMIT} \
	  --platform ${PLATFORM} \
	  --tag ${REGISTRY}/backend:${VERSION} \
	  --tag $(REGISTRY)/backend:latest \
	  .

wire:
	wire cmd/wire.go cmd/main.go

swag:
	swag fmt -d internal && swag init --pd -g cmd/main.go -ot "json"

migrate_sql:
	migrate create -ext sql -dir migration -seq ${SEQ}

# 生成迁移文件（基于当前Ent schema）
generate-migration:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make generate-migration NAME=migration_name"; \
		exit 1; \
	fi
	cd script && go run create_migration.go $(NAME)

# Go 代码格式化和 lint 相关命令
.PHONY: fmt lint vet tidy check-fmt install-tools

# 格式化 Go 代码
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	goimports -w .

# 检查代码格式是否正确
check-fmt:
	@echo "Checking Go code format..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted:"; \
		gofmt -l .; \
		echo "Please run 'make fmt' to format them."; \
		exit 1; \
	fi

# 运行 go vet 检查
vet:
	@echo "Running go vet..."
	go vet ./...

# 运行 golangci-lint
lint:
	@echo "Running golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found. Please install it first: make install-tools"; \
		exit 1; \
	fi
	golangci-lint run

# 整理 go.mod 文件
tidy:
	@echo "Tidying go.mod..."
	go mod tidy

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行所有代码质量检查
check: tidy fmt vet lint
	@echo "All code quality checks passed!"
