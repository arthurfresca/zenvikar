# Stage 1: Install dependencies
FROM node:20-alpine AS deps

RUN apk add --no-cache libc6-compat
RUN corepack enable && corepack prepare pnpm@9.15.0 --activate

WORKDIR /app

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml* .npmrc ./
COPY apps/tenant-web/package.json ./apps/tenant-web/
COPY packages/types/package.json ./packages/types/
COPY packages/config/package.json ./packages/config/
COPY packages/ui/package.json ./packages/ui/

RUN pnpm install --frozen-lockfile

# Stage 2: Build
FROM node:20-alpine AS builder

RUN corepack enable && corepack prepare pnpm@9.15.0 --activate

WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules

COPY packages/ ./packages/
COPY apps/tenant-web/ ./apps/tenant-web/
COPY package.json pnpm-workspace.yaml tsconfig.json .npmrc ./

ENV NEXT_TELEMETRY_DISABLED=1

RUN pnpm --filter @zenvikar/tenant-web build

# Stage 3: Run
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/apps/tenant-web/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/apps/tenant-web/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/apps/tenant-web/.next/static ./apps/tenant-web/.next/static

USER nextjs

EXPOSE 3002
ENV PORT=3002
ENV HOSTNAME="0.0.0.0"

CMD ["node", "apps/tenant-web/server.js"]
