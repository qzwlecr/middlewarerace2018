package consumer

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"protocol"
	"sync"
	"sync/atomic"
	"time"
	"utility/timing"
)

type Connection struct {
	isActive bool
	answer   sync.Map
	consumer *Consumer
	provider *Provider
}

func (connection *Connection) write(conn net.Conn) {
	lb := make([]byte, 4)
	if conn == nil {
		log.Panic("Conn boom in writer!")
	}
	lens := uint32(0)
	var ti time.Time
	for {
		select {
		case ms := <-connection.provider.chanIn:
			if connection.isActive == false {
				connection.isActive = true
				atomic.AddUint32(&connection.provider.active, 1)
			}
			ti = time.Now()
			connection.answer.Store(ms.cr.Identifier, ms.chanAnswer)

			if logger {
				log.Println("ID has been stored:", ms.cr.Identifier)
				if ms.chanAnswer == nil {
					log.Println("Pass para boom!")
				}
			}

			cbreq, err := ms.cr.ToByteArr()
			if err != nil {
				log.Fatal(err)
				return
			}

			lens = uint32(len(cbreq))
			binary.BigEndian.PutUint32(lb, lens)
			fullp := append(lb, cbreq...)

			if logger {
				log.Println("Write Packages:", fullp)
			}

			conn.Write(fullp)
			timing.Since(ti, "[INFO]Writing: ")
		case <-time.Tick(checkTimeout):
			if connection.isActive == true {
				connection.isActive = false
				atomic.AddUint32(&connection.provider.active, ^uint32(0))
			}
		}
	}
}

func (connection *Connection) read(conn net.Conn) {
	lb := make([]byte, 4)
	if conn == nil {
		log.Panic("Conn boom in reader!")
	}
	for {
		n, err := io.ReadFull(conn, lb)
		if n != 4 || err != nil {
			log.Fatal(err)
			return
		}
		ti := time.Now()
		lens := binary.BigEndian.Uint32(lb)
		cbrep := make([]byte, lens)
		n, err = io.ReadFull(conn, cbrep)

		if n != int(lens) || err != nil {
			log.Fatal(err)
			return
		}
		var cprep protocol.CustResponse
		cprep.FromByteArr(cbrep)

		if logger {
			log.Println("Read Packages:", cprep)
		}

		go func(cprep protocol.CustResponse) {
			ch, ok := connection.answer.Load(cprep.Identifier)
			if ok != true {
				log.Panic( "Can't get the channel with ID:", cprep.Identifier)
			}

			ch.(chan []byte) <- cprep.Reply
			connection.answer.Delete(cprep.Identifier)
		}(cprep)

		connection.provider.delay = (oldWeight*connection.provider.delay + newWeight*cprep.Delay) / 10
		if logger {
			log.Println("Get reply: Lantency = ", cprep.Delay)
		}
		timing.Since(ti, "[INFO]Reading: ")
	}
}
