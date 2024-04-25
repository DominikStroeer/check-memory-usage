package main

import (
	"fmt"
	"time"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
	"github.com/shirou/gopsutil/v3/mem"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Critical float64
	Warning  float64
	Interval int
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "check-memory-usage",
			Short:    "Check memory usage and provide metrics",
			Keyspace: "sensu.io/plugins/check-memory-usage/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "critical",
			Argument:  "critical",
			Shorthand: "c",
			Default:   float64(90),
			Usage:     "Critical threshold for overall memory usage",
			Value:     &plugin.Critical,
		},
		{
			Path:      "warning",
			Argument:  "warning",
			Shorthand: "w",
			Default:   float64(75),
			Usage:     "Warning threshold for overall memory usage",
			Value:     &plugin.Warning,
		},
		{
			Path:      "sample-interval",
			Argument:  "sample-interval",
			Shorthand: "s",
			Default:   20,
			Usage:     "Length of sample interval in seconds",
			Value:     &plugin.Interval,
		},
	}
)

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	if plugin.Critical == 0 {
		return sensu.CheckStateWarning, fmt.Errorf("--critical is required")
	}
	if plugin.Warning == 0 {
		return sensu.CheckStateWarning, fmt.Errorf("--warning is required")
	}
	if plugin.Warning > plugin.Critical {
		return sensu.CheckStateWarning, fmt.Errorf("--warning cannot be greater than --critical")
	}
	if plugin.Interval == 0 {
		return sensu.CheckStateWarning, fmt.Errorf("--interval is required")
	}
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {
	//vmStat, err := mem.VirtualMemory()
	vmStart, err := mem.VirtualMemory()
	duration, err := time.ParseDuration(fmt.Sprintf("%ds", plugin.Interval))

	if err != nil {
		return sensu.CheckStateCritical, fmt.Errorf("failed to get virtual memory statistics: %v", err)
	}
	startTotal := vmStart.UsedPercent
	time.Sleep(duration)

	vmEnd, err := mem.VirtualMemory()
	endTotal := vmEnd.UsedPercent

	perfData := fmt.Sprintf("mem_total=%d, mem_available=%d, mem_used=%d, mem_free=%d", vmStart.Total, vmStart.Available, vmStart.Used, vmStart.Free)
	if (startTotal > plugin.Critical) && (endTotal > plugin.Critical) {
		fmt.Printf("%s Critical: %.2f%% memory usage in the beginning and %.2f%% afterwards | %s\n", plugin.PluginConfig.Name, vmStart.UsedPercent, vmEnd.UsedPercent, perfData)

		return sensu.CheckStateCritical, nil
	} else if (startTotal > plugin.Warning) && (endTotal > plugin.Warning) {
		fmt.Printf("%s Warning: %.2f%% memory usage in the beginning and %.2f%% afterwards | %s\n", plugin.PluginConfig.Name, vmStart.UsedPercent, vmEnd.UsedPercent, perfData)
		return sensu.CheckStateWarning, nil
	}

	fmt.Printf("%s OK: %.2f%% memory usage in the beginning and %.2f%% afterwards | %s\n", plugin.PluginConfig.Name, vmStart.UsedPercent, vmEnd.UsedPercent, perfData)
	return sensu.CheckStateOK, nil
}
