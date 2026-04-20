package device

import (
	"fmt"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type FakeDevice struct {
	ID        string
	Model     string
	MemoryMiB int
	Health    string
}

func BuildDevices(count, sharingFactor int, model string, memoryMiB int) []FakeDevice {
	if sharingFactor < 1 {
		sharingFactor = 1
	}
	devices := make([]FakeDevice, 0, count*sharingFactor)
	for gpu := 0; gpu < count; gpu++ {
		for share := 0; share < sharingFactor; share++ {
			id := fmt.Sprintf("fake-gpu-%d", gpu)
			if sharingFactor > 1 {
				id = fmt.Sprintf("fake-gpu-%d-share-%d", gpu, share)
			}
			devices = append(devices, FakeDevice{
				ID:        id,
				Model:     model,
				MemoryMiB: memoryMiB,
				Health:    pluginapi.Healthy,
			})
		}
	}
	return devices
}

func ToPluginDevices(devs []FakeDevice) []*pluginapi.Device {
	out := make([]*pluginapi.Device, 0, len(devs))
	for _, d := range devs {
		out = append(out, &pluginapi.Device{
			ID:     d.ID,
			Health: d.Health,
		})
	}
	return out
}
