package timing

import (
	"fmt"
	"time"
)

// Since behaves..
func Since(start time.Time, hint string) {
	el := time.Since(start)
	fmt.Println(hint, "executed", el)
}
