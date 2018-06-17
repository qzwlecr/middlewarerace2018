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
	for cpreq := range c.provider.chanRequest {
		c.provider.activeMu.Lock()
		c.provider.active ++
		c.provider.activeMu.Unlock()
		cbreq, err := cpreq.ToByteArr()
		if err != nil {
			log.Fatalln(err)
			continue
		}

		lens = uint32(len(cbreq))
		binary.BigEndian.PutUint32(header, lens)
		fullp := append(header, cbreq...)

		_, err = conn.Write(fullp)
		if err != nil {
			log.Fatalln(err)
			continue
		}

		_, err = io.ReadFull(conn, header)
		if err != nil {
			log.Fatalln(err)
			continue
		}

		lens = binary.BigEndian.Uint32(header)
		body := make([]byte, lens)
		_, err = io.ReadFull(conn, body)
		if err != nil {
			log.Fatalln(err)
			continue
		}

		cprep.FromByteArr(body)
		ans := answer{
			connId: c.connId,
			id:     cprep.Identifier,
			reply:  cprep.Reply,
		}
		c.consumer.chanAnswer <- ans

		c.provider.activeMu.Lock()
		c.provider.active --
		c.provider.activeMu.Unlock()
	}
}
