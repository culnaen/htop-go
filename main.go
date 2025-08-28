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

func readCPUData() ([]CpuData, error) {
	var cpus []CpuData
	file, err := os.Open("/proc/stat")

	if err != nil {
		fmt.Println(err)
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

			nicePeriod, err := strconv.Atoi(data[2])
			if err != nil {
				log.Printf("Error converting NicePeriod: %v", err)
				nicePeriod = 0
			}
			userPeriod, err := strconv.Atoi(data[1])
			if err != nil {
				log.Printf("Error converting UserPeriod: %v", err)
				userPeriod = 0
			}
			systemPeriod, err := strconv.Atoi(data[3])
			if err != nil {
				log.Printf("Error converting SystemPeriod: %v", err)
				systemPeriod = 0
			}
			irqPeriod, err := strconv.Atoi(data[6])
			if err != nil {
				log.Printf("Error converting IrqPeriod: %v", err)
				irqPeriod = 0
			}
			softIrqPeriod, err := strconv.Atoi(data[7])
			if err != nil {
				log.Printf("Error converting SoftIrqPeriod: %v", err)
				softIrqPeriod = 0
			}
			stealPeriod, err := strconv.Atoi(data[8])
			if err != nil {
				log.Printf("Error converting StealPeriod: %v", err)
				stealPeriod = 0
			}
			guestPeriod, err := strconv.Atoi(data[9])
			if err != nil {
				log.Printf("Error converting GuestPeriod: %v", err)
				guestPeriod = 0
			}
			ioWaitPeriod, err := strconv.Atoi(data[5])
			if err != nil {
				log.Printf("Error converting IoWaitPeriod: %v", err)
				ioWaitPeriod = 0
			}
			idlePeriod, err := strconv.Atoi(data[4])
			if err != nil {
				log.Printf("Error converting idlePeriod: %v", err)
				idlePeriod = 0
			}

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
	return cpus, fmt.Errorf("no CPU data found")
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
			MemTotal, err := strconv.Atoi(strings.TrimSpace(strings.Trim(strings.TrimPrefix(row, "MemTotal:"), "kB")))
			if err != nil {
				log.Fatal(err)
			}
			result.MemTotal = MemTotal
		}
		if strings.HasPrefix(row, "MemFree:") {
			MemFree, err := strconv.Atoi(strings.TrimSpace(strings.Trim(strings.TrimPrefix(row, "MemFree:"), "kB")))
			if err != nil {
				log.Fatal(err)
			}
			result.MemFree = MemFree
		}
	}
	file.Close()
	return result, err
}

func main() {
	for {
		prev, _ := readCPUData()
		time.Sleep(1 * time.Second)
		curr, _ := readCPUData()
		curr_mem, _ := readMemData()

		fmt.Print("\033[H\033[2J")
		fmt.Printf("%s\nMemory: %.2fG/%.1fG\n", getCPUName()[1:], float32((curr_mem.MemTotal-curr_mem.MemFree)/1024)*0.001, float32(curr_mem.MemTotal/1024)*0.001)
		for n, cpu := range curr {
			fmt.Printf("CPU%d %.2f%%\n", n, calcCPUUsage(prev[n], cpu))
		}

	}
}
