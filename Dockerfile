FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /motson ./cmd/motson

# distroless/static ships CA certificates for HTTPS to the provider
# and TLS to Neon.
FROM gcr.io/distroless/static-debian12
COPY --from=build /motson /motson
ENTRYPOINT ["/motson"]
