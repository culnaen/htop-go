package cpu

import (
	"slices"
	"strings"

	"htop-go/internal/files"
	"htop-go/internal/utils"
)

const (
	ProcCPUinfoPath = "/proc/cpuinfo"
	ProcStatPath    = "/proc/stat"
)

type CPUData struct {
	Name          string
	NicePeriod    int
	UserPeriod    int
	SystemPeriod  int
	IrqPeriod     int
	SoftIrqPeriod int
	StealPeriod   int
	GuestPeriod   int
	IoWaitPeriod  int
	IdlePeriod    int
}

func GetCPUName() string {
	var unique []string
	file := files.Open(ProcCPUinfoPath)
	defer file.Close()
	bytes := files.Read(file)

	data := strings.Split(string(bytes), "\n")
	for _, row := range data {
		if strings.Contains(row, "model name") {
			skip := slices.Contains(unique, row)
			if !skip {
				unique = append(unique, row)
			}
		}
	}

	return strings.TrimSpace(strings.Split(unique[0], ":")[1])
}

func GetCPUData() []CPUData {
	var cpus []CPUData
	file := files.Open(ProcStatPath)
	defer file.Close()
	bytes := files.Read(file)

	stats := strings.Split(string(bytes), "\n")

	for _, row := range stats {
		if strings.Contains(row, "cpu") {
			data := strings.Fields(row)

			name := data[0]
			nicePeriod := utils.TryConvertToInt(data[2])
			userPeriod := utils.TryConvertToInt(data[1])
			systemPeriod := utils.TryConvertToInt(data[3])
			irqPeriod := utils.TryConvertToInt(data[6])
			softIrqPeriod := utils.TryConvertToInt(data[7])
			stealPeriod := utils.TryConvertToInt(data[8])
			guestPeriod := utils.TryConvertToInt(data[9])
			ioWaitPeriod := utils.TryConvertToInt(data[5])
			idlePeriod := utils.TryConvertToInt(data[4])

			cpuData := CPUData{
				Name:          name,
				NicePeriod:    nicePeriod,
				UserPeriod:    userPeriod,
				SystemPeriod:  systemPeriod,
				IrqPeriod:     irqPeriod,
				SoftIrqPeriod: softIrqPeriod,
				StealPeriod:   stealPeriod,
				GuestPeriod:   guestPeriod,
				IoWaitPeriod:  ioWaitPeriod,
				IdlePeriod:    idlePeriod,
			}
			cpus = append(cpus, cpuData)

		}
	}

	return cpus
}

func CalcCPUUsage(prev, curr CPUData) (float64, float64) {
	prevIdle := prev.IdlePeriod + prev.IoWaitPeriod
	idle := curr.IdlePeriod + curr.IoWaitPeriod

	prevNonIdle := prev.UserPeriod + prev.NicePeriod + prev.SystemPeriod + prev.IrqPeriod + prev.SoftIrqPeriod + prev.StealPeriod
	nonIdle := curr.UserPeriod + curr.NicePeriod + curr.SystemPeriod + curr.IrqPeriod + curr.SoftIrqPeriod + curr.StealPeriod
	prevTotal := prevIdle + prevNonIdle
	total := idle + nonIdle

	totald := float64(total - prevTotal)
	idled := float64(idle - prevIdle)

	if totald == 0 {
		totald = 1
	}

	return totald, idled
}
