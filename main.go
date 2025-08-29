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
	MemTotal int
	MemFree  int
}

func tryConvertToInt(value string) int {
	if result, err := strconv.Atoi(value); err != nil {
		log.Fatalf("Error converting to integer: %v", err)
		return 0
	} else {
		return result
	}
}

func openFile(path string) *os.File {
	if file, err := os.Open(path); err != nil {
		log.Fatalf("Error open file: %v", err)
		return nil
	} else {
		return file
	}
}

func readFile(file *os.File) []byte {
	if bytes, err := io.ReadAll(file); err != nil {
		log.Fatalf("Error read file: %v", err)
		return nil
	} else {
		return bytes
	}
}

func getCPUName() string {
	var unique []string
	file := openFile("/proc/cpuinfo")
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
	file.Close()
	return strings.Split(unique[0], ":")[1]
}

func readCPUData() []CpuData {
	var cpus []CpuData
	file := openFile("/proc/stat")
	bytes := readFile(file)

	stats := strings.Split(string(bytes), "\n")

	for _, row := range stats[1:] {
		if strings.Contains(row, "cpu") {
			data := strings.Split(row[4:], " ")

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
				Name:          row[:5],
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

	file.Close()
	return cpus
}

func calcCPUUsage(prev, curr CpuData) float64 {
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

	return (totald - idled) / totald * 100
}

func readMemData() MemData {
	var result MemData
	file := openFile("/proc/meminfo")
	bytes := readFile(file)

	data := strings.Split(string(bytes), "\n")
	for _, row := range data {
		if after, ok := strings.CutPrefix(row, "MemTotal:"); ok {
			MemTotal := tryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.MemTotal = MemTotal
		}
		if after, ok := strings.CutPrefix(row, "MemFree:"); ok {
			MemFree := tryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.MemFree = MemFree
		}
	}
	file.Close()
	return result
}

func calcMemUsage(MemTotal, MemFree float32) float32 {
	return MemTotal - MemFree
}

func readUptimeData() (int, int) {
	file := openFile("/proc/uptime")
	bytes := readFile(file)

	data := strings.Split(strings.TrimSpace(string(bytes)), " ")

	uptimeSystem, err := strconv.ParseFloat(data[0], 64)
	if err != nil {
		log.Fatalf("Error parsing uptime: %v", err)
	}
	idleTime, err := strconv.ParseFloat(data[1], 64)
	if err != nil {
		log.Fatalf("Error parsing idle: %v", err)
	}
	return int(uptimeSystem), int(idleTime)
}

func main() {
	for {
		prevCpu := readCPUData()
		time.Sleep(1 * time.Second)
		currCpu := readCPUData()
		currMem := readMemData()
		memTotalFloat := float32(currMem.MemTotal/1024) * 0.001
		memFreeFloat := float32(currMem.MemFree/1024) * 0.001
		uptimeSystem, _ := readUptimeData()

		fmt.Print("\033[H\033[2J")
		fmt.Printf("CPU: %s\n", getCPUName()[1:])
		fmt.Printf("Memory: %.2fG/%.1fG\n", calcMemUsage(memTotalFloat, memFreeFloat), memTotalFloat)
		fmt.Printf("Uptime: %v\n", time.Duration(uptimeSystem)*time.Second)
		for n, cpu := range currCpu {
			fmt.Printf("CPU%d %.2f%%\n", n, calcCPUUsage(prevCpu[n], cpu))
		}

	}
}
