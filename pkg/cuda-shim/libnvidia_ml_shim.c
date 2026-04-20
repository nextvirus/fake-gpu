#define _GNU_SOURCE
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

typedef int nvmlReturn_t;
typedef void *nvmlDevice_t;

#define NVML_SUCCESS 0
#define NVML_ERROR_INVALID_ARGUMENT 2

nvmlReturn_t nvmlInit_v2(void) {
  return NVML_SUCCESS;
}

nvmlReturn_t nvmlInit(void) {
  return NVML_SUCCESS;
}

nvmlReturn_t nvmlInitWithFlags(unsigned int flags) {
  (void)flags;
  return NVML_SUCCESS;
}

nvmlReturn_t nvmlShutdown(void) {
  return NVML_SUCCESS;
}

nvmlReturn_t nvmlSystemGetDriverVersion(char *version, unsigned int length) {
  if (!version || length == 0) return NVML_ERROR_INVALID_ARGUMENT;
  const char *fake = "555.42.02";
  strncpy(version, fake, length - 1);
  version[length - 1] = '\0';
  return NVML_SUCCESS;
}

nvmlReturn_t nvmlDeviceGetCount_v2(unsigned int *deviceCount) {
  if (!deviceCount) return NVML_ERROR_INVALID_ARGUMENT;
  const char *env = getenv("FAKE_GPU_COUNT");
  if (!env) {
    *deviceCount = 1;
    return NVML_SUCCESS;
  }
  int v = atoi(env);
  *deviceCount = (v > 0) ? (unsigned int)v : 1U;
  return NVML_SUCCESS;
}

nvmlReturn_t nvmlDeviceGetCount(unsigned int *deviceCount) {
  return nvmlDeviceGetCount_v2(deviceCount);
}

nvmlReturn_t nvmlDeviceGetHandleByIndex_v2(unsigned int index, nvmlDevice_t *device) {
  if (!device) return NVML_ERROR_INVALID_ARGUMENT;
  *device = (void *)(uintptr_t)(index + 1);
  return NVML_SUCCESS;
}

nvmlReturn_t nvmlDeviceGetHandleByIndex(unsigned int index, nvmlDevice_t *device) {
  return nvmlDeviceGetHandleByIndex_v2(index, device);
}

nvmlReturn_t nvmlDeviceGetName(nvmlDevice_t device, char *name, unsigned int length) {
  (void)device;
  if (!name || length == 0) return NVML_ERROR_INVALID_ARGUMENT;
  const char *model = getenv("FAKE_GPU_MODEL");
  if (!model || model[0] == '\0') model = "Fake GPU";
  strncpy(name, model, length - 1);
  name[length - 1] = '\0';
  return NVML_SUCCESS;
}

nvmlReturn_t nvmlDeviceGetCudaComputeCapability(nvmlDevice_t device, int *major, int *minor) {
  (void)device;
  if (!major || !minor) return NVML_ERROR_INVALID_ARGUMENT;
  *major = 9;
  *minor = 0;
  return NVML_SUCCESS;
}

const char *nvmlErrorString(nvmlReturn_t result) {
  (void)result;
  return "NVML shim";
}
