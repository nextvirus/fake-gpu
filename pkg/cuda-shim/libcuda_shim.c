#define _GNU_SOURCE
#include <stdlib.h>
#include <string.h>

typedef int cudaError_t;
typedef int CUresult;
typedef int CUdevice;
typedef void *CUcontext;
typedef void *CUmodule;
typedef void *CUfunction;
typedef unsigned long long CUdeviceptr;

#define cudaSuccess 0
#define cudaErrorInvalidValue 1
#define CUDA_SUCCESS 0
#define CUDA_ERROR_INVALID_VALUE 1

cudaError_t cudaGetDeviceCount(int *count) {
  if (!count) return cudaErrorInvalidValue;
  const char *env = getenv("FAKE_GPU_COUNT");
  if (!env) {
    *count = 1;
    return cudaSuccess;
  }
  int v = atoi(env);
  *count = (v > 0) ? v : 1;
  return cudaSuccess;
}

cudaError_t cudaSetDevice(int device) {
  (void)device;
  return cudaSuccess;
}

cudaError_t cudaMalloc(void **devPtr, size_t size) {
  if (!devPtr) return cudaErrorInvalidValue;
  *devPtr = malloc(size);
  if (!*devPtr) return 2;
  return cudaSuccess;
}

cudaError_t cudaFree(void *devPtr) {
  free(devPtr);
  return cudaSuccess;
}

cudaError_t cudaMemcpy(void *dst, const void *src, size_t count, int kind) {
  (void)kind;
  if (!dst || !src) return cudaErrorInvalidValue;
  memcpy(dst, src, count);
  return cudaSuccess;
}

CUresult cuInit(unsigned int flags) {
  (void)flags;
  return CUDA_SUCCESS;
}

CUresult cuDriverGetVersion(int *driverVersion) {
  if (!driverVersion) return CUDA_ERROR_INVALID_VALUE;
  *driverVersion = 12040;
  return CUDA_SUCCESS;
}

CUresult cuDeviceGetCount(int *count) {
  if (!count) return CUDA_ERROR_INVALID_VALUE;
  const char *env = getenv("FAKE_GPU_COUNT");
  if (!env) {
    *count = 1;
    return CUDA_SUCCESS;
  }
  int v = atoi(env);
  *count = (v > 0) ? v : 1;
  return CUDA_SUCCESS;
}

CUresult cuDeviceGet(CUdevice *device, int ordinal) {
  if (!device || ordinal < 0) return CUDA_ERROR_INVALID_VALUE;
  *device = ordinal;
  return CUDA_SUCCESS;
}

CUresult cuDeviceGetName(char *name, int len, CUdevice dev) {
  (void)dev;
  if (!name || len <= 0) return CUDA_ERROR_INVALID_VALUE;
  const char *model = getenv("FAKE_GPU_MODEL");
  if (!model || model[0] == '\0') model = "Fake GPU";
  strncpy(name, model, (size_t)(len - 1));
  name[len - 1] = '\0';
  return CUDA_SUCCESS;
}

CUresult cuCtxCreate_v2(CUcontext *pctx, unsigned int flags, CUdevice dev) {
  (void)flags;
  (void)dev;
  if (!pctx) return CUDA_ERROR_INVALID_VALUE;
  *pctx = (void *)0x1;
  return CUDA_SUCCESS;
}

CUresult cuCtxDestroy_v2(CUcontext ctx) {
  (void)ctx;
  return CUDA_SUCCESS;
}

CUresult cuModuleLoadData(CUmodule *module, const void *image) {
  (void)image;
  if (!module) return CUDA_ERROR_INVALID_VALUE;
  *module = (void *)0x1;
  return CUDA_SUCCESS;
}

CUresult cuModuleUnload(CUmodule hmod) {
  (void)hmod;
  return CUDA_SUCCESS;
}

CUresult cuModuleGetFunction(CUfunction *hfunc, CUmodule hmod, const char *name) {
  (void)hmod;
  (void)name;
  if (!hfunc) return CUDA_ERROR_INVALID_VALUE;
  *hfunc = (void *)0x1;
  return CUDA_SUCCESS;
}
