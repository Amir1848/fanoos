# Build context is the REPO ROOT (not /fanoos), because the image also needs
# to build the Angular frontend that the Go server serves as static assets:
#   docker build -f fanoos/Dockerfile .

# --- Stage 1: build the Angular frontend -------------------------------
FROM docker.arvancloud.ir/node:26-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# --- Stage 2: build the Go server ---------------------------------------
FROM docker.arvancloud.ir/golang:1.27-alpine AS backend
WORKDIR /app/fanoos
COPY fanoos/go.mod fanoos/go.sum ./
RUN go mod download
COPY fanoos/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/fanoos-web ./cmd/web

# --- Stage 3: runtime ----------------------------------------------------
FROM docker.arvancloud.ir/alpine:3.24
WORKDIR /app
COPY --from=backend /out/fanoos-web ./fanoos-web
# The "application" builder emits the browser bundle under dist/<project>/browser;
# main.go serves everything from ./assets with a SPA fallback to index.html.
COPY --from=frontend /app/frontend/dist/fanoos-admin/browser ./assets
EXPOSE 3000
USER nobody
ENTRYPOINT ["./fanoos-web"]
