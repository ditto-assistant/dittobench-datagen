# Build the generate service from this repo's own packages — no external
# module dependencies, so the image is byte-reproducible from a tag.
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/generate-service ./cmd/generate-service

# Static binary, no shell: the service has no exec/filesystem needs and runs
# as nonroot. Cloud Run injects PORT; main.go reads it (default 8090).
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/generate-service /generate-service
ENTRYPOINT ["/generate-service"]
