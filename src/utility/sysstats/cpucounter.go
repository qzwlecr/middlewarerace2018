package sysstats

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

// Interval is the calculation interval.
const Interval = 1000

// CPULoad is the load!
type CPULoad int

// This will return the int.
func (cpuload *CPULoad) This() int {
	return int(*cpuload)
}

// Monitor will automatically change the refered load, if you go it out.
func (cpuload *CPULoad) Monitor() {
	for {
		ibegin, tbegin := getCPUSample()
		<-time.After(Interval * time.Millisecond)
		iend, tend := getCPUSample()
		idleTicks := float64(iend - ibegin)
		totalTicks := float64(tend - tbegin)
		cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
		*cpuload = CPULoad(cpuUsage)
	}
}

func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}
