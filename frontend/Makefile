# WhaleHire Frontend Makefile

# 构建配置
PLATFORM=linux/amd64
REGISTRY=portus.in.chaitin.net/whalehire
VERSION=v0.1.18
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse HEAD)

# Docker 构建参数
DOCKER_BUILDKIT=1

.PHONY: help build image push clean install dev lint fmt check test

# 默认目标
help:
	@echo "WhaleHire Frontend Build Commands:"
	@echo ""
	@echo "Development:"
	@echo "  install     - Install dependencies"
	@echo "  dev         - Start development server"
	@echo "  build       - Build production bundle"
	@echo "  preview     - Preview production build"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint        - Run ESLint"
	@echo "  fmt         - Format code with Prettier"
	@echo "  check       - Type check with TypeScript"
	@echo "  test        - Run tests (if available)"
	@echo ""
	@echo "Docker:"
	@echo "  image       - Build Docker image for AMD64"
	@echo "  push        - Push image to registry"
	@echo "  clean       - Clean build artifacts and Docker images"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION     - Image version (default: $(VERSION))"
	@echo "  PLATFORM    - Target platform (default: $(PLATFORM))"
	@echo "  REGISTRY    - Docker registry (default: $(REGISTRY))"

# 安装依赖
install:
	@echo "Installing dependencies..."
	npm ci --silent

# 开发服务器
dev:
	@echo "Starting development server..."
	npm run dev

# 构建生产版本
build:
	@echo "Building production bundle..."
	npm run clean
	npm run build:prod

# 预览生产版本
preview:
	@echo "Starting preview server..."
	npm run preview:prod

# 代码检查和格式化
lint:
	@echo "Running ESLint..."
	npm run lint

fmt:
	@echo "Formatting code with Prettier..."
	npx prettier --write "src/**/*.{ts,tsx,js,jsx,json,css,md}"

check:
	@echo "Type checking with TypeScript..."
	npm run check

# 测试（如果有测试的话）
test:
	@echo "Running tests..."
	@if [ -f "package.json" ] && grep -q '"test"' package.json; then \
		npm test; \
	else \
		echo "No tests configured"; \
	fi

# 构建 Docker 镜像
image:
	@echo "Building Docker image for $(PLATFORM)..."
	@echo "Registry: $(REGISTRY)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) docker buildx build \
		--platform $(PLATFORM) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--tag $(REGISTRY)/frontend:$(VERSION) \
		--tag $(REGISTRY)/frontend:latest \
		.

# 推送镜像到仓库
push: image
	@echo "Pushing image to registry..."
	docker push $(REGISTRY)/frontend:$(VERSION)
	docker push $(REGISTRY)/frontend:latest

# 构建并推送（一步完成）
deploy: push
	@echo "Image deployed successfully!"
	@echo "Image: $(REGISTRY)/frontend:$(VERSION)"

# 清理构建产物和 Docker 镜像
clean:
	@echo "Cleaning build artifacts..."
	npm run clean
	@echo "Cleaning Docker images..."
	-docker rmi $(REGISTRY)/frontend:$(VERSION) 2>/dev/null || true
	-docker rmi $(REGISTRY)/frontend:latest 2>/dev/null || true
	@echo "Cleaning Docker build cache..."
	-docker builder prune -f

# 清理所有（包括 node_modules）
clean-all: clean
	@echo "Removing node_modules..."
	rm -rf node_modules
	rm -rf package-lock.json

# 重新安装依赖
reinstall: clean-all install

# 运行所有代码质量检查
quality-check: lint check fmt
	@echo "All code quality checks completed!"

# 完整构建流程（质量检查 + 构建 + 镜像）
ci: quality-check build image
	@echo "CI build completed successfully!"

# 显示构建信息
info:
	@echo "Build Information:"
	@echo "  Platform: $(PLATFORM)"
	@echo "  Registry: $(REGISTRY)"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Node Version: $(NODE_VERSION)"
	@echo "  Image Tag: $(REGISTRY)/frontend:$(VERSION)"