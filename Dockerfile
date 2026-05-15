# SmoothNAS plugin: llama.cpp with bearer-auth wrapper.
#
# Builds three variants from one Dockerfile via LLAMA_BASE — CI's
# release workflow renders three tags (cuda / vulkan / cpu) by passing
# the matching upstream image tag.
#
# The wrapper is a tiny Go binary built in a separate stage so the
# final image carries only the upstream layers + one statically-
# linked binary.

# Override LLAMA_BASE in the build args to pick a variant. Docker only
# allows FROM to see ARG values declared before the stage starts.
ARG LLAMA_BASE=ghcr.io/ggml-org/llama.cpp:server

# --- wrapper build ---
FROM golang:1.25-alpine AS wrapper-build
WORKDIR /src
COPY wrapper/go.mod wrapper/main.go ./
# CGO off + static link so the binary runs on any of the upstream
# llama.cpp base images regardless of glibc/musl differences.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /smoothnas-wrapper .

# --- final image ---
# Pinned tags are written into the manifests by the release workflow
# before publishing, so operators always install a known-good
# combination.
FROM ${LLAMA_BASE}

COPY --from=wrapper-build /smoothnas-wrapper /usr/local/bin/smoothnas-wrapper

# llama-server lives at /llama-server in the upstream image; the
# wrapper finds it via $LLAMA_BIN. Override at runtime if upstream
# changes the path.
ENV LLAMA_BIN=/llama-server
ENV LLAMA_PORT=8081
ENV LISTEN_ADDR=:8080

# The container listens on 8080; SmoothNAS' nginx route proxies
# here. Upstream llama-server stays on loopback.
EXPOSE 8080

# The wrapper exec's llama-server itself; passes any extra args
# through. SmoothNAS sets the manifest's container.command, which
# becomes our argv past program name.
ENTRYPOINT ["/usr/local/bin/smoothnas-wrapper"]
