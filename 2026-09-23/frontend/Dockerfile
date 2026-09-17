FROM oven/bun:1.3.13 AS builder

WORKDIR /app
COPY . .
RUN bun install
RUN bun run build

FROM nginx:alpine AS runtime

WORKDIR /usr/share/nginx/html
COPY --from=builder /app/dist .
CMD ["nginx", "-g", "daemon off;"]
