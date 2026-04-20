package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath       = "/etc/fake-gpu/config.yaml"
	DefaultResourceName     = "nvidia.com/gpu"
	DefaultGPUCount         = 8
	DefaultGPUMemoryMiB     = 8192
	DefaultGPUModel         = "Fake RTX 4090"
	DefaultCUDAVersion      = "12.4"
	DefaultMetricsBindAddr  = ":9095"
	DefaultSharingFactor    = 1
	DefaultRequireNodeLabel = false
	DefaultNodeLabelKey     = "fake-gpu/enabled"
	DefaultNodeLabelValue   = "true"
)

type Config struct {
	ResourceNames     []string `yaml:"resourceNames"`
	GPUCount          int      `yaml:"gpuCount"`
	GPUMemoryMiB      int      `yaml:"gpuMemoryMiB"`
	GPUModel          string   `yaml:"gpuModel"`
	CUDAVersion       string   `yaml:"cudaVersion"`
	RequireNodeLabel  bool     `yaml:"requireNodeLabel"`
	NodeLabelKey      string   `yaml:"nodeLabelKey"`
	NodeLabelValue    string   `yaml:"nodeLabelValue"`
	SharingFactor     int      `yaml:"sharingFactor"`
	MetricsBindAddr   string   `yaml:"metricsBindAddr"`
	PluginSocketDir   string   `yaml:"pluginSocketDir"`
	KubeletSocketPath string   `yaml:"kubeletSocketPath"`
}

func Load(path string) (Config, error) {
	cfg := defaultConfig()
	if path == "" {
		path = DefaultConfigPath
	}

	if b, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			return Config{}, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	overrideFromEnv(&cfg)
	normalize(&cfg)
	return cfg, validate(cfg)
}

func defaultConfig() Config {
	return Config{
		ResourceNames:     []string{DefaultResourceName},
		GPUCount:          DefaultGPUCount,
		GPUMemoryMiB:      DefaultGPUMemoryMiB,
		GPUModel:          DefaultGPUModel,
		CUDAVersion:       DefaultCUDAVersion,
		RequireNodeLabel:  DefaultRequireNodeLabel,
		NodeLabelKey:      DefaultNodeLabelKey,
		NodeLabelValue:    DefaultNodeLabelValue,
		SharingFactor:     DefaultSharingFactor,
		MetricsBindAddr:   DefaultMetricsBindAddr,
		PluginSocketDir:   "/var/lib/kubelet/device-plugins",
		KubeletSocketPath: "/var/lib/kubelet/device-plugins/kubelet.sock",
	}
}

func overrideFromEnv(cfg *Config) {
	if v := os.Getenv("FAKE_GPU_RESOURCE_NAMES"); v != "" {
		cfg.ResourceNames = strings.Split(v, ",")
	}
	if v := os.Getenv("FAKE_GPU_COUNT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.GPUCount = i
		}
	}
	if v := os.Getenv("FAKE_GPU_MEMORY_MIB"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.GPUMemoryMiB = i
		}
	}
	if v := os.Getenv("FAKE_GPU_MODEL"); v != "" {
		cfg.GPUModel = v
	}
	if v := os.Getenv("FAKE_GPU_CUDA_VERSION"); v != "" {
		cfg.CUDAVersion = v
	}
	if v := os.Getenv("FAKE_GPU_REQUIRE_NODE_LABEL"); v != "" {
		cfg.RequireNodeLabel = strings.EqualFold(v, "true")
	}
	if v := os.Getenv("FAKE_GPU_NODE_LABEL_KEY"); v != "" {
		cfg.NodeLabelKey = v
	}
	if v := os.Getenv("FAKE_GPU_NODE_LABEL_VALUE"); v != "" {
		cfg.NodeLabelValue = v
	}
	if v := os.Getenv("FAKE_GPU_SHARING_FACTOR"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.SharingFactor = i
		}
	}
	if v := os.Getenv("FAKE_GPU_METRICS_BIND_ADDR"); v != "" {
		cfg.MetricsBindAddr = v
	}
	if v := os.Getenv("FAKE_GPU_PLUGIN_SOCKET_DIR"); v != "" {
		cfg.PluginSocketDir = v
	}
	if v := os.Getenv("FAKE_GPU_KUBELET_SOCKET"); v != "" {
		cfg.KubeletSocketPath = v
	}
}

func normalize(cfg *Config) {
	for i := range cfg.ResourceNames {
		cfg.ResourceNames[i] = strings.TrimSpace(cfg.ResourceNames[i])
	}
}

func validate(cfg Config) error {
	if len(cfg.ResourceNames) == 0 {
		return errors.New("at least one resource name is required")
	}
	for _, rn := range cfg.ResourceNames {
		if rn == "" {
			return errors.New("resource name cannot be empty")
		}
	}
	if cfg.GPUCount < 1 {
		return errors.New("gpuCount must be >= 1")
	}
	if cfg.SharingFactor < 1 {
		return errors.New("sharingFactor must be >= 1")
	}
	if cfg.GPUMemoryMiB < 1 {
		return errors.New("gpuMemoryMiB must be >= 1")
	}
	return nil
}
