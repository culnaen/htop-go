package memory

import (
	"strings"
	"syscall"

	"htop-go/internal/files"
	"htop-go/internal/utils"
)

const (
	ProcMeminfoPath = "/proc/meminfo"
	OneK            = 1024
)

type MemData struct {
	MemTotal     int
	MemFree      int
	MemAvailable int
	Buffers      int
	Cached       int
	SReclaimable int
	Shmem        int
}

func GetMemData() MemData {
	var result MemData
	file := files.Open(ProcMeminfoPath)
	bytes := files.Read(file)

	data := strings.Split(string(bytes), "\n")
	for _, row := range data {
		if after, ok := strings.CutPrefix(row, "MemTotal:"); ok {
			MemTotal := utils.TryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.MemTotal = MemTotal
		}
		if after, ok := strings.CutPrefix(row, "MemFree:"); ok {
			MemFree := utils.TryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.MemFree = MemFree
		}
		if after, ok := strings.CutPrefix(row, "MemAvailable:"); ok {
			MemAvailable := utils.TryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.MemAvailable = MemAvailable
		}
		if after, ok := strings.CutPrefix(row, "Buffers:"); ok {
			Buffers := utils.TryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.Buffers = Buffers
		}
		if after, ok := strings.CutPrefix(row, "Cached:"); ok {
			Cached := utils.TryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.Cached = Cached
		}
		if after, ok := strings.CutPrefix(row, "SReclaimable:"); ok {
			SReclaimable := utils.TryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.SReclaimable = SReclaimable
		}
		if after, ok := strings.CutPrefix(row, "Shmem:"); ok {
			Shmem := utils.TryConvertToInt(strings.TrimSpace(strings.Trim(after, "kB")))
			result.Shmem = Shmem
		}
	}
	file.Close()
	return result
}

func CalcMemUsage(data MemData) float32 {
	var used int

	cached := data.Cached + data.SReclaimable - data.Shmem
	usedDiff := data.MemFree + cached + data.SReclaimable + data.Buffers

	if data.MemTotal >= usedDiff {
		used = data.MemTotal - usedDiff
	} else {
		used = data.MemTotal - data.MemFree
	}

	return float32(used) / 1024 * 0.001
}

func GetPageSize() float64 {
	return float64(syscall.Getpagesize() / OneK)
}
