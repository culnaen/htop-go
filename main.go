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
	result, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Error converting to integer: %v", err)
		return 0
	}
	return result
}
func getCPUName() string {
	var unique []string
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Error read file: %v", err)
	}

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
	file, err := os.Open("/proc/stat")

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Error read file: %v", err)
	}

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

			cpu_data := CpuData{
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
			cpus = append(cpus, cpu_data)

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

func readMemData() (MemData, error) {
	var result MemData
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Error read file: %v", err)
	}

	data := strings.Split(string(bytes), "\n")
	for _, row := range data {
		if strings.HasPrefix(row, "MemTotal:") {
			MemTotal := tryConvertToInt(strings.TrimSpace(strings.Trim(strings.TrimPrefix(row, "MemTotal:"), "kB")))
			result.MemTotal = MemTotal
		}
		if strings.HasPrefix(row, "MemFree:") {
			MemFree := tryConvertToInt(strings.TrimSpace(strings.Trim(strings.TrimPrefix(row, "MemFree:"), "kB")))
			result.MemFree = MemFree
		}
	}
	file.Close()
	return result, err
}

func main() {
	for {
		prev := readCPUData()
		time.Sleep(1 * time.Second)
		curr := readCPUData()
		curr_mem, _ := readMemData()

		fmt.Print("\033[H\033[2J")
		fmt.Printf("%s\nMemory: %.2fG/%.1fG\n", getCPUName()[1:], float32((curr_mem.MemTotal-curr_mem.MemFree)/1024)*0.001, float32(curr_mem.MemTotal/1024)*0.001)
		for n, cpu := range curr {
			fmt.Printf("CPU%d %.2f%%\n", n, calcCPUUsage(prev[n], cpu))
		}

	}
}
