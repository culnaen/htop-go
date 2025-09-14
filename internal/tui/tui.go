package tui

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"htop-go/internal/cpu"
	"htop-go/internal/memory"
	"htop-go/internal/process"
	"htop-go/internal/uptime"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	CharsPerLine = 65
)

func Init() {

	pageSize := memory.GetPageSize()

	var CPUDataString strings.Builder

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	// cpu usage
	prevCPUData := cpu.GetCPUData()
	prevCPU := prevCPUData[0]

	// cpu data render
	cpuDataWidget := widgets.NewParagraph()
	cpuDataWidget.SetRect(0, 6, 65, 15)

	// memory usage
	currentMemory := memory.GetMemData()

	// memory render
	memoryWidget := widgets.NewParagraph()
	memoryWidget.Title = "Memory"
	memoryWidget.Text = fmt.Sprintf(
		"%.2fG/%.1fG",
		memory.CalcMemUsage(currentMemory),
		float32(currentMemory.MemTotal/1024)*0.001,
	)
	memoryWidget.SetRect(0, 3, 65, 6)

	// CPUName render
	CPUName := widgets.NewParagraph()
	CPUName.Title = "CPU"
	CPUName.Text = cpu.GetCPUName()
	CPUName.SetRect(0, 0, 65, 3)

	// Uptime render
	uptimeWidget := widgets.NewParagraph()
	uptimeWidget.Title = "Uptime"
	uptimeData, _ := uptime.GetUptimeData()
	uptimeWidget.Text = (time.Duration(uptimeData) * time.Second).String()
	uptimeWidget.SetRect(50, 0, 65, 3)

	// processes render
	processes := widgets.NewTable()
	processes.Rows = [][]string{
		{"PID", "CPU%", "Mem", "NAME"},
	}
	processes.TextStyle = ui.NewStyle(ui.ColorWhite)
	processes.SetRect(0, 15, 65, 60)
	processes.RowSeparator = true
	processes.BorderStyle = ui.NewStyle(ui.ColorGreen)
	processes.FillRow = true
	processes.RowStyles[0] = ui.NewStyle(ui.ColorWhite)

	// processes statistics
	prevProcesses := process.Get()
	prevProcessesStats := make(map[int]process.Stat)
	for _, p := range prevProcesses {
		processStat := process.GetStat(p)
		prevProcessesStats[processStat.PID] = processStat
	}

	// render
	ui.Render(CPUName, uptimeWidget, memoryWidget, processes)

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
			currentCPUData := cpu.GetCPUData()
			currentCPU := currentCPUData[0]

			totalPeriod, _ := cpu.CalcCPUUsage(prevCPU, currentCPU)
			periodPerCore := totalPeriod / float64(len(currentCPUData)-1)

			prevCPU = currentCPU

			//memory
			currentMemory = memory.GetMemData()
			memoryWidget.Text = fmt.Sprintf(
				"%.2fG/%.1fG",
				memory.CalcMemUsage(currentMemory),
				float32(currentMemory.MemTotal/1024)*0.001,
			)

			// processes statistics
			currentProcesses := process.Get()

			currentProcessesStats := make(map[int]process.Stat)
			for _, processStat := range currentProcesses {
				processStat := process.GetStat(processStat)
				currentProcessesStats[processStat.PID] = processStat
			}

			tmpProcesses := []process.Stat{}

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
					tmpProcesses = append(tmpProcesses, process.Stat{PID: pid, Name: processStat.Name, Utime: 0, Stime: 0, CPUUsage: 0.0})
				}
			}

			sort.Slice(tmpProcesses, func(i, j int) bool {
				return tmpProcesses[i].CPUUsage > tmpProcesses[j].CPUUsage
			})

			clear(prevProcessesStats)
			clear(processes.Rows)
			// processes stat cpu + mem
			processes.Rows = [][]string{
				{"PID", "CPU%", "Mem%", "NAME"},
			}

			for _, tmpP := range tmpProcesses {

				procStatMemory := process.GetStatMemory(tmpP.PID)
				procMemoryUsage := float64(procStatMemory.Resident) / pageSize / float64(currentMemory.MemTotal) * 100

				processes.Rows = append(processes.Rows, []string{
					strconv.Itoa(tmpP.PID),
					strconv.FormatFloat(tmpP.CPUUsage, 'f', 1, 32),
					strconv.FormatFloat(procMemoryUsage, 'f', 1, 32),
					tmpP.Name,
				})
				prevProcessesStats[tmpP.PID] = tmpP
			}

			// uptime
			uptimeData, _ := uptime.GetUptimeData()
			uptimeWidget.Text = (time.Duration(uptimeData) * time.Second).String()

			// cpu data
			strTmp := ""
			for n, c := range currentCPUData[1:] {
				totald, idled := cpu.CalcCPUUsage(prevCPUData[n+1], c)
				cpuusage := strconv.FormatFloat((totald-idled)/totald*100, 'f', 1, 32)
				strTmp += "CPU" + strconv.Itoa(n) + " " + cpuusage + "%   "

			}
			for i, r := range strTmp {
				CPUDataString.WriteRune(r)
				if (i+i)%CharsPerLine == 0 && i != len(strTmp)-1 {
					CPUDataString.WriteString("\n")
				}
			}
			prevCPUData = currentCPUData
			cpuDataWidget.Text = strTmp

			ui.Render(CPUName, cpuDataWidget, uptimeWidget, memoryWidget, processes)
		}
	}
}
