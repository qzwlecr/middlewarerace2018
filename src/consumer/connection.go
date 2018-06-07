package consumer

import (
	"time"
	"encoding/binary"
	"io"
	"log"
	"protocol"
	"utility/timing"
	"net"
	"sync/atomic"
)

type Connection struct {
	conn     net.Conn
	isActive bool
	consumer *Consumer
	provider *Provider
}

func (connection Connection) write() {
	lb := make([]byte, 4)
	lens := uint32(0)
	var ti time.Time
	for {
		select {
		case cpreq := <-connection.provider.chanIn:
			if connection.isActive == false {
				connection.isActive = true
				atomic.AddUint32(&connection.provider.active, 1)
			}
			ti = time.Now()
			cbreq, err := cpreq.ToByteArr()
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

			connection.conn.Write(fullp)
			timing.Since(ti, "[INFO]Writing: ")
		case <-time.Tick(checkTimeout):
			if connection.isActive == true {
				connection.isActive = false
				atomic.AddUint32(&connection.provider.active, ^uint32(0))
			}
		}
	}
}

func (connection Connection) read() {
	lb := make([]byte, 4)
	if connection.conn == nil {
		log.Panic("Conn boom in reader!")
	}
	for {
		n, err := io.ReadFull(connection.conn, lb)
		if n != 4 || err != nil {
			log.Fatal(err)
			return
		}
		ti := time.Now()
		lens := binary.BigEndian.Uint32(lb)
		cbrep := make([]byte, lens)
		n, err = io.ReadFull(connection.conn, cbrep)

		if n != int(lens) || err != nil {
			log.Fatal(err)
			return
		}
		var cprep protocol.CustResponse
		cprep.FromByteArr(cbrep)
		if logger {
			log.Println("Read Packages:", cbrep)
		}
		ch, _ := connection.consumer.answer.Load(cprep.Identifier)
		go func() { ch.(chan []byte) <- cprep.Reply }()
		connection.provider.delay = (connection.provider.delay + cprep.Delay) / 2
		timing.Since(ti, "[INFO]Reading: ")
	}
}
