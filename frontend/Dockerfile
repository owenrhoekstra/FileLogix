FROM node:22-alpine AS builder
WORKDIR /app

ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

FROM alpine:latest
COPY --from=builder /app/dist /app/dist
CMD ["cp", "-r", "/app/dist/.", "/app/dist-volume/"]