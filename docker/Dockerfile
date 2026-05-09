FROM node:22-alpine AS web-builder
WORKDIR /web
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm build

FROM golang:1.26-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -o /api ./cmd/api/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=go-builder /api /usr/local/bin/api
COPY --from=web-builder /web/dist /web/dist
WORKDIR /
EXPOSE 8080
ENTRYPOINT ["api"]
