package rrselector

// RRSelector is a round-robin selector, is used on parallel-biased scenes.
type RRSelector int

//SelectBetween selects between num numbers of elements.
func (sl *RRSelector) SelectBetween(num int) int {
	defer func() {
		if int(*sl) >= num {
			*sl = 0
		} else {
			*sl = *sl + 1
		}
	}()
	return int(*sl) % num
}
