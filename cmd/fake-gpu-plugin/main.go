package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"fake-gpu-platform/pkg/allocator"
	"fake-gpu-platform/pkg/config"
	"fake-gpu-platform/pkg/device"
	"fake-gpu-platform/pkg/plugin"
)

func main() {
	cfgPath := os.Getenv("FAKE_GPU_CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if !nodeEnabled(cfg) {
		log.Printf("fake gpu plugin disabled by node label gate")
		select {}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	devices := device.BuildDevices(cfg.GPUCount, cfg.SharingFactor, cfg.GPUModel, cfg.GPUMemoryMiB)
	alloc := allocator.New()

	for _, resourceName := range cfg.ResourceNames {
		srv := plugin.New(resourceName, cfg, devices, alloc)
		go keepRegistered(ctx, srv, resourceName)
	}

	go func() {
		if err := plugin.StartMetricsServer(ctx, cfg.MetricsBindAddr, alloc); err != nil {
			log.Printf("metrics server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("fake gpu plugin stopped")
}

func keepRegistered(ctx context.Context, srv *plugin.Server, resourceName string) {
	for {
		if err := srv.Run(ctx); err != nil {
			log.Printf("plugin %s run failure, retrying: %v", resourceName, err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func nodeEnabled(cfg config.Config) bool {
	if !cfg.RequireNodeLabel {
		return true
	}
	nodeLabels := os.Getenv("FAKE_GPU_NODE_LABELS")
	if nodeLabels == "" {
		return false
	}
	for _, kv := range strings.Split(nodeLabels, ",") {
		pair := strings.SplitN(strings.TrimSpace(kv), "=", 2)
		if len(pair) != 2 {
			continue
		}
		if pair[0] == cfg.NodeLabelKey && pair[1] == cfg.NodeLabelValue {
			return true
		}
	}
	return false
}
