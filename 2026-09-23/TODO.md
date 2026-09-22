# TODO

A little list of things to do for anyone looking to do this demo on their own time

## 1. Write a Dockerfile for the backend

Note the backend uses `go 1.27.1` (this can be identified in the `go.mod`)

A basic Dockerfile may look like the following
```Dockerfile
FROM golang:1.27.1-trixie AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./gen gen/
COPY ./cmd cmd/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o chat_backend \
    ./cmd/server

CMD [ "/app/chat_backend" ]
```

However, this copies all of the build dependencies and code to the distributed version of the application. We can instead distribute only the binary by using a *multi-stage build*

```diff
FROM golang:1.27.1-trixie AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./gen gen/
COPY ./cmd cmd/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o chat_backend \
    ./cmd/server
+
+ FROM alpine:3.24 AS server
+ 
+ WORKDIR /app
+ 
+ COPY --from=builder /app/chat_backend .
+
CMD [ "/app/chat_backend" ]
```

Only the `server` stage gets sent to registries (and thus pulled by our servers), which for larger applications can save on storage costs and deploy times.

## 2. Add the backend to the docker-compose file

The `docker-compose.yml` already has a number of services. Note the order of dependencies with regards to our backend:
- The backend needs the database
- The backend also needs the database to be setup (so the `migrate` container must have finished succesfully)
- The `proxy` (more on reverse proxy's in another meeting) shouldn't start until the backend is running

The compose file can be finished with the following lines:

```diff
services:
  db:
    image: postgres:18-alpine
    environment:
      POSTGRES_USER: chat
      POSTGRES_PASSWORD: chat
      POSTGRES_DB: chat
    volumes:
      - db_data:/var/lib/postgresql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U chat -d chat"]
      interval: 5s
      timeout: 5s
      retries: 5

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "8080:80"

  migrate:
    build:
      context: ./backend
      dockerfile: migrate.Dockerfile
    environment:
      DATABASE_URL: "postgres://chat:chat@db:5432/chat"
    depends_on:
      db:
        condition: service_healthy
    restart: on-failure

+ backend:
+   build:
+     context: ./backend
+     dockerfile: Dockerfile
+   environment:
+     PORT: "8080"
+     DATABASE_URL: "postgres://chat:chat@db:5432/chat"
+   ports:
+     - "8081:8080"
+   depends_on:
+     db:
+       condition: service_healthy
+     migrate:
+       condition: service_completed_successfully

  proxy:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
+     - backend
      - frontend

volumes: # use volumes for data persistence, remember that writable upper-dir in OverlayFS does not persist if your container is deleted
  db_data:
```

Now you should be able to run `docker compose up`, and the service will be accessible in your browser under `localhost/`
