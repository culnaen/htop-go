package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	Proc               = "/proc"
	ProcCPUinfoPath    = "/proc/cpuinfo"
	ProcStatPath       = "/proc/stat"
	ProcMeminfoPath    = "/proc/meminfo"
	ProcUptimePath     = "/proc/uptime"
	ProcPIDCPUStatPath = "/proc/%d/stat"
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

type MemData struct {
	MemTotal     int
	MemAvailable int
	Buffers      int
	Cached       int
}

type ProcessStat struct {
	PID      int
	Name     string
	Utime    int
	Stime    int
	CPUUsage float64
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
	file := openFile(ProcCPUinfoPath)
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

func readCPUData() []CPUData {
	var cpus []CPUData
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

func calcCPUUsage(prev, curr CPUData) (float64, float64) {
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
	charsPerLine := 65
	var CPUDataString strings.Builder

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// cpu usage
	prevCPUData := readCPUData()
	prevCPU := prevCPUData[0]

	// cpu data render
	cpuDataW := widgets.NewParagraph()
	cpuDataW.SetRect(0, 6, 65, 15)

	// memory usage
	currentMemory := readMemData()

	// memory render
	memory := widgets.NewParagraph()
	memory.Title = "Memory"
	memory.Text = fmt.Sprintf("%.2fG/%.1fG", float32(calcMemUsage(currentMemory.MemTotal, currentMemory.MemAvailable, currentMemory.Buffers, currentMemory.Cached)/1024)*0.001, float32(currentMemory.MemTotal/1024)*0.001)
	memory.SetRect(0, 3, 65, 6)

	// CPUName render
	CPUName := widgets.NewParagraph()
	CPUName.Title = "CPU"
	CPUName.Text = getCPUName()
	CPUName.SetRect(0, 0, 65, 3)

	// Uptime render
	uptimew := widgets.NewParagraph()
	uptimew.Title = "Uptime"
	uptime, _ := readUptimeData()
	uptimew.Text = (time.Duration(uptime) * time.Second).String()
	uptimew.SetRect(50, 0, 65, 3)

	// processes render
	processes := widgets.NewTable()
	processes.Rows = [][]string{
		{"PID", "CPU%", "NAME"},
	}
	processes.TextStyle = ui.NewStyle(ui.ColorWhite)
	processes.SetRect(0, 15, 65, 60)
	processes.RowSeparator = true
	processes.BorderStyle = ui.NewStyle(ui.ColorGreen)
	processes.FillRow = true
	processes.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)

	// processes statistics
	prevProcesses := getProcesses()
	prevProcessesStats := make(map[int]ProcessStat)
	for _, process := range prevProcesses {
		processStat := getProcessStat(process)
		prevProcessesStats[processStat.PID] = processStat
	}

	// render
	ui.Render(CPUName, uptimew, memory, processes)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		time.Sleep(1 * time.Second)
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		case <-ticker.C:

			// cpu usage
			currentCPUData := readCPUData()
			currentCPU := currentCPUData[0]

			totalPeriod, _ := calcCPUUsage(prevCPU, currentCPU)
			periodPerCore := totalPeriod / float64(len(currentCPUData)-1)

			prevCPU = currentCPU

			// processes statistics
			currentProcesses := getProcesses()

			currentProcessesStats := make(map[int]ProcessStat)
			for _, processStat := range currentProcesses {
				processStat := getProcessStat(processStat)
				currentProcessesStats[processStat.PID] = processStat
			}

			tmpProcesses := []ProcessStat{}

			for pid, processStat := range currentProcessesStats {
				_, exists := prevProcessesStats[pid]
				if exists {
					lasttimes := prevProcessesStats[pid].Utime + prevProcessesStats[pid].Stime
					currentTimes := processStat.Utime + processStat.Stime

					var percentCPU float64
					if currentTimes > lasttimes {
						percentCPU = float64(currentTimes - lasttimes)
					} else {
						percentCPU = 0
					}
					percentCPU = percentCPU / float64(periodPerCore) * 100.0
					percentCPU = min(percentCPU, float64(len(currentCPUData)*100.0))
					processStat.CPUUsage = percentCPU

					tmpProcesses = append(tmpProcesses, processStat)
				} else {
					tmpProcesses = append(tmpProcesses, ProcessStat{PID: pid, Name: processStat.Name, Utime: 0, Stime: 0, CPUUsage: 0.0})
				}
			}

			sort.Slice(tmpProcesses, func(i, j int) bool {
				return tmpProcesses[i].CPUUsage > tmpProcesses[j].CPUUsage
			})

			clear(prevProcessesStats)
			clear(processes.Rows)
			processes.Rows = [][]string{
				{"PID", "CPU%", "NAME"},
			}
			for _, tmpP := range tmpProcesses {
				processes.Rows = append(processes.Rows, []string{
					strconv.Itoa(tmpP.PID),
					strconv.FormatFloat(tmpP.CPUUsage, 'f', 1, 32),
					tmpP.Name,
				})
				prevProcessesStats[tmpP.PID] = tmpP
			}

			// uptime
			uptime, _ := readUptimeData()
			uptimew.Text = (time.Duration(uptime) * time.Second).String()

			//memory
			currentMemory = readMemData()
			memory.Text = fmt.Sprintf("%.2fG/%.1fG", float32(calcMemUsage(currentMemory.MemTotal, currentMemory.MemAvailable, currentMemory.Buffers, currentMemory.Cached)/1024)*0.001, float32(currentMemory.MemTotal/1024)*0.001)

			// cpu data
			strTmp := ""
			for n, cpu := range currentCPUData[1:] {
				totald, idled := calcCPUUsage(prevCPUData[n+1], cpu)
				cpuusage := strconv.FormatFloat((totald-idled)/totald*100, 'f', 1, 32)
				strTmp += "CPU" + strconv.Itoa(n) + " " + cpuusage + "%   "

			}
			for i, r := range strTmp {
				CPUDataString.WriteRune(r)
				if (i+i)%charsPerLine == 0 && i != len(strTmp)-1 {
					CPUDataString.WriteString("\n")
				}
			}
			prevCPUData = currentCPUData
			cpuDataW.Text = strTmp

			ui.Render(CPUName, cpuDataW, uptimew, memory, processes)
		}
	}
}
