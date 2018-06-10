package consumer

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"protocol"
	"sync/atomic"
	"time"
	"utility/timing"
)

type Connection struct {
	//isActive bool
	consumer *Consumer
	provider *Provider
}

func (connection *Connection) write(conn net.Conn) {
	lb := make([]byte, 4)
	lens := uint32(0)
	var ti time.Time
	for {
		select {
		case cpreq := <-connection.provider.chanIn:
			//if connection.isActive == false {
			//	connection.isActive = true
			//	atomic.AddUint32(&connection.provider.active, 1)
			//}
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

			conn.Write(fullp)
			timing.Since(ti, "[INFO]Writing: ")
			//case <-time.Tick(checkTimeout):
			//if connection.isActive == true {
			//	connection.isActive = false
			//	atomic.AddUint32(&connection.provider.active, ^uint32(0))
			//}
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
			log.Println("Read Packages:", cbrep)
		}
		ch, _ := connection.consumer.answer.Load(cprep.Identifier)
		go func(ch chan []byte, cprep protocol.CustResponse) {
			ch <- cprep.Reply
		}(ch.(chan []byte), cprep)
		active, _ := connection.provider.idQueueMap.Load(cprep.Identifier)
		index := active.(int) / activeDiv
		if index < len(connection.provider.delay) {
			oldDelay := connection.provider.delay[index]
			// connection.provider.delay[index] = (oldWeight*oldDelay + newWeight*cprep.Delay) / 10
			connection.provider.delay[index] = (oldDelay + cprep.Delay) / 2
		} else {
			connection.provider.delay = append(connection.provider.delay, cprep.Delay)
		}
		connection.provider.idQueueMap.Delete(cprep.Identifier)
		atomic.AddUint32(&connection.provider.activeCnt, ^uint32(0))
		if logger {
			log.Println("Get reply: Lantency = ", cprep.Delay)
		}
		timing.Since(ti, "[INFO]Reading: ")
	}
}
