#!/usr/bin/env bash
set -euo pipefail

ref="${REF:?REF is required}"
backend="${LLAMA_BACKEND:?LLAMA_BACKEND is required}"
build_base="${LLAMA_BUILD_BASE:?LLAMA_BUILD_BASE is required}"
runtime_base="${LLAMA_RUNTIME_BASE:?LLAMA_RUNTIME_BASE is required}"
llama_repo="${LLAMA_REPO:-https://github.com/AtomicBot-ai/atomic-llama-cpp-turboquant.git}"
llama_ref="${LLAMA_REF:-0a635dcd92ba66c75fccfef91c3e106f4668f367}"
cuda_arch="${CUDA_DOCKER_ARCH:-default}"

suffix="${GITHUB_RUN_ID:-manual}-${backend}-$$"
builder="llama-build-${suffix}"
runtime="llama-runtime-${suffix}"
tmp="$(mktemp -d)"

cleanup() {
  docker rm -f "$builder" "$runtime" >/dev/null 2>&1 || true
  rm -rf "$tmp"
}
trap cleanup EXIT

run_container() {
  local name="$1"
  docker start "$name" >/dev/null
  local status
  status="$(docker wait "$name")"
  docker logs "$name"
  if [[ "$status" != "0" ]]; then
    echo "$name exited with status $status" >&2
    exit "$status"
  fi
}

case "$backend" in
  cuda)
    build_deps="gcc-14 g++-14 build-essential cmake git libssl-dev libgomp1 ca-certificates"
    runtime_deps="libgomp1 curl ca-certificates"
    ;;
  vulkan)
    build_deps="git build-essential cmake libssl-dev ca-certificates libxcb-xinput0 libxcb-xinerama0 libxcb-cursor-dev libvulkan-dev glslc spirv-headers"
    runtime_deps="libgomp1 curl ca-certificates libvulkan1 mesa-vulkan-drivers libglvnd0 libgl1 libglx0 libegl1 libgles2"
    ;;
  cpu)
    build_deps="gcc-14 g++-14 build-essential cmake git libssl-dev ca-certificates"
    runtime_deps="libgomp1 curl ca-certificates"
    ;;
  *)
    echo "unsupported LLAMA_BACKEND=$backend" >&2
    exit 1
    ;;
esac

docker create \
  --name "$builder" \
  -e "DEBIAN_FRONTEND=noninteractive" \
  -e "LLAMA_BACKEND=$backend" \
  -e "LLAMA_REPO=$llama_repo" \
  -e "LLAMA_REF=$llama_ref" \
  -e "CUDA_DOCKER_ARCH=$cuda_arch" \
  "$build_base" \
  bash -lc "
    set -euxo pipefail
    apt-get update
    apt-get install -y $build_deps
    if [[ \"\$LLAMA_BACKEND\" == \"cuda\" || \"\$LLAMA_BACKEND\" == \"cpu\" ]]; then
      update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-14 100
      update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-14 100
    fi
    mkdir -p /src
    cd /src
    git clone --depth 1 \"\$LLAMA_REPO\" llama.cpp
    cd llama.cpp
    git fetch --depth 1 origin \"\$LLAMA_REF\"
    git checkout --detach \"\$LLAMA_REF\"
    cmake_args=(
      -S .
      -B build
      -DCMAKE_BUILD_TYPE=Release
      -DGGML_NATIVE=OFF
      -DGGML_BACKEND_DL=ON
      -DGGML_CPU_ALL_VARIANTS=ON
      -DLLAMA_BUILD_TESTS=OFF
    )
    case \"\$LLAMA_BACKEND\" in
      cuda)
        if [[ \"\$CUDA_DOCKER_ARCH\" != \"default\" ]]; then
          cmake_args+=(\"-DCMAKE_CUDA_ARCHITECTURES=\$CUDA_DOCKER_ARCH\")
        fi
        cmake_args+=(-DGGML_CUDA=ON -DCMAKE_EXE_LINKER_FLAGS=-Wl,--allow-shlib-undefined)
        ;;
      vulkan)
        cmake_args+=(-DGGML_VULKAN=ON)
        ;;
    esac
    cmake \"\${cmake_args[@]}\"
    cmake --build build --config Release -j\"\$(nproc)\"
    mkdir -p /out/lib /out/bin
    find build -name '*.so*' -exec cp -P {} /out/lib \;
    cp build/bin/llama-server /out/bin/llama-server
  "
run_container "$builder"
docker cp "$builder:/out" "$tmp/out"

docker create \
  --name "$runtime" \
  -e "DEBIAN_FRONTEND=noninteractive" \
  "$runtime_base" \
  bash -lc "
    set -euxo pipefail
    apt-get update
    apt-get install -y $runtime_deps
    apt-get autoremove -y
    apt-get clean -y
    mkdir -p /app /usr/local/bin
    rm -rf /tmp/* /var/tmp/* /var/lib/apt/lists/*
  "
run_container "$runtime"

docker cp "$tmp/out/lib/." "$runtime:/app/"
docker cp "$tmp/out/bin/llama-server" "$runtime:/app/llama-server"
docker cp smoothnas-wrapper "$runtime:/usr/local/bin/smoothnas-wrapper"

docker commit \
  --change 'ENV LLAMA_BIN=/app/llama-server' \
  --change 'ENV LLAMA_PORT=8081' \
  --change 'ENV LISTEN_ADDR=:8080' \
  --change 'EXPOSE 8080' \
  --change 'ENTRYPOINT ["/usr/local/bin/smoothnas-wrapper"]' \
  "$runtime" \
  "$ref"
