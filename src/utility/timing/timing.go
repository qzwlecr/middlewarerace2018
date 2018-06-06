package timing

import (
	"log"
	"time"
)

const TIME_LOGGING = true
const TOT_LOGGING = true

// sample usage:
// in some func: defer timing.Since(time.Now(), "LOL")

// Since behaves..

var tottime = make(map[string]time.Duration)

func Since(start time.Time, hint string) {
	if !TIME_LOGGING {
		return
	}
	el := time.Since(start)
	if TOT_LOGGING {
		tm, exist := tottime[hint]
		if !exist {
			tottime[hint] = el
		} else {
			tottime[hint] = tm + el
		}
	}
	if TOT_LOGGING {
		log.Println(hint, "executed", el, "total", tottime[hint])
	} else {
		log.Println(hint, "executed", el)
	}
}
