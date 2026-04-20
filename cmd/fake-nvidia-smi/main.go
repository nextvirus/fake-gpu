package main

import (
	"os"
	"strconv"

	"fake-gpu-platform/pkg/config"
	"fake-gpu-platform/pkg/fakesmi"
)

func main() {
	cfg, err := config.Load(os.Getenv("FAKE_GPU_CONFIG_PATH"))
	if err != nil {
		cfg = config.Config{
			GPUCount:     config.DefaultGPUCount,
			GPUModel:     config.DefaultGPUModel,
			GPUMemoryMiB: config.DefaultGPUMemoryMiB,
			CUDAVersion:  config.DefaultCUDAVersion,
		}
	}
	if v := os.Getenv("FAKE_GPU_COUNT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.GPUCount = i
		}
	}
	fakesmi.Write(os.Stdout, fakesmi.Options{
		GPUCount:     cfg.GPUCount,
		GPUModel:     cfg.GPUModel,
		GPUMemoryMiB: cfg.GPUMemoryMiB,
		CUDAVersion:  cfg.CUDAVersion,
	})
}
