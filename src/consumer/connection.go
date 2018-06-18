package consumer

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"protocol"
)

type connection struct {
	connId    int
	ignoreNum int
	consumer  *Consumer
	provider  *provider
}

func (c *connection) readFromProvider(conn net.Conn) {
	header := make([]byte, headerMaxSize)
	//body := make([]byte, bodyMaxSize)
	var lens uint32
	var cprep protocol.CustResponse
	if conn == nil {
		log.Panic("[PANIC]connection is nil when reading from provider!")
	}
	for {
		_, err := io.ReadFull(conn, header)
		if err != nil {
			log.Fatalln(err)
			return
		}

		lens = binary.BigEndian.Uint32(header)
		body := make([]byte, lens)
		_, err = io.ReadFull(conn, body)
		if err != nil {
			log.Fatalln(err)
			return
		}

		cprep.FromByteArr(body)
		ans := answer{
			connId: c.connId,
			id:     cprep.Identifier,
			reply:  cprep.Reply,
		}
		c.consumer.chanIn <- ans
	}
}

func (c *connection) writeToProvider(conn net.Conn) {
	header := make([]byte, headerMaxSize)
	var lens uint32
	for cpreq := range c.provider.chanOut {
		cbreq, err := cpreq.ToByteArr()
		if err != nil {
			log.Fatalln(err)
			return
		}

		lens = uint32(len(cbreq))
		binary.BigEndian.PutUint32(header, lens)
		fullp := append(header, cbreq...)

		go conn.Write(fullp)
	}

}
