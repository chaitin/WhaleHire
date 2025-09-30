# 多阶段构建 Dockerfile for React + Vite 项目

# 第一阶段：构建阶段
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 package.json 和 package-lock.json（如果存在）
COPY package*.json ./

# 安装依赖（包括开发依赖，构建时需要）
RUN npm ci --silent

# 复制环境变量文件（如果存在）
COPY .env.production* ./

# 复制源代码
COPY . .

# 设置生产环境模式，确保 Vite 读取 .env.production 文件
ENV NODE_ENV=production

# 构建应用
RUN npm run build

# 第二阶段：生产阶段
FROM nginx:alpine AS production

# 安装 curl 用于健康检查
RUN apk add --no-cache curl

# 从构建阶段复制构建产物到 nginx 静态文件目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 创建非 root 用户
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

# 设置正确的权限
RUN chown -R nextjs:nodejs /usr/share/nginx/html && \
    chown -R nextjs:nodejs /var/cache/nginx && \
    chown -R nextjs:nodejs /var/log/nginx && \
    chown -R nextjs:nodejs /etc/nginx/conf.d

# 创建 nginx 运行时需要的目录
RUN touch /var/run/nginx.pid && \
    chown -R nextjs:nodejs /var/run/nginx.pid

# 切换到非 root 用户
USER nextjs

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:80/ || exit 1

# 启动 nginx
CMD ["nginx", "-g", "daemon off;"]