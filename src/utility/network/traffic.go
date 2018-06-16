package network

import (
	"errors"
	"log"
	"time"

	psutilNet "github.com/shirou/gopsutil/net"
)

const timeGap = time.Duration(500) * time.Millisecond

// Handler is a descriptor for oop
type Handler struct {
	lastRecv      uint64
	lastTrans     uint64
	lastTimePoint time.Time
	speedMap      map[string][2]float64
	statMap       map[string]psutilNet.IOCountersStat
	stop          chan int
}

// StartMonitor starts monitoring network traffic and initialize a descriptor
func StartMonitor(desc *Handler) {
	desc.stop = make(chan int, 1)
	desc.speedMap = make(map[string][2]float64)
	desc.statMap = make(map[string]psutilNet.IOCountersStat)
	go monitor(desc)
}

// StopMonitor stops the monitoring
func StopMonitor(desc *Handler) {
	desc.stop <- 0
}

// GetSpeed get the network speed
func GetSpeed(desc *Handler, inter string) (float64, float64, error) {
	spds, ok := desc.speedMap[inter]
	if !ok {
		return 0, 0, errors.New("no such interface found")
	}
	return spds[0], spds[1], nil
}

func monitor(desc *Handler) {
	for {
		select {
		case <-desc.stop:
			return
		default:
			stat, err := psutilNet.IOCounters(true)
			if err != nil {
				log.Println("error reading network info: ", err)
				break
			}

			t := time.Now()
			elapsed := t.Sub(desc.lastTimePoint)
			for i := range stat {
				inter := stat[i].Name
				desc.statMap[inter] = stat[i]
				sendSpd := float64(stat[i].BytesSent-desc.statMap[inter].BytesSent) / elapsed.Seconds()
				recvSpd := float64(stat[i].BytesRecv-desc.statMap[inter].BytesRecv) / elapsed.Seconds()
				desc.speedMap[inter] = [2]float64{sendSpd, recvSpd}
			}

			desc.lastTimePoint = t
		}
		time.Sleep(timeGap)
	}
}
