package uptime

import (
	"log"
	"strconv"
	"strings"

	"htop-go/internal/files"
)

const (
	ProcUptimePath = "/proc/uptime"
)

func GetUptimeData() (int, int) {
	file := files.Open(ProcUptimePath)
	bytes := files.Read(file)

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
