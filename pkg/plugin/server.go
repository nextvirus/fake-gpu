package plugin

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fake-gpu-platform/pkg/allocator"
	"fake-gpu-platform/pkg/config"
	"fake-gpu-platform/pkg/device"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type Server struct {
	pluginapi.UnimplementedDevicePluginServer

	resourceName string
	socketPath   string
	kubeletSock  string
	devices      []device.FakeDevice
	alloc        *allocator.Allocator

	grpcServer *grpc.Server
	mu         sync.Mutex
}

func New(resourceName string, cfg config.Config, devices []device.FakeDevice, alloc *allocator.Allocator) *Server {
	socket := filepath.Join(cfg.PluginSocketDir, socketName(resourceName))
	return &Server{
		resourceName: resourceName,
		socketPath:   socket,
		kubeletSock:  cfg.KubeletSocketPath,
		devices:      devices,
		alloc:        alloc,
	}
}

func socketName(resourceName string) string {
	r := strings.ReplaceAll(resourceName, "/", "-")
	r = strings.ReplaceAll(r, ".", "-")
	return fmt.Sprintf("fake-gpu-%s.sock", r)
}

func (s *Server) Run(ctx context.Context) error {
	if err := os.Remove(s.socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	lis, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}

	s.grpcServer = grpc.NewServer()
	pluginapi.RegisterDevicePluginServer(s.grpcServer, s)

	go func() {
		<-ctx.Done()
		s.Stop()
	}()
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Printf("device plugin server stopped for %s: %v", s.resourceName, err)
		}
	}()

	if err := s.register(); err != nil {
		return err
	}
	log.Printf("registered fake gpu plugin for %s with %d devices", s.resourceName, len(s.devices))
	return nil
}

func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
}

func (s *Server) register() error {
	conn, err := grpc.Dial(
		"unix://"+s.kubeletSock,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return fmt.Errorf("connect kubelet socket: %w", err)
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	_, err = client.Register(context.Background(), &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     filepath.Base(s.socketPath),
		ResourceName: s.resourceName,
		Options: &pluginapi.DevicePluginOptions{
			PreStartRequired: false,
		},
	})
	if err != nil {
		return fmt.Errorf("register device plugin: %w", err)
	}
	return nil
}

func (s *Server) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{PreStartRequired: false}, nil
}

func (s *Server) ListAndWatch(_ *pluginapi.Empty, stream pluginapi.DevicePlugin_ListAndWatchServer) error {
	for {
		resp := &pluginapi.ListAndWatchResponse{
			Devices: device.ToPluginDevices(s.devices),
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
		time.Sleep(30 * time.Second)
	}
}

func (s *Server) Allocate(_ context.Context, req *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	resp := &pluginapi.AllocateResponse{
		ContainerResponses: make([]*pluginapi.ContainerAllocateResponse, 0, len(req.ContainerRequests)),
	}
	for _, cr := range req.ContainerRequests {
		envs := s.alloc.BuildEnvs(cr.DevicesIDs)
		resp.ContainerResponses = append(resp.ContainerResponses, &pluginapi.ContainerAllocateResponse{
			Envs: envs,
		})
	}
	return resp, nil
}

func (s *Server) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (s *Server) GetPreferredAllocation(context.Context, *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func StartMetricsServer(ctx context.Context, bind string, alloc *allocator.Allocator) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "# TYPE fake_gpu_allocations_total counter\nfake_gpu_allocations_total %d\n", alloc.TotalAllocations())
	})

	srv := &http.Server{
		Addr:    bind,
		Handler: mux,
	}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
