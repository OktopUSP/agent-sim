package simulator

import (
	"log/slog"
	"time"

	"github.com/shirou/gopsutil/process"
	"github.com/shirou/gopsutil/v3/net"
)

func monitorProcesses(pids []int) {
	processList := []*process.Process{}

	for _, pid := range pids {
		proc, err := process.NewProcess(int32(pid))
		if err != nil {
			slog.Error("Error getting process info", "pid", pid, "error", err)
			continue
		}
		processList = append(processList, proc)
	}

	for {
		var totalCPU float64
		var totalMemory uint64
		var totalConnections int

		for _, proc := range processList {
			cpuPercent, err := proc.CPUPercent()
			if err != nil {
				slog.Error("Error getting CPU percent", "pid", proc.Pid, "error", err)
				continue
			}
			memInfo, err := proc.MemoryInfo()
			if err != nil {
				slog.Error("Error getting memory info", "pid", proc.Pid, "error", err)
				continue
			}
			connections, err := net.ConnectionsPid("tcp", proc.Pid)
			if err != nil {
				slog.Error("Error getting tcp connections", "pid", proc.Pid, "error", err)
				continue
			}
			numConnections := len(connections)

			slog.Debug("Process stats",
				slog.Int("pid", int(proc.Pid)),
				slog.Float64("cpu_percent", cpuPercent),
				slog.Uint64("memory_kb", memInfo.RSS/(1024*1024)),
				slog.Int("connections", numConnections),
			)

			totalCPU += cpuPercent
			totalMemory += memInfo.RSS / (1024 * 1024)
			totalConnections += numConnections
		}

		slog.Info("Total compute resources usage",
			slog.Float64("cpu_percent", totalCPU),
			slog.Uint64("memory_mb", totalMemory),
			slog.Int("tcp_connections", totalConnections),
		)

		time.Sleep(1 * time.Second)
	}
}
