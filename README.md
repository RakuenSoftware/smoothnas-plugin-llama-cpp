# smoothnas-plugin-llama-cpp

The reference SmoothNAS plugin for [llama.cpp](https://github.com/ggml-org/llama.cpp). Runs `llama-server` inside a SmoothNAS-managed LXC system container with GPU passthrough, tier-bound model storage, and bearer-injected auth from the SmoothNAS UI.

This is the first reference plugin built against the [SmoothNAS plugin system](https://github.com/RakuenSoftware/smoothnas/blob/main/docs/proposals/pending/smoothnas-plugins.md). It exists as a worked example for plugin authors *and* as a real workload SmoothNAS operators can install.

## Variants

Three published variants, one per accelerator:

| Manifest | Image tag | Profile | Use when |
|----------|-----------|---------|----------|
| `smoothnas-plugin.yaml`        | `ghcr.io/rakuensoftware/smoothnas-plugin-llama-cpp:VER-cuda`   | `gpu-nvidia` | Host has an NVIDIA GPU |
| `smoothnas-plugin-vulkan.yaml` | `ghcr.io/rakuensoftware/smoothnas-plugin-llama-cpp:VER-vulkan` | `gpu-amd`    | Host has an AMD GPU. Uses Vulkan via `/dev/dri` — no ROCm runtime needed. |
| `smoothnas-plugin-cpu.yaml`    | `ghcr.io/rakuensoftware/smoothnas-plugin-llama-cpp:VER-cpu`    | none         | No GPU, or quick experiments before paying the GPU pull cost |

Pick the matching manifest at install time in the SmoothNAS UI's plugin install wizard, or via `tierd-cli plugin install --tier <pool> <manifest.yaml>`.

## Why a wrapper image instead of pointing at upstream directly?

Upstream `llama-server` has no built-in HTTP auth. SmoothNAS phase-7 injects a bearer token at the nginx layer, but the plugin still has to *check* it — otherwise an operator who exposes the SmoothNAS UI publicly would also be exposing an unauthenticated inference server.

This repo's `wrapper/` is a tiny Go binary (~150 LoC) that:

1. Starts upstream `llama-server` as a child on `127.0.0.1:8081`
2. Listens on `:8080` (the port the SmoothNAS nginx route forwards to)
3. Validates `Authorization: Bearer $SMOOTHNAS_BEARER_EXPECTED` on every request — constant-time compare, rejects with 401 on missing or wrong token
4. Forwards valid requests to the child via `httputil.NewSingleHostReverseProxy` with streaming-friendly defaults (so SSE / token-by-token completions work)

`SMOOTHNAS_BEARER_EXPECTED` is auto-populated by tierd at install time and rotated via `POST /api/plugins/llama-cpp/rotate-token`. The operator never sees or touches the token directly — it's owned by SmoothNAS.

## Operator workflow

In the SmoothNAS UI:

1. **Install** → paste this manifest into the wizard, pick a tier with NVME slot capacity, install
2. **Stage models** → copy GGUF files into `/mnt/<tier>/.plugins/llama-cpp/models/` over SMB / NFS / SCP
3. **Configure** → set `MODEL_PATH` on the detail page (default `/models/qwen3.6-27b-128k-q5.gguf`)
4. **Start** → click Start; tierd pulls the wrapper image, creates the container, captures the bridge IP, writes the nginx route with the bearer token
5. **Open** → click Open on the card; the llama.cpp UI renders inside the SmoothNAS chrome at `https://<smoothnas>/plugins/llama-cpp/`

The default runtime profile is Qwen3.6 27B 128K Q5:

- `MODEL_PATH=/models/qwen3.6-27b-128k-q5.gguf`
- `CTX_SIZE=131072`
- `PARALLEL_SLOTS=2`
- `N_GPU_LAYERS=999` on CUDA/Vulkan manifests

Quantization is selected by the GGUF file itself. To use a different
quantization, stage that GGUF under the models volume and update
`MODEL_PATH`; the default is intentionally Q5, not Q4.

Uninstall via the UI's Danger Zone removes the container, the cached image, the nginx route, and **the model files** (per the parent doc's all-or-none policy). Back up `/mnt/<tier>/.plugins/llama-cpp/models/` first if you care.

## Local development

```sh
# wrapper smoke build + tests
cd wrapper && go build ./... && go test ./...

# image build (any variant)
docker buildx build \
  --build-arg LLAMA_BASE=ghcr.io/ggml-org/llama.cpp:server-cuda \
  -t smoothnas-plugin-llama-cpp:dev-cuda .
```

To sideload a dev image into a SmoothNAS dev VM, edit the matching manifest's `artifact.image` to your local tag and use `tierd-cli plugin install`.

## Release flow

`.github/workflows/release.yml` runs on tag push (`v*`):

1. Builds three image variants (CUDA / Vulkan / CPU) via buildx + GHCR push
2. Resolves the pushed digests and rewrites `artifact.digest` in each manifest
3. Creates a GitHub release attaching the three rewritten manifests
4. Pushes a release-channel `index.json` so the SmoothNAS install wizard can offer the variants alphabetically

The smoke test that actually installs the plugin against a SmoothNAS dev VM lives in the SmoothNAS repo's CI, not here. Triggered nightly so a SmoothNAS release that broke the plugin contract is caught quickly.

## License

Add a LICENSE file at publish time. The wrapper code in this repo is original; downstream images carry their own license terms (upstream llama.cpp is MIT).
