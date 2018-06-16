package network

import (
	"errors"
	"log"
	"time"

	psutilNet "github.com/shirou/gopsutil/net"
)

const timeGap = time.Duration(500) * time.Millisecond

// Monitor is a descriptor for oop
type Monitor struct {
	lastTimePoint time.Time
	speedMap      map[string][2]float64
	statMap       map[string]psutilNet.IOCountersStat
	stop          chan int
}

// NewMonitor returns a new Monitor
func NewMonitor() *Monitor {
	return &Monitor{
		stop:     make(chan int, 1),
		speedMap: make(map[string][2]float64),
		statMap:  make(map[string]psutilNet.IOCountersStat),
	}
}

// StartMonitor starts monitoring network traffic and initialize a descriptor
func (desc *Monitor) StartMonitor() error {
	stat, err := psutilNet.IOCounters(true)
	if err != nil {
		log.Println("error reading network info: ", err)
		return err
	}

	t := time.Now()
	for i := range stat {
		inter := stat[i].Name
		desc.statMap[inter] = stat[i]
	}

	desc.lastTimePoint = t

	go desc.monitor()
	return nil
}

// StopMonitor stops the monitoring
func (desc *Monitor) StopMonitor() {
	desc.stop <- 0
}

// GetSpeed gets the network speed
func (desc *Monitor) GetSpeed(inter string) (float64, float64, error) {
	spds, ok := desc.speedMap[inter]
	if !ok {
		return 0, 0, errors.New("no such interface found")
	}
	return spds[0], spds[1], nil
}

func (desc *Monitor) monitor() {
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
				sendSpd := float64(stat[i].BytesSent-desc.statMap[inter].BytesSent) / elapsed.Seconds()
				recvSpd := float64(stat[i].BytesRecv-desc.statMap[inter].BytesRecv) / elapsed.Seconds()
				desc.statMap[inter] = stat[i]
				desc.speedMap[inter] = [2]float64{sendSpd, recvSpd}
			}

			desc.lastTimePoint = t
		}
		time.Sleep(timeGap)
	}
}
