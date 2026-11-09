package collector

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemMetrics struct {
	CPUPercent       float64 `json:"cpu_percent"`
	CPUTEMPC         float64 `json:"cpu_temp_c"`
	MemoryUsedBytes  uint64  `json:"memory_used_bytes"`
	MemoryTotalBytes uint64  `json:"memory_total_bytes"`
	DiskUsedBytes    uint64  `json:"disk_used_bytes"`
	DiskTotalBytes   uint64  `json:"disk_total_bytes"`
}

func Collect() (*SystemMetrics, error) {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	memory, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	diskInfo, err := disk.Usage("C:\\")
	if err != nil {
		return nil, err
	}

	metrics := &SystemMetrics{
		CPUPercent:       0,
		CPUTEMPC:         0,
		MemoryUsedBytes:  memory.Used,
		MemoryTotalBytes: memory.Total,
		DiskUsedBytes:    diskInfo.Used,
		DiskTotalBytes:   diskInfo.Total,
	}

	if len(cpuPercent) > 0 {
		metrics.CPUPercent = cpuPercent[0]
	}

	return metrics, nil
}
