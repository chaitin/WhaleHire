# ==============================
# Stage 1: Install all dependencies
# ==============================
FROM node:20-alpine AS deps
WORKDIR /app
RUN apk add --no-cache libc6-compat
COPY package.json package-lock.json* ./
RUN npm ci

# ==============================
# Stage 2: Build Next.js
# ==============================
FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# 添加构建参数
ARG BACKEND_HOST=http://whalehire-backend:8888

# 构建应用，并删除缓存
RUN npm run build

# ==============================
# Stage 3: Final runtime (Standalone mode)
# ==============================
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000

# 创建非 root 用户
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs

# 复制 standalone 构建输出
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/public ./public

# 使用非 root 用户运行
USER nextjs

EXPOSE 3000
CMD ["node", "server.js"]
