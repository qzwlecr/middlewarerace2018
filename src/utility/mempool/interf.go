package mempool

import (
	"log"
)

// MemoryPool : for faster buffer allocation
type MemoryPool interface {
	Acquire() (id uint32, arr []byte, err error)
	Release(id uint32) (err error)
	Attempt(id uint32) (arr []byte, err error)
	Share(id uint32) (arr []byte, err error)
}

const LOGGING = true
const FORCE_ASSERTION = true

func assert(cond bool, hinter func() error) (err error) {
	if FORCE_ASSERTION {
		if !cond {
			err = hinter()
			panic(err)
		}
		return nil
	} else if LOGGING {
		if !cond {
			err = hinter()
			log.Fatalln(err)
			return err
		}
		return nil
	} else if !cond {
		return hinter()
	}
	return nil
}
