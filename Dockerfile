# SmoothNAS plugin: llama.cpp with bearer-auth wrapper.
#
# The final image carries a SmoothNAS auth wrapper plus a llama-server
# built from upstream llama.cpp.

ARG LLAMA_BUILD_BASE=ubuntu:24.04
ARG LLAMA_RUNTIME_BASE=ubuntu:24.04

# --- llama.cpp build ---
FROM ${LLAMA_BUILD_BASE} AS llama-build

ARG LLAMA_BACKEND=cpu
ARG LLAMA_REPO=https://github.com/ggml-org/llama.cpp
ARG LLAMA_REF=latest
ARG LLAMA_ARCHIVE_URL=
ARG CUDA_DOCKER_ARCH=default
ARG BUILD_JOBS=

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN set -eux; \
    if [[ ! -c /dev/null ]]; then \
      rm -f /dev/null; \
      mknod -m 666 /dev/null c 1 3 || touch /dev/null; \
    fi; \
    if [[ ! -c /dev/random ]]; then \
      rm -f /dev/random; \
      mknod -m 666 /dev/random c 1 8; \
    fi; \
    if [[ ! -c /dev/urandom ]]; then \
      rm -f /dev/urandom; \
      mknod -m 666 /dev/urandom c 1 9; \
    fi; \
    chmod 666 /dev/null /dev/random /dev/urandom; \
    mkdir -p /tmp /var/tmp /var/tmp/apt/lists/partial /var/tmp/apt/archives/partial; \
    mkdir -p /var/lib/dpkg/updates /var/lib/dpkg/info /var/lib/dpkg/parts /var/lib/dpkg/triggers; \
    touch /var/lib/dpkg/status /var/lib/dpkg/available; \
    chmod 1777 /tmp /var/tmp; \
    export TMPDIR=/var/tmp; \
    rm -f /etc/apt/apt.conf.d/70debconf; \
    if [[ "${LLAMA_BACKEND}" == "cuda" ]]; then \
      rm -f /etc/apt/sources.list.d/ubuntu.sources; \
      printf '%s\n' \
        'Types: deb' \
        'URIs: http://archive.ubuntu.com/ubuntu/' \
        'Suites: noble noble-updates noble-security' \
        'Components: main universe restricted multiverse' \
        'Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg' \
        > /etc/apt/sources.list.d/ubuntu-main.sources; \
    fi; \
    apt-get -o APT::Sandbox::User=root -o Dir::Temp=/var/tmp -o Acquire::Retries=5 -o Acquire::Queue-Mode=access -o Acquire::http::Pipeline-Depth=0 -o Dir::State::lists=/var/tmp/apt/lists -o Dir::Cache::archives=/var/tmp/apt/archives update && \
    case "${LLAMA_BACKEND}" in \
      cuda) \
        apt-get -o APT::Sandbox::User=root -o Dir::Temp=/var/tmp -o Acquire::Retries=5 -o Acquire::Queue-Mode=access -o Acquire::http::Pipeline-Depth=0 -o Dir::State::lists=/var/tmp/apt/lists -o Dir::Cache::archives=/var/tmp/apt/archives install -y gcc-14 g++-14 binutils build-essential cmake cmake-data curl libgomp1 ca-certificates \
          cuda-compiler-12-8 cuda-nvvm-12-8 cuda-cudart-dev-12-8 cuda-driver-dev-12-8 && \
        curl --retry 5 --retry-delay 2 -fsSLo /var/tmp/cuda-crt-12-8.deb https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2404/x86_64/cuda-crt-12-8_12.8.93-1_amd64.deb && \
        curl --retry 5 --retry-delay 2 -fsSLo /var/tmp/cuda-nvvm-12-8.deb https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2404/x86_64/cuda-nvvm-12-8_12.8.93-1_amd64.deb && \
        curl --retry 5 --retry-delay 2 -fsSLo /var/tmp/cuda-nvcc-12-8.deb https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2404/x86_64/cuda-nvcc-12-8_12.8.93-1_amd64.deb && \
        dpkg -i /var/tmp/cuda-crt-12-8.deb /var/tmp/cuda-nvvm-12-8.deb /var/tmp/cuda-nvcc-12-8.deb ;; \
      vulkan) \
        apt-get -o APT::Sandbox::User=root -o Dir::Temp=/var/tmp -o Acquire::Retries=5 -o Acquire::Queue-Mode=access -o Acquire::http::Pipeline-Depth=0 -o Dir::State::lists=/var/tmp/apt/lists -o Dir::Cache::archives=/var/tmp/apt/archives install -y curl binutils build-essential cmake libssl-dev ca-certificates \
          libxcb-xinput0 libxcb-xinerama0 libxcb-cursor-dev libvulkan-dev glslc spirv-headers ;; \
      cpu) \
        apt-get -o APT::Sandbox::User=root -o Dir::Temp=/var/tmp -o Acquire::Retries=5 -o Acquire::Queue-Mode=access -o Acquire::http::Pipeline-Depth=0 -o Dir::State::lists=/var/tmp/apt/lists -o Dir::Cache::archives=/var/tmp/apt/archives install -y gcc-14 g++-14 binutils build-essential cmake curl libssl-dev ca-certificates ;; \
      *) \
        echo "unsupported LLAMA_BACKEND=${LLAMA_BACKEND}" >&2; exit 1 ;; \
    esac && \
    rm -rf /var/tmp/apt /var/lib/apt/lists/*

RUN if [[ "${LLAMA_BACKEND}" == "cuda" || "${LLAMA_BACKEND}" == "cpu" ]] && command -v gcc-14 >/dev/null && command -v g++-14 >/dev/null; then \
      update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-14 100 && \
      update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-14 100; \
    fi

RUN if [[ "${LLAMA_BACKEND}" == "cuda" ]]; then \
      test -x /usr/local/cuda/bin/nvcc && \
      test -x /usr/local/cuda/nvvm/bin/cicc; \
    fi

RUN set -eux; \
    mkdir -p /src/llama.cpp; \
    archive_url="${LLAMA_ARCHIVE_URL}"; \
    if [[ -z "$archive_url" ]]; then \
      resolved_ref="${LLAMA_REF}"; \
      if [[ "$resolved_ref" == "latest" ]]; then \
        resolved_ref="$(curl -fsSIL -o /dev/null -w '%{url_effective}' "${LLAMA_REPO%.git}/releases/latest" | sed 's#.*/tag/##')"; \
      fi; \
      archive_url="${LLAMA_REPO%.git}/archive/${resolved_ref}.tar.gz"; \
    fi; \
    curl -fsSL -o /var/tmp/llama-cpp.tar.gz "$archive_url"; \
    tar -xzf /var/tmp/llama-cpp.tar.gz --strip-components=1 -C /src/llama.cpp
WORKDIR /src/llama.cpp
RUN set -eux; \
    case "${LLAMA_BACKEND}" in \
      cuda) \
        if [[ "${CUDA_DOCKER_ARCH}" != "default" ]]; then \
          cmake -S . -B build -DCMAKE_BUILD_TYPE=Release -DCMAKE_MODULE_PATH=/usr/share/cmake-3.28/Modules -DCUDAToolkit_ROOT=/usr/local/cuda -DCMAKE_CUDA_COMPILER=/usr/local/cuda/bin/nvcc -DGGML_NATIVE=OFF -DGGML_BACKEND_DL=ON -DGGML_CPU_ALL_VARIANTS=ON -DLLAMA_BUILD_TESTS=OFF -DGGML_CUDA=ON -DCMAKE_CUDA_ARCHITECTURES="${CUDA_DOCKER_ARCH}" -DCMAKE_EXE_LINKER_FLAGS=-Wl,--allow-shlib-undefined; \
        else \
          cmake -S . -B build -DCMAKE_BUILD_TYPE=Release -DCMAKE_MODULE_PATH=/usr/share/cmake-3.28/Modules -DCUDAToolkit_ROOT=/usr/local/cuda -DCMAKE_CUDA_COMPILER=/usr/local/cuda/bin/nvcc -DGGML_NATIVE=OFF -DGGML_BACKEND_DL=ON -DGGML_CPU_ALL_VARIANTS=ON -DLLAMA_BUILD_TESTS=OFF -DGGML_CUDA=ON -DCMAKE_EXE_LINKER_FLAGS=-Wl,--allow-shlib-undefined; \
        fi; \
        ;; \
      vulkan) \
        cmake -S . -B build -DCMAKE_BUILD_TYPE=Release -DGGML_NATIVE=OFF -DGGML_BACKEND_DL=ON -DGGML_CPU_ALL_VARIANTS=ON -DLLAMA_BUILD_TESTS=OFF -DGGML_VULKAN=ON; \
        ;; \
      cpu) \
        cmake -S . -B build -DCMAKE_BUILD_TYPE=Release -DGGML_NATIVE=OFF -DGGML_BACKEND_DL=ON -DGGML_CPU_ALL_VARIANTS=ON -DLLAMA_BUILD_TESTS=OFF; \
        ;; \
    esac; \
    jobs="${BUILD_JOBS:-$(nproc)}"; \
    echo "using build parallelism: ${jobs}"; \
    cmake --build build --config Release --target llama-server --parallel "$jobs"; \
    case "${LLAMA_BACKEND}" in \
      cuda) cmake --build build --config Release --target ggml-cuda --parallel "$jobs" ;; \
      vulkan) cmake --build build --config Release --target ggml-vulkan --parallel "$jobs" ;; \
    esac

RUN mkdir -p /out/lib /out/bin && \
    find build -name "*.so*" -exec cp -P {} /out/lib \; && \
    cp build/bin/llama-server /out/bin/llama-server

# --- final image ---
FROM ${LLAMA_RUNTIME_BASE}

ARG LLAMA_BACKEND=cpu

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN set -eux; \
    if [[ ! -c /dev/null ]]; then \
      rm -f /dev/null; \
      mknod -m 666 /dev/null c 1 3 || touch /dev/null; \
    fi; \
    if [[ ! -c /dev/random ]]; then \
      rm -f /dev/random; \
      mknod -m 666 /dev/random c 1 8; \
    fi; \
    if [[ ! -c /dev/urandom ]]; then \
      rm -f /dev/urandom; \
      mknod -m 666 /dev/urandom c 1 9; \
    fi; \
    chmod 666 /dev/null /dev/random /dev/urandom; \
    mkdir -p /tmp /var/tmp /var/tmp/apt/lists/partial /var/tmp/apt/archives/partial; \
    mkdir -p /var/lib/dpkg/updates /var/lib/dpkg/info /var/lib/dpkg/parts /var/lib/dpkg/triggers; \
    touch /var/lib/dpkg/status /var/lib/dpkg/available; \
    chmod 1777 /tmp /var/tmp; \
    export TMPDIR=/var/tmp; \
    rm -f /etc/apt/apt.conf.d/70debconf; \
    if [[ "${LLAMA_BACKEND}" == "cuda" ]]; then \
      rm -f /etc/apt/sources.list.d/ubuntu.sources; \
      printf '%s\n' \
        'Types: deb' \
        'URIs: http://archive.ubuntu.com/ubuntu/' \
        'Suites: noble' \
        'Components: main universe restricted multiverse' \
        'Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg' \
        > /etc/apt/sources.list.d/ubuntu-main.sources; \
    fi; \
    apt-get -o APT::Sandbox::User=root -o Dir::Temp=/var/tmp -o Acquire::Retries=5 -o Acquire::Queue-Mode=access -o Acquire::http::Pipeline-Depth=0 -o Dir::State::lists=/var/tmp/apt/lists -o Dir::Cache::archives=/var/tmp/apt/archives update && \
    case "${LLAMA_BACKEND}" in \
      cuda) \
        apt-get -o APT::Sandbox::User=root -o Dir::Temp=/var/tmp -o Acquire::Retries=5 -o Acquire::Queue-Mode=access -o Acquire::http::Pipeline-Depth=0 -o Dir::State::lists=/var/tmp/apt/lists -o Dir::Cache::archives=/var/tmp/apt/archives install -y libgomp1 curl ca-certificates ;; \
      vulkan) \
        apt-get -o APT::Sandbox::User=root -o Dir::Temp=/var/tmp -o Acquire::Retries=5 -o Acquire::Queue-Mode=access -o Acquire::http::Pipeline-Depth=0 -o Dir::State::lists=/var/tmp/apt/lists -o Dir::Cache::archives=/var/tmp/apt/archives install -y libgomp1 curl ca-certificates libvulkan1 mesa-vulkan-drivers \
          libglvnd0 libgl1 libglx0 libegl1 libgles2 ;; \
      cpu) \
        apt-get -o APT::Sandbox::User=root -o Dir::Temp=/var/tmp -o Acquire::Retries=5 -o Acquire::Queue-Mode=access -o Acquire::http::Pipeline-Depth=0 -o Dir::State::lists=/var/tmp/apt/lists -o Dir::Cache::archives=/var/tmp/apt/archives install -y libgomp1 curl ca-certificates ;; \
      *) \
        echo "unsupported LLAMA_BACKEND=${LLAMA_BACKEND}" >&2; exit 1 ;; \
    esac && \
    apt-get clean -y

COPY --from=llama-build /out/lib/ /app/
COPY --from=llama-build /out/bin/llama-server /app/llama-server
COPY smoothnas-wrapper /usr/local/bin/smoothnas-wrapper

ENV LLAMA_BIN=/app/llama-server
ENV LLAMA_PORT=8081
ENV LISTEN_ADDR=:8080

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/smoothnas-wrapper"]
