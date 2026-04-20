## 中文说明（无 GPU 场景）

这个项目的目标是：在没有真实 GPU 的节点上，尽量模拟 GPU 服务的调度与启动流程，主要用于 CI、联调、功能演示。

### 能解决什么

- 通过 Device Plugin 向 Kubernetes 上报 `nvidia.com/gpu` 资源
- 让声明了 GPU 资源请求的 Pod 可以被正常调度
- 注入 `FAKE_GPU`、`CUDA_VISIBLE_DEVICES` 等常见环境变量
- 通过 fake `libcuda.so.1` + `LD_PRELOAD` 通过一部分 CUDA 存在性检查

### 不能解决什么

- 不能提供真实 CUDA 计算能力
- 不能替代真实 GPU 驱动、算子、内核执行
- 对于强依赖 CUDA 内核执行的框架，只能“启动兼容”，不能保证完整功能

### 推荐部署方式（vLLM 示例）

在 CPU-only 节点上，如需尽量跑通“GPU 服务链路”，建议同时满足：

1. 调度层声明 GPU 资源（例如 `nvidia.com/gpu: 1`），走 fake-gpu 调度路径
2. 推理层显式使用 CPU（例如 vLLM 使用 `--device cpu`）
3. 通过 `cuda-shim-configmap.yaml` + `initContainer` 构建并注入 fake `libcuda.so.1`

### 建议执行顺序

```bash
kubectl apply -f deploy/k8s/fake-gpu-plugin.yaml
kubectl apply -f deploy/k8s/cuda-shim-configmap.yaml
kubectl apply -f deploy/k8s/vllm.yaml
```

### 适用边界

- 适合：资源调度验证、平台联调、接口联调、自动化测试占位
- 不适合：性能测试、吞吐/延迟评估、真实 GPU 训练与推理压测

## Metrics

`/metrics` endpoint exposes:

- `fake_gpu_allocations_total`

Default bind address: `:9095`.

# Fake GPU Mock Platform for Kubernetes

Production-style fake GPU platform for Kubernetes CI/testing environments.  
It advertises virtual GPUs via the official Device Plugin API so GPU-requesting Pods can schedule and run on CPU-only nodes.

## What It Provides

- Kubernetes Device Plugin (`v1beta1`) with resource name `nvidia.com/gpu` (and optional `fake.com/gpu`)
- Configurable fake GPU inventory per node (default: 8)
- Health reporting (`Healthy` always)
- `Allocate` behavior that injects GPU-related environment variables into containers
- Fake `nvidia-smi` binary (`/usr/local/bin/nvidia-smi`) with configurable model/memory/CUDA version
- Optional CUDA runtime shim (`LD_PRELOAD`) for lightweight GPU detection testing
- DaemonSet + RBAC + ServiceAccount manifests for `kube-system`
- Helm chart for deployment customization

## Repository Layout

```text
cmd/
  fake-gpu-plugin/       # Device plugin daemon
  fake-nvidia-smi/       # Fake nvidia-smi command
pkg/
  allocator/             # Allocate env injection + counters
  config/                # YAML/env config loader
  cuda-shim/             # Optional C LD_PRELOAD shim
  device/                # Fake device modeling
  fakesmi/               # nvidia-smi output renderer
  plugin/                # Device plugin gRPC server
deploy/
  k8s/                   # Raw manifests
  helm/                  # Helm chart
```

## Device Plugin Behavior

- `ListAndWatch`: reports `fake-gpu-0..N` (or share-suffixed IDs when sharing enabled)
- `Allocate`: returns:
  - `FAKE_GPU=true`
  - `CUDA_VISIBLE_DEVICES=<indices>`
  - `NVIDIA_VISIBLE_DEVICES=all`
- Health: always `Healthy`

## Configuration

The plugin loads config from `/etc/fake-gpu/config.yaml` by default, then applies env overrides.

Key fields:

- `resourceNames`: list of resource names, e.g. `["nvidia.com/gpu","fake.com/gpu"]`
- `gpuCount`: fake GPU count per node
- `gpuMemoryMiB`: fake per-GPU memory
- `gpuModel`: fake model name
- `cudaVersion`: fake CUDA version used by `nvidia-smi`
- `sharingFactor`: creates multiple allocatable shares per GPU
- `requireNodeLabel`: enable node label gate
- `nodeLabelKey` / `nodeLabelValue`: gate label pair (`fake-gpu/enabled=true`)

Environment overrides:

- `FAKE_GPU_RESOURCE_NAMES`
- `FAKE_GPU_COUNT`
- `FAKE_GPU_MEMORY_MIB`
- `FAKE_GPU_MODEL`
- `FAKE_GPU_CUDA_VERSION`
- `FAKE_GPU_SHARING_FACTOR`
- `FAKE_GPU_REQUIRE_NODE_LABEL`
- `FAKE_GPU_NODE_LABEL_KEY`
- `FAKE_GPU_NODE_LABEL_VALUE`
- `FAKE_GPU_NODE_LABELS` (comma-separated `k=v` list used for label gate)

## Build

```bash
go mod tidy
go build ./...
```

Docker image:

```bash
docker build -t fake-gpu-plugin:latest .
```

## Deploy to kind (local)

1. Create cluster:

```bash
kind create cluster --name fake-gpu
```

2. Build + load image:

```bash
docker build -t fake-gpu-plugin:latest .
kind load docker-image fake-gpu-plugin:latest --name fake-gpu
```

3. Deploy plugin:

```bash
kubectl apply -f deploy/k8s/fake-gpu-plugin.yaml
```

4. Verify node allocatable resources:

```bash
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.allocatable.nvidia\.com/gpu}{"\n"}{end}'
```

## Test GPU Scheduling

```bash
kubectl apply -f deploy/k8s/gpu-test-pod.yaml
kubectl get pod fake-gpu-test -o wide
kubectl describe pod fake-gpu-test
kubectl exec fake-gpu-test -- env | grep -E 'FAKE_GPU|CUDA_VISIBLE_DEVICES|NVIDIA_VISIBLE_DEVICES'
```

If scheduled and running with `nvidia.com/gpu` limit, Kubernetes scheduler behavior is validated.

## Fake nvidia-smi

The container image installs `/usr/local/bin/nvidia-smi` from `cmd/fake-nvidia-smi`.

Example:

```bash
nvidia-smi
```

Outputs a realistic table with fake GPU model/memory/CUDA values and multiple devices.

## Optional CUDA Shim (LD_PRELOAD)

Source: `pkg/cuda-shim/libcuda_shim.c`

Build shared library:

```bash
gcc -shared -fPIC -o libcuda.so pkg/cuda-shim/libcuda_shim.c
```

Run app with shim:

```bash
LD_PRELOAD=/path/to/libcuda.so python your_gpu_detection_test.py
```

Implemented shim functions:

- `cudaGetDeviceCount()`
- `cudaSetDevice()`
- `cudaMalloc()` / `cudaFree()`
- `cudaMemcpy()`

All calls are CPU-backed stubs intended only for compatibility tests.

### Kubernetes injection example

For GPU-like scheduling on CPU-only nodes, keep Pod resource requests as `nvidia.com/gpu`
and force your workload to CPU execution (for example, vLLM `--device cpu`).

This repository also includes:

- `deploy/k8s/cuda-shim-configmap.yaml`: stores shim source code
- `deploy/k8s/vllm.yaml`: uses an `initContainer` to compile `/opt/fake-cuda/libcuda.so.1`
  and injects it with `LD_PRELOAD`/`LD_LIBRARY_PATH`

Apply order:

```bash
kubectl apply -f deploy/k8s/fake-gpu-plugin.yaml
kubectl apply -f deploy/k8s/cuda-shim-configmap.yaml
kubectl apply -f deploy/k8s/vllm.yaml
```

This setup can pass many CUDA presence checks, but does not provide real GPU compute.


