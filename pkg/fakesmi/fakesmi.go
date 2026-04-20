package fakesmi

import (
	"fmt"
	"io"
	"time"
)

type Options struct {
	GPUCount     int
	GPUModel     string
	GPUMemoryMiB int
	CUDAVersion  string
}

func Write(w io.Writer, o Options) {
	now := time.Now().Format("Mon Jan 02 15:04:05 2006")
	fmt.Fprintf(w, "Tue %s\n", now)
	fmt.Fprintln(w, "+-----------------------------------------------------------------------------------------+")
	fmt.Fprintf(w, "| NVIDIA-SMI 555.55.55              Driver Version: 555.55.55      CUDA Version: %-8s|\n", o.CUDAVersion)
	fmt.Fprintln(w, "|-----------------------------------------+------------------------+----------------------+")
	fmt.Fprintln(w, "| GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |")
	fmt.Fprintln(w, "| Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |")
	fmt.Fprintln(w, "|                                         |                        |               MIG M. |")
	fmt.Fprintln(w, "|=========================================+========================+======================|")
	for i := 0; i < o.GPUCount; i++ {
		fmt.Fprintf(w, "| %3d  %-20s Off | 00000000:%02X:00.0 Off |                  Off |\n", i, trim(o.GPUModel, 20), i+1)
		fmt.Fprintf(w, "| N/A   35C    P0              25W / 300W |      0MiB / %5dMiB |      0%%      Default |\n", o.GPUMemoryMiB)
		fmt.Fprintln(w, "|                                         |                        |                  N/A |")
		fmt.Fprintln(w, "+-----------------------------------------+------------------------+----------------------+")
	}
}

func trim(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
