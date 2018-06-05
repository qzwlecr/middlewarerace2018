package timing

import (
	"log"
	"time"
)

const TIME_LOGGING = true

// sample usage:
// in some func: defer timing.Since(time.Now(), "LOL")

// Since behaves..
func Since(start time.Time, hint string) {
	if !TIME_LOGGING {
		return
	}
	el := time.Since(start)
	log.Println(hint, "executed", el)
}
