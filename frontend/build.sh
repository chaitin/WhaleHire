#!/bin/bash

# ================================
# WhaleHire 前端项目 Docker 构建脚本
# DevOps 优化版本 - 支持多环境构建
# ================================

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="wh-frontend"
DEFAULT_TAG="latest"
DEFAULT_REGISTRY=""
BUILD_CONTEXT="."

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
WhaleHire 前端项目 Docker 构建脚本

用法: $0 [选项]

选项:
    -e, --env ENV           构建环境 (development|staging|production) [默认: production]
    -t, --tag TAG           镜像标签 [默认: latest]
    -r, --registry REGISTRY 镜像仓库地址
    -p, --platform PLATFORM 目标平台 (linux/amd64,linux/arm64) [默认: linux/amd64]
    --push                  构建后推送到仓库
    --no-cache              不使用构建缓存
    --multi-arch            构建多架构镜像
    --build-arg KEY=VALUE   传递构建参数
    -v, --verbose           详细输出
    -h, --help              显示此帮助信息

环境变量配置:
    DOCKER_REGISTRY         默认镜像仓库
    BUILD_NUMBER            构建编号
    GIT_COMMIT              Git 提交哈希

示例:
    # 构建生产环境镜像
    $0 --env production --tag v1.0.0

    # 构建并推送到仓库
    $0 --env production --tag v1.0.0 --registry registry.example.com --push

    # 多架构构建
    $0 --env production --multi-arch --platform linux/amd64,linux/arm64

    # 开发环境构建
    $0 --env development --tag dev-latest --no-cache

EOF
}

# 解析命令行参数
parse_args() {
    ENV="production"
    TAG="$DEFAULT_TAG"
    REGISTRY="$DEFAULT_REGISTRY"
    PLATFORM="linux/amd64"
    PUSH=false
    NO_CACHE=false
    MULTI_ARCH=false
    VERBOSE=false
    BUILD_ARGS=()

    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--env)
                ENV="$2"
                shift 2
                ;;
            -t|--tag)
                TAG="$2"
                shift 2
                ;;
            -r|--registry)
                REGISTRY="$2"
                shift 2
                ;;
            -p|--platform)
                PLATFORM="$2"
                shift 2
                ;;
            --push)
                PUSH=true
                shift
                ;;
            --no-cache)
                NO_CACHE=true
                shift
                ;;
            --multi-arch)
                MULTI_ARCH=true
                shift
                ;;
            --build-arg)
                BUILD_ARGS+=("--build-arg" "$2")
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 验证环境
validate_environment() {
    if [[ ! "$ENV" =~ ^(development|staging|production)$ ]]; then
        log_error "无效的环境: $ENV"
        exit 1
    fi

    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装或不在 PATH 中"
        exit 1
    fi

    if [[ "$MULTI_ARCH" == true ]] && ! docker buildx version &> /dev/null; then
        log_error "Docker Buildx 未安装，无法进行多架构构建"
        exit 1
    fi
}

# 设置环境变量
setup_environment() {
    # Git 信息
    GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"
    GIT_BRANCH="${GIT_BRANCH:-$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')}"
    
    # 构建信息
    BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    BUILD_NUMBER="${BUILD_NUMBER:-$(date +%Y%m%d%H%M%S)}"
    
    # 镜像名称
    if [[ -n "$REGISTRY" ]]; then
        IMAGE_NAME="$REGISTRY/$PROJECT_NAME"
    else
        IMAGE_NAME="$PROJECT_NAME"
    fi
    
    FULL_IMAGE_NAME="$IMAGE_NAME:$TAG"

    log_info "构建配置:"
    log_info "  环境: $ENV"
    log_info "  镜像: $FULL_IMAGE_NAME"
    log_info "  平台: $PLATFORM"
    log_info "  Git 提交: $GIT_COMMIT"
    log_info "  Git 分支: $GIT_BRANCH"
    log_info "  构建时间: $BUILD_DATE"
}

# 根据环境设置构建参数
setup_build_args() {
    case "$ENV" in
        development)
            BUILD_ARGS+=(
                "--build-arg" "VITE_APP_ENV=development"
                "--build-arg" "VITE_DEBUG=true"
                "--build-arg" "VITE_ENABLE_DEVTOOLS=true"
                "--build-arg" "VITE_BUILD_SOURCEMAP=true"
                "--build-arg" "VITE_BUILD_MINIFY=false"
                "--build-arg" "VITE_API_BASE_URL=http://localhost:8888/api"
            )
            ;;
        staging)
            BUILD_ARGS+=(
                "--build-arg" "VITE_APP_ENV=staging"
                "--build-arg" "VITE_DEBUG=false"
                "--build-arg" "VITE_ENABLE_DEVTOOLS=false"
                "--build-arg" "VITE_BUILD_SOURCEMAP=true"
                "--build-arg" "VITE_BUILD_MINIFY=true"
                "--build-arg" "VITE_API_BASE_URL=https://staging.hire.chaitin.net/api"
            )
            ;;
        production)
            BUILD_ARGS+=(
                "--build-arg" "VITE_APP_ENV=production"
                "--build-arg" "VITE_DEBUG=false"
                "--build-arg" "VITE_ENABLE_DEVTOOLS=false"
                "--build-arg" "VITE_BUILD_SOURCEMAP=false"
                "--build-arg" "VITE_BUILD_MINIFY=true"
                "--build-arg" "VITE_API_BASE_URL=https://hire.chaitin.net/api"
            )
            ;;
    esac

    # 添加通用构建参数
    BUILD_ARGS+=(
        "--build-arg" "BUILD_DATE=$BUILD_DATE"
        "--build-arg" "GIT_COMMIT=$GIT_COMMIT"
        "--build-arg" "GIT_BRANCH=$GIT_BRANCH"
        "--build-arg" "BUILD_NUMBER=$BUILD_NUMBER"
    )
}

# 构建 Docker 镜像
build_image() {
    log_info "开始构建 Docker 镜像..."

    # 构建命令
    BUILD_CMD=("docker")
    
    if [[ "$MULTI_ARCH" == true ]]; then
        BUILD_CMD+=("buildx" "build")
        BUILD_CMD+=("--platform" "$PLATFORM")
        
        if [[ "$PUSH" == true ]]; then
            BUILD_CMD+=("--push")
        else
            BUILD_CMD+=("--load")
        fi
    else
        BUILD_CMD+=("build")
        BUILD_CMD+=("--platform" "$PLATFORM")
    fi

    # 添加构建参数
    BUILD_CMD+=("${BUILD_ARGS[@]}")
    
    # 添加标签
    BUILD_CMD+=("--tag" "$FULL_IMAGE_NAME")
    
    # 添加额外标签
    if [[ "$TAG" != "latest" ]]; then
        BUILD_CMD+=("--tag" "$IMAGE_NAME:latest")
    fi
    
    # 缓存选项
    if [[ "$NO_CACHE" == true ]]; then
        BUILD_CMD+=("--no-cache")
    fi
    
    # 构建上下文
    BUILD_CMD+=("$BUILD_CONTEXT")

    # 执行构建
    if [[ "$VERBOSE" == true ]]; then
        log_info "执行命令: ${BUILD_CMD[*]}"
    fi

    if "${BUILD_CMD[@]}"; then
        log_success "镜像构建成功: $FULL_IMAGE_NAME"
    else
        log_error "镜像构建失败"
        exit 1
    fi
}

# 推送镜像
push_image() {
    if [[ "$PUSH" == true ]] && [[ "$MULTI_ARCH" != true ]]; then
        log_info "推送镜像到仓库..."
        
        if docker push "$FULL_IMAGE_NAME"; then
            log_success "镜像推送成功: $FULL_IMAGE_NAME"
            
            if [[ "$TAG" != "latest" ]]; then
                docker push "$IMAGE_NAME:latest"
                log_success "镜像推送成功: $IMAGE_NAME:latest"
            fi
        else
            log_error "镜像推送失败"
            exit 1
        fi
    fi
}

# 显示镜像信息
show_image_info() {
    log_info "镜像信息:"
    
    if [[ "$MULTI_ARCH" != true ]]; then
        docker images "$IMAGE_NAME" --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.CreatedAt}}\t{{.Size}}"
    fi
    
    log_info "构建完成!"
    log_info "使用以下命令运行容器:"
    log_info "  docker run -d -p 80:80 --name wh-frontend $FULL_IMAGE_NAME"
}

# 主函数
main() {
    log_info "WhaleHire 前端项目 Docker 构建脚本"
    log_info "========================================"
    
    parse_args "$@"
    validate_environment
    setup_environment
    setup_build_args
    build_image
    push_image
    show_image_info
}

# 执行主函数
main "$@"