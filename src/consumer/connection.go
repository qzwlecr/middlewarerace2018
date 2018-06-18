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
	readBuf   []byte
	writeBuf  []byte
}

func (c *connection) readFromProvider(conn net.Conn) {
	var lens uint32
	var cprep protocol.CustResponse
	header := make([]byte, headerMaxSize)
	c.readBuf = make([]byte, bodyMaxSize)
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
		if int(lens) > len(c.readBuf) {
			c.readBuf = make([]byte, lens)
		}
		_, err = io.ReadFull(conn, c.readBuf[:lens])
		if err != nil {
			log.Fatalln(err)
			return
		}
		cprep.FromByteArr(c.readBuf[:lens])
		//log.Println(cprep.Identifier, time.Now().UnixNano()/int64(time.Millisecond), "Recv from ProvAgnt Get")
		ans := answer{
			connId: c.connId,
			id:     cprep.Identifier,
			reply:  cprep.Reply,
		}
		c.consumer.chanIn <- ans
	}
}

func (c *connection) writeToProvider(conn net.Conn) {
	var lens uint32
	header := make([]byte, headerMaxSize)
	for cpreq := range c.provider.chanOut {
		cbreq, err := cpreq.ToByteArr()
		if err != nil {
			log.Fatalln(err)
			return
		}

		lens = uint32(len(cbreq))
		binary.BigEndian.PutUint32(header, lens)
		fullp := append(header, cbreq...)

		go func(cpreq protocol.CustRequest) {
			conn.Write(fullp)
			//log.Println(cpreq.Identifier, time.Now().UnixNano()/int64(time.Millisecond), "Send to ProvAgnt Complete")
		}(cpreq)
	}

}
