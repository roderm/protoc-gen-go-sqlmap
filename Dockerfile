# Image for the BSR remote plugin (buf.build/snaerverk/go-sqlmap).
# Built and pushed by `buf beta registry plugin push`, which requires
# linux/amd64, a minimal base and a non-root user.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

WORKDIR /src
# Copied first so the module cache survives source-only changes.
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY pkg/ pkg/

# Static, stripped, and reproducible: the plugin runs in a scratch image with
# no libc and no build metadata worth carrying.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/protoc-gen-go-sqlmap ./cmd/protoc-gen-go-sqlmap

FROM scratch
COPY --from=build /out/protoc-gen-go-sqlmap /protoc-gen-go-sqlmap
USER 65532:65532
ENTRYPOINT ["/protoc-gen-go-sqlmap"]
