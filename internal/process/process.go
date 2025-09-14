package process

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"htop-go/internal/files"
	"htop-go/internal/utils"
)

const (
	Proc                = "/proc"
	ProcPIDCPUStatPath  = "/proc/%d/stat"
	ProcPIDCPUStatMPath = "/proc/%d/statm"
)

type Stat struct {
	PID      int
	Name     string
	Utime    int
	Stime    int
	CPUUsage float64
}

type StatMemory struct {
	Size     int // VmSize
	Resident int //VmRSS
	Shared   int // RssFile+RssShmem
	Text     int
	Lib      int // 0
	Data     int
	Dt       int // 0
}

func Get() []int {
	var processes []int
	if dirs, err := os.ReadDir(Proc); err != nil {
		log.Printf("Error Reading directory: %v", err)
	} else {
		for _, dir := range dirs {
			if result, err := strconv.Atoi(dir.Name()); err != nil {
				// pass
			} else {
				processes = append(processes, result)
			}
		}
	}
	return processes
}

func GetStat(pid int) Stat {
	file := files.Open(fmt.Sprintf(ProcPIDCPUStatPath, pid))
	data := files.Read(file)

	fields := strings.Fields(string(data))

	processStat := Stat{
		PID:   utils.TryConvertToInt(fields[0]),
		Name:  fields[1],
		Utime: utils.TryConvertToInt(fields[13]),
		Stime: utils.TryConvertToInt(fields[14]),
	}

	return processStat
}

func GetStatMemory(pid int) StatMemory {
	var result StatMemory
	file := files.Open(fmt.Sprintf(ProcPIDCPUStatMPath, pid))
	bytes := files.Read(file)

	fields := strings.Fields(string(bytes))

	result = StatMemory{
		Size:     utils.TryConvertToInt(fields[0]),
		Resident: utils.TryConvertToInt(fields[1]),
		Shared:   utils.TryConvertToInt(fields[2]),
		Text:     utils.TryConvertToInt(fields[3]),
		Lib:      utils.TryConvertToInt(fields[4]),
		Data:     utils.TryConvertToInt(fields[5]),
		Dt:       utils.TryConvertToInt(fields[6]),
	}
	return result
}
