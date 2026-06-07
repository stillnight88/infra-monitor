package collector

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/stillnight88/infra-monitor/shared"
	"time"
)

func Collect(agentID, hostname string) (shared.MetricsPayload, error) {
	cpuPercents, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return shared.MetricsPayload{}, fmt.Errorf("cpu: %w", err)
	}

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return shared.MetricsPayload{}, fmt.Errorf("ram: %w", err)
	}

	diskStat, err := disk.Usage("/")
	if err != nil {
		return shared.MetricsPayload{}, fmt.Errorf("disk: %w", err)
	}

	return shared.MetricsPayload{
		AgentID:   agentID,
		Hostname:  hostname,
		CPU:       cpuPercents[0],
		RAM:       vmStat.UsedPercent,
		Disk:      diskStat.UsedPercent,
		Timestamp: time.Now().Unix(),
	}, nil
}
