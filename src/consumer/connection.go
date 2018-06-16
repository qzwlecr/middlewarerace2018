package consumer

import (
	"net"
	"log"
	"encoding/binary"
	"protocol"
	"io"
)

type connection struct {
	connId    int
	ignoreNum int
	consumer  *Consumer
	provider  *provider
}

func (c *connection) dealWithConnection(conn net.Conn) {
	header := make([]byte, headerMaxSize)
	var lens uint32
	var cprep protocol.CustResponse
	for cpreq := range c.consumer.chanOut {
		cbreq, err := cpreq.ToByteArr()
		if err != nil {
			log.Fatalln(err)
			return
		}

		lens = uint32(len(cbreq))
		binary.BigEndian.PutUint32(header, lens)
		fullp := append(header, cbreq...)

		conn.Write(fullp)

		_, err = io.ReadFull(conn, header)
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
