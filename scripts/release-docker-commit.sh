#!/usr/bin/env bash
set -euo pipefail

export PATH="/usr/local/bin:/usr/bin:/bin:${PATH:-}"

ref="${REF:?REF is required}"
backend="${LLAMA_BACKEND:?LLAMA_BACKEND is required}"
build_base="${LLAMA_BUILD_BASE:?LLAMA_BUILD_BASE is required}"
runtime_base="${LLAMA_RUNTIME_BASE:?LLAMA_RUNTIME_BASE is required}"
llama_repo="${LLAMA_REPO:-https://github.com/AtomicBot-ai/atomic-llama-cpp-turboquant.git}"
llama_ref="${LLAMA_REF:-24cabf4d08d460cfb6e73fa308a15b34e2b04600}"
llama_archive_url="${LLAMA_ARCHIVE_URL:-${llama_repo%.git}/archive/${llama_ref}.tar.gz}"
llama_source_dir="${LLAMA_SOURCE_DIR:-}"
cuda_arch="${CUDA_DOCKER_ARCH:-default}"
build_jobs="${BUILD_JOBS:-}"
image_build_mode="${LLAMA_IMAGE_BUILD_MODE:-crane}"
require_preinstalled_deps="${LLAMA_REQUIRE_PREINSTALLED_DEPS:-false}"

suffix="${GITHUB_RUN_ID:-manual}-${backend}-$$"
tmp="$(mktemp -d)"
src="$tmp/src"
rootfs="$tmp/rootfs"

cleanup() {
  rm -rf "$tmp"
}
trap cleanup EXIT

case "$backend" in
  cuda)
    build_deps="gcc-14 g++-14 binutils build-essential cmake git libssl-dev libgomp1 ca-certificates cuda-nvcc-12-8 cuda-nvvm-12-8 cuda-cudart-dev-12-8 cuda-driver-dev-12-8"
    runtime_deps="libgomp1 curl ca-certificates"
    ;;
  vulkan)
    build_deps="gcc-14 g++-14 binutils git build-essential cmake libssl-dev ca-certificates libxcb-xinput0 libxcb-xinerama0 libxcb-cursor-dev libvulkan-dev glslc spirv-headers"
    runtime_deps="libgomp1 curl ca-certificates libvulkan1 mesa-vulkan-drivers libglvnd0 libgl1 libglx0 libegl1 libgles2"
    ;;
  cpu)
    build_deps="binutils build-essential cmake git libssl-dev ca-certificates"
    runtime_deps="libgomp1 curl ca-certificates"
    ;;
  *)
    echo "unsupported LLAMA_BACKEND=$backend" >&2
    exit 1
    ;;
esac

install_runner_deps() {
  export DEBIAN_FRONTEND=noninteractive
  if packages_installed $build_deps $runtime_deps; then
    echo "build dependencies already installed; skipping apt-get"
    return
  fi
  if [[ "$require_preinstalled_deps" == "true" ]]; then
    echo "required build dependencies are missing from the runner image; refusing runtime package install" >&2
    return 1
  fi
  if [[ "$backend" == "cuda" ]]; then
    ensure_cuda_apt_repo
  fi
  apt_get update
  apt_get install -y $build_deps $runtime_deps
  apt_get clean -y
}

ensure_cuda_apt_repo() {
  if [[ -f /etc/apt/sources.list.d/cuda-ubuntu2404-x86_64.list ]]; then
    return
  fi

  apt_get update
  apt_get install -y ca-certificates curl
  local source_list="$tmp/cuda-ubuntu2404-x86_64.list"
  printf '%s\n' 'deb [trusted=yes] https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2404/x86_64/ /' > "$source_list"
  if [[ "$(id -u)" -eq 0 ]]; then
    install -D -m 0644 "$source_list" /etc/apt/sources.list.d/cuda-ubuntu2404-x86_64.list
  elif command -v sudo >/dev/null 2>&1; then
    sudo install -D -m 0644 "$source_list" /etc/apt/sources.list.d/cuda-ubuntu2404-x86_64.list
  else
    echo "installing CUDA apt source requires root or sudo" >&2
    return 1
  fi
}

parallel_jobs() {
  local quota period jobs
  if [[ -r /sys/fs/cgroup/cpu.max ]]; then
    read -r quota period < /sys/fs/cgroup/cpu.max || true
    if [[ "${quota:-max}" != "max" && "${period:-0}" -gt 0 ]]; then
      jobs=$(((quota + period - 1) / period))
      if ((jobs > 0)); then
        echo "$jobs"
        return
      fi
    fi
  fi
  nproc
}

ensure_cxx_toolchain() {
  local cc cxx cc1plus cc1plus_dir major

  cc="$(command -v gcc-14 || command -v gcc || true)"
  cxx="$(command -v g++-14 || command -v g++ || true)"
  if [[ -z "$cc" || -z "$cxx" ]]; then
    if [[ "$require_preinstalled_deps" == "true" ]]; then
      echo "gcc/g++ are missing from the runner image; refusing runtime package install" >&2
      return 1
    fi
    apt_get update
    apt_get install -y build-essential g++
    cc="$(command -v gcc-14 || command -v gcc || true)"
    cxx="$(command -v g++-14 || command -v g++ || true)"
  fi
  if [[ -z "$cc" || -z "$cxx" ]]; then
    echo "gcc/g++ are not available after installing build dependencies" >&2
    return 1
  fi

  cc1plus="$("$cxx" -print-prog-name=cc1plus)"
  if [[ "$cc1plus" == "cc1plus" || ! -x "$cc1plus" ]]; then
    for cc1plus_dir in \
      /usr/local/lib/gcc/x86_64-linux-gnu/14 \
      /usr/libexec/gcc/x86_64-linux-gnu/14 \
      /usr/lib/gcc/x86_64-linux-gnu/14; do
      if [[ -x "$cc1plus_dir/cc1plus" ]]; then
        export COMPILER_PATH="${cc1plus_dir}${COMPILER_PATH:+:${COMPILER_PATH}}"
        cc1plus="$("$cxx" -print-prog-name=cc1plus)"
        break
      fi
    done
  fi
  if [[ "$cc1plus" == "cc1plus" || ! -x "$cc1plus" ]]; then
    if [[ "$require_preinstalled_deps" == "true" ]]; then
      echo "cc1plus is missing from the runner image; refusing runtime package install" >&2
      return 1
    fi
    major="$("$cxx" -dumpversion | cut -d. -f1)"
    apt_get update
    apt_get install --reinstall -y "g++" "g++-${major}" "g++-${major}-x86-64-linux-gnu" || apt_get install --reinstall -y g++
    cc1plus="$("$cxx" -print-prog-name=cc1plus)"
  fi
  if [[ "$cc1plus" == "cc1plus" || ! -x "$cc1plus" ]]; then
    echo "cc1plus is not available to $cxx" >&2
    "$cxx" -v >&2 || true
    return 1
  fi

  cc1plus_dir="$(dirname "$cc1plus")"
  export CC="$cc"
  export CXX="$cxx"
  export COMPILER_PATH="${cc1plus_dir}${COMPILER_PATH:+:${COMPILER_PATH}}"
  printf 'int main() { return 0; }\n' > "$tmp/cxx-sanity.cpp"
  "$CXX" "$tmp/cxx-sanity.cpp" -o "$tmp/cxx-sanity"
}

apt_get() {
  if [[ "$(id -u)" -eq 0 ]]; then
    apt-get "$@"
  elif command -v sudo >/dev/null 2>&1; then
    sudo apt-get "$@"
  else
    echo "apt-get requires root or sudo" >&2
    return 1
  fi
}

packages_installed() {
  local package
  for package in "$@"; do
    dpkg-query -W -f='${Status}' "$package" 2>/dev/null | grep -q 'install ok installed' || return 1
  done
}

resolve_llama_source_dir() {
  local candidate
  if [[ -n "$llama_source_dir" ]]; then
    if [[ -f "$llama_source_dir/CMakeLists.txt" ]]; then
      printf '%s\n' "$llama_source_dir"
      return
    fi
    echo "LLAMA_SOURCE_DIR=$llama_source_dir does not look like a llama.cpp source tree" >&2
    return 1
  fi

  for candidate in \
    /opt/atomic-llama-cpp-turboquant \
    /opt/smoothnas/atomic-llama-cpp-turboquant \
    /opt/llama.cpp \
    /opt/smoothnas/llama.cpp; do
    if [[ -f "$candidate/CMakeLists.txt" ]]; then
      printf '%s\n' "$candidate"
      return
    fi
  done

  echo "no baked llama.cpp source tree found; set LLAMA_SOURCE_DIR on the runner image" >&2
  return 1
}

stage_llama_source() {
  mkdir -p "$src"

  if [[ "$require_preinstalled_deps" == "true" ]]; then
    local baked_source
    baked_source="$(resolve_llama_source_dir)"
    echo "using baked llama.cpp source: $baked_source"
    if [[ -d "$baked_source/.git" ]] && command -v git >/dev/null 2>&1; then
      local baked_ref
      baked_ref="$(git -C "$baked_source" rev-parse HEAD)"
      if [[ "$baked_ref" != "$llama_ref" ]]; then
        echo "baked llama.cpp source is at $baked_ref, expected $llama_ref" >&2
        return 1
      fi
    elif [[ -f "$baked_source/.smoothnas-llama-ref" ]]; then
      local baked_ref
      baked_ref="$(tr -d '[:space:]' < "$baked_source/.smoothnas-llama-ref")"
      if [[ "$baked_ref" != "$llama_ref" ]]; then
        echo "baked llama.cpp source marker is $baked_ref, expected $llama_ref" >&2
        return 1
      fi
    else
      echo "baked llama.cpp source must include .git metadata or .smoothnas-llama-ref" >&2
      return 1
    fi
    cp -a "$baked_source" "$src/llama.cpp"
    rm -rf "$src/llama.cpp/build"
    return
  fi

  git clone --depth 1 "$llama_repo" "$src/llama.cpp"
  (
    cd "$src/llama.cpp"
    git fetch --depth 1 origin "$llama_ref"
    git checkout --detach "$llama_ref"
  )
}

build_llama() {
  stage_llama_source
  (
    cd "$src/llama.cpp"

    cmake_args=(
      -S .
      -B build
      -DCMAKE_BUILD_TYPE=Release
      -DGGML_NATIVE=OFF
      -DGGML_BACKEND_DL=ON
      -DLLAMA_BUILD_TESTS=OFF
    )
    if [[ "$backend" == "cpu" ]]; then
      cmake_args+=(-DGGML_CPU_ALL_VARIANTS=ON)
    else
      cmake_args+=(-DGGML_CPU_ALL_VARIANTS=OFF)
    fi
    case "$backend" in
      cuda)
        if [[ "$cuda_arch" != "default" ]]; then
          cmake_args+=("-DCMAKE_CUDA_ARCHITECTURES=$cuda_arch")
        fi
        cmake_args+=(
          -DGGML_CUDA=ON
          -DCUDAToolkit_ROOT=/usr/local/cuda
          -DCMAKE_CUDA_COMPILER=/usr/local/cuda/bin/nvcc
          -DCMAKE_EXE_LINKER_FLAGS=-Wl,--allow-shlib-undefined
        )
        ;;
      vulkan)
        cmake_args+=(-DGGML_VULKAN=ON)
        ;;
    esac

    cmake "${cmake_args[@]}"
    cmake --build build --config Release --target llama-server -j"$build_jobs"
    case "$backend" in
      cuda)
        cmake --build build --config Release --target ggml-cuda -j"$build_jobs"
        ;;
      vulkan)
        cmake --build build --config Release --target ggml-vulkan -j"$build_jobs"
        ;;
    esac
  )
}

copy_one_lib() {
  local lib="$1"
  [[ -f "$lib" ]] || return 0
  local dest="$rootfs$(dirname "$lib")"
  mkdir -p "$dest"
  cp -L "$lib" "$dest/"
}

copy_ldd_deps() {
  local target="$1"
  local deps
  deps="$(ldd "$target" 2>/dev/null || true)"
  awk '
    /=> \// { print $3 }
    /^\// { print $1 }
  ' <<<"$deps" | sort -u | while read -r lib; do
    copy_one_lib "$lib"
  done
}

copy_runtime_closure() {
  local before after
  for _ in 1 2 3 4; do
    before="$(find "$rootfs" -type f | wc -l)"
    while IFS= read -r file; do
      copy_ldd_deps "$file"
    done < <(find "$rootfs" -type f \( -perm -111 -o -name '*.so' -o -name '*.so.*' \))
    after="$(find "$rootfs" -type f | wc -l)"
    [[ "$before" == "$after" ]] && break
  done
}

stage_rootfs() {
  mkdir -p "$rootfs/app" "$rootfs/usr/local/bin"
  find "$src/llama.cpp/build" -name '*.so*' -exec cp -P {} "$rootfs/app/" \;
  cp "$src/llama.cpp/build/bin/llama-server" "$rootfs/app/llama-server"
  cp smoothnas-wrapper "$rootfs/usr/local/bin/smoothnas-wrapper"
  if [[ -d /etc/ssl/certs ]]; then
    mkdir -p "$rootfs/etc/ssl"
    cp -a /etc/ssl/certs "$rootfs/etc/ssl/"
  fi
  copy_one_lib /lib64/ld-linux-x86-64.so.2

  copy_runtime_closure

  if [[ "$backend" == "vulkan" ]]; then
    if [[ -d /usr/share/vulkan ]]; then
      mkdir -p "$rootfs/usr/share"
      cp -a /usr/share/vulkan "$rootfs/usr/share/"
    fi
    shopt -s nullglob
    local vulkan_libs=(/usr/lib/x86_64-linux-gnu/libvulkan_*.so /usr/lib/x86_64-linux-gnu/libVkLayer_*.so)
    if ((${#vulkan_libs[@]})); then
      mkdir -p "$rootfs/usr/lib/x86_64-linux-gnu"
      cp -a "${vulkan_libs[@]}" "$rootfs/usr/lib/x86_64-linux-gnu/"
    fi
    shopt -u nullglob
    if [[ -d /usr/lib/x86_64-linux-gnu/dri ]]; then
      mkdir -p "$rootfs/usr/lib/x86_64-linux-gnu"
      cp -a /usr/lib/x86_64-linux-gnu/dri "$rootfs/usr/lib/x86_64-linux-gnu/"
    fi
    copy_runtime_closure
  fi
}

commit_image_with_crane() {
  local layer="$tmp/rootfs.tar"

  command -v crane >/dev/null 2>&1 || {
    echo "crane is required to package and push the image" >&2
    return 1
  }

  tar \
    --sort=name \
    --mtime='UTC 1970-01-01' \
    --owner=0 \
    --group=0 \
    --numeric-owner \
    -cf "$layer" \
    -C "$rootfs" \
    .

  crane append \
    --oci-empty-base \
    --new_layer "$layer" \
    --new_tag "$ref"

  crane mutate "$ref" \
    --tag "$ref" \
    --set-platform linux/amd64 \
    --entrypoint /usr/local/bin/smoothnas-wrapper \
    --env LLAMA_BIN=/app/llama-server \
    --env LLAMA_PORT=8081 \
    --env LISTEN_ADDR=:8080 \
    --env LD_LIBRARY_PATH=/app:/usr/local/cuda/lib64:/usr/local/lib:/usr/lib/x86_64-linux-gnu \
    --exposed-ports 8080
}

commit_image_with_docker() {
  cat > "$tmp/Dockerfile" <<'EOF'
FROM scratch
COPY rootfs/. /
ENV LLAMA_BIN=/app/llama-server
ENV LLAMA_PORT=8081
ENV LISTEN_ADDR=:8080
ENV LD_LIBRARY_PATH=/app:/usr/local/cuda/lib64:/usr/local/lib:/usr/lib/x86_64-linux-gnu
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/smoothnas-wrapper"]
EOF
  docker build -f "$tmp/Dockerfile" -t "$ref" "$tmp"
}

build_docker_image() {
  docker build \
    --build-arg LLAMA_BACKEND="$backend" \
    --build-arg LLAMA_BUILD_BASE="$build_base" \
    --build-arg LLAMA_RUNTIME_BASE="$runtime_base" \
    --build-arg LLAMA_REPO="$llama_repo" \
    --build-arg LLAMA_REF="$llama_ref" \
    --build-arg LLAMA_ARCHIVE_URL="$llama_archive_url" \
    --build-arg CUDA_DOCKER_ARCH="$cuda_arch" \
    --build-arg BUILD_JOBS="$build_jobs" \
    -t "$ref" \
    .
}

if [[ -z "$build_jobs" ]]; then
  build_jobs="$(parallel_jobs)"
fi
echo "using build parallelism: $build_jobs"

case "$image_build_mode" in
  crane|docker|native) ;;
  *)
    echo "unsupported LLAMA_IMAGE_BUILD_MODE=$image_build_mode" >&2
    exit 1
    ;;
esac

if [[ "$image_build_mode" == "docker" && ( "$backend" == "cuda" || "$backend" == "vulkan" ) ]]; then
  build_docker_image
  exit 0
fi

install_runner_deps
ensure_cxx_toolchain
build_llama
stage_rootfs
case "$image_build_mode" in
  crane|native)
    commit_image_with_crane
    ;;
  docker)
    commit_image_with_docker
    ;;
esac
