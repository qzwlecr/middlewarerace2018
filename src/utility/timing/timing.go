package timing

import (
	"log"
	"sync"
	"time"
)

const TIME_LOGGING = false
const TOT_LOGGING = true

// sample usage:
// in some func: defer timing.Since(time.Now(), "LOL")

// Since behaves..

var tottime = make(map[string]time.Duration)
var mu sync.Mutex

func Since(start time.Time, hint string) {
	if !TIME_LOGGING {
		return
	}
	el := time.Since(start)
	if TOT_LOGGING {
		mu.Lock()
		tm, exist := tottime[hint]
		if !exist {
			tottime[hint] = el
		} else {
			tottime[hint] = tm + el
		}
		mu.Unlock()
	}
	if TOT_LOGGING {
		mu.Lock()
		log.Println(hint, "executed", el, "total", tottime[hint])
		mu.Unlock()
	} else {
		log.Println(hint, "executed", el)
	}
}
