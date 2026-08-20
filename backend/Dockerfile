# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.25 AS build

WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

# Build the server binary (static, linux).
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

# ---- Runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/server /app/server

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/server"]
