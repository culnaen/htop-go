package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	Proc               = "/proc"
	ProcCpuinfoPath    = "/proc/cpuinfo"
	ProcStatPath       = "/proc/stat"
	ProcMeminfoPath    = "/proc/meminfo"
	ProcUptimePath     = "/proc/uptime"
	ProcPIDCPUStatPath = "/proc/%d/stat"
)

type CpuData struct {
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

type MemData struct {
	MemTotal     int
	MemAvailable int
	Buffers      int
	Cached       int
}

type ProcessStat struct {
	PID   int
	Name  string
	Utime int
	Stime int
}

func tryConvertToInt(value string) int {
	if result, err := strconv.Atoi(value); err != nil {
		log.Printf("Error converting to integer: %v", err)
		return 0
	} else {
		return result
	}
}

func openFile(path string) *os.File {
	if file, err := os.Open(path); err != nil {
		log.Printf("Error open file: %v", err)
		return nil
	} else {
		return file
	}
}

func readFile(file *os.File) []byte {
	if bytes, err := io.ReadAll(file); err != nil {
		log.Printf("Error read file: %v", err)
		return nil
	} else {
		return bytes
	}
}

func getCPUName() string {
	var unique []string
	file := openFile(ProcCpuinfoPath)
	defer file.Close()
	bytes := readFile(file)

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

func readCPUData() []CpuData {
	var cpus []CpuData
	file := openFile(ProcStatPath)
	defer file.Close()
	bytes := readFile(file)

	stats := strings.Split(string(bytes), "\n")

	for _, row := range stats {
		if strings.Contains(row, "cpu") {
			data := strings.Fields(row)

			name := data[0]
			nicePeriod := tryConvertToInt(data[2])
			userPeriod := tryConvertToInt(data[1])
			systemPeriod := tryConvertToInt(data[3])
			irqPeriod := tryConvertToInt(data[6])
			softIrqPeriod := tryConvertToInt(data[7])
			stealPeriod := tryConvertToInt(data[8])
			guestPeriod := tryConvertToInt(data[9])
			ioWaitPeriod := tryConvertToInt(data[5])
			idlePeriod := tryConvertToInt(data[4])

			cpuData := CpuData{
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

func calcCPUUsage(prev, curr CpuData) (float32, float32) {
	prevIdle := prev.IdlePeriod + prev.IoWaitPeriod
	idle := curr.IdlePeriod + curr.IoWaitPeriod

	prevNonIdle := prev.UserPeriod + prev.NicePeriod + prev.SystemPeriod + prev.IrqPeriod + prev.SoftIrqPeriod + prev.StealPeriod
	nonIdle := curr.UserPeriod + curr.NicePeriod + curr.SystemPeriod + curr.IrqPeriod + curr.SoftIrqPeriod + curr.StealPeriod
	prevTotal := prevIdle + prevNonIdle
	total := idle + nonIdle

	totald := float32(total - prevTotal)
	idled := float32(idle - prevIdle)

	if totald == 0 {
		totald = 1
	}

	return totald, idled
}

func readMemData() MemData {
	var result MemData
	file := openFile(ProcMeminfoPath)
	bytes := readFile(file)

	data := strings.Split(string(bytes), "\n")
	for _, row := range data {
		if after, ok := strings.CutPrefix(row, "MemTotal:"); ok {
			MemTotal := tryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.MemTotal = MemTotal
		}
		if after, ok := strings.CutPrefix(row, "MemAvailable:"); ok {
			MemAvailable := tryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.MemAvailable = MemAvailable
		}
		if after, ok := strings.CutPrefix(row, "Buffers:"); ok {
			Buffers := tryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.Buffers = Buffers
		}
		if after, ok := strings.CutPrefix(row, "Cached:"); ok {
			Cached := tryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.Cached = Cached
		}
	}
	file.Close()
	return result
}

func calcMemUsage(MemTotal, MemFree, Buffers, Cached int) int {
	return MemTotal - MemFree - Buffers - Cached
}

func readUptimeData() (int, int) {
	file := openFile(ProcUptimePath)
	bytes := readFile(file)

	data := strings.Split(strings.TrimSpace(string(bytes)), " ")

	uptimeSystem, err := strconv.ParseFloat(data[0], 64)
	if err != nil {
		log.Printf("Error parsing uptime: %v", err)
	}
	idleTime, err := strconv.ParseFloat(data[1], 64)
	if err != nil {
		log.Printf("Error parsing idle: %v", err)
	}
	return int(uptimeSystem), int(idleTime)
}

func getProcesses() []int {
	var processes []int
	if dirs, err := os.ReadDir(Proc); err != nil {
		log.Printf("Error reading directory: %v", err)
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

func getProcessStat(pid int) ProcessStat {
	file := openFile(fmt.Sprintf(ProcPIDCPUStatPath, pid))
	data := readFile(file)

	fields := strings.Fields(string(data))

	processStat := ProcessStat{
		PID:   tryConvertToInt(fields[0]),
		Name:  fields[1],
		Utime: tryConvertToInt(fields[13]),
		Stime: tryConvertToInt(fields[14]),
	}

	return processStat
}

func main() {
	for {
		prevCpuData := readCPUData()
		prevProcesses := getProcesses()
		prevCpu := prevCpuData[0]

		prevProcessesWithStat := make(map[int]ProcessStat)
		for _, process := range prevProcesses {
			processStat := getProcessStat(process)
			prevProcessesWithStat[processStat.PID] = processStat
		}
		time.Sleep(1 * time.Second)

		currCpuData := readCPUData()
		currCpu := currCpuData[0]
		currMem := readMemData()
		uptimeSystem, _ := readUptimeData()

		processes := getProcesses()

		var processesWithStat []ProcessStat
		for _, process := range processes {
			processStat := getProcessStat(process)
			processesWithStat = append(processesWithStat, processStat)
		}

		total, _ := calcCPUUsage(prevCpu, currCpu)
		period := total / float32(len(currCpuData)-1)

		fmt.Print("\033[H\033[2J")
		fmt.Printf("CPU: %s\n", getCPUName())
		fmt.Printf("Memory: %.2fG/%.1fG\n", float32(calcMemUsage(currMem.MemTotal, currMem.MemAvailable, currMem.Buffers, currMem.Cached)/1024)*0.001, float32(currMem.MemTotal/1024)*0.001)
		fmt.Printf("Uptime: %v\n", time.Duration(uptimeSystem)*time.Second)
		for n, cpu := range currCpuData[1:] {
			totald, idled := calcCPUUsage(prevCpuData[n+1], cpu)
			fmt.Printf("CPU%d %.1f%%\n", n, (totald-idled)/totald*100)
		}
		fmt.Print("PID    NAME    CPU%    MEM%\n")
		for _, process := range processesWithStat {
			lasttimes := prevProcessesWithStat[process.PID].Utime + prevProcessesWithStat[process.PID].Stime
			currentTimes := process.Utime + process.Stime

			var percentCpu float32
			if currentTimes > lasttimes {
				percentCpu = float32(currentTimes - lasttimes)
			} else {
				percentCpu = 0
			}
			percentCpu = percentCpu / float32(period) * 100.0
			percentCpu = min(percentCpu, float32(len(currCpuData)*100.0))

			fmt.Printf("%d    %s    %.1f%%\n", process.PID, process.Name, percentCpu)
		}
	}
}
