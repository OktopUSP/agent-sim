package simulator

import (
	"log/slog"
	"time"

	"github.com/shirou/gopsutil/process"
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

		for _, proc := range processList {
			cpuPercent, err := proc.CPUPercent()
			if err != nil {
				slog.Error("Error getting CPU percent", "pid", proc.Pid, "error", err)
			}
			memInfo, err := proc.MemoryInfo()
			if err != nil {
				slog.Error("Error getting memory info", "pid", proc.Pid, "error", err)
			}

			slog.Debug("Process stats",
				slog.Int("pid", int(proc.Pid)),
				slog.Float64("cpu_percent", cpuPercent),
				slog.Uint64("memory_kb", memInfo.RSS/(1024*1024)),
			)

			totalCPU += cpuPercent
			totalMemory += memInfo.RSS / (1024 * 1024)
		}

		slog.Info("Total compute resources usage",
			slog.Float64("total_cpu_percent", totalCPU),
			slog.Uint64("total_memory_mb", totalMemory),
		)

		time.Sleep(1 * time.Second)
	}
}
