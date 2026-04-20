package allocator

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type Allocator struct {
	totalAllocations uint64
}

func New() *Allocator {
	return &Allocator{}
}

func (a *Allocator) BuildEnvs(deviceIDs []string) map[string]string {
	atomic.AddUint64(&a.totalAllocations, 1)
	gpuIndexes := make([]string, 0, len(deviceIDs))
	for _, id := range deviceIDs {
		gpuIndexes = append(gpuIndexes, extractIndex(id))
	}
	return map[string]string{
		"FAKE_GPU":               "true",
		"CUDA_VISIBLE_DEVICES":   strings.Join(gpuIndexes, ","),
		"NVIDIA_VISIBLE_DEVICES": "all",
	}
}

func (a *Allocator) TotalAllocations() uint64 {
	return atomic.LoadUint64(&a.totalAllocations)
}

func extractIndex(deviceID string) string {
	parts := strings.Split(deviceID, "-")
	for i, p := range parts {
		if p == "gpu" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return fmt.Sprintf("%q", deviceID)
}
