package consumer

import (
	"io"
	"log"
	"net"
	"encoding/binary"
	"protocol"
	"time"
)

type connection struct {
	connId int
	//ignoreNum int
	consumer *Consumer
	provider *provider
}

func (c *connection) readFromProvider(conn net.Conn) {
	header := make([]byte, headerMaxSize)
	body := make([]byte, bodyMaxSize)
	var lens uint32
	var cprep protocol.CustResponse
	var ans answer
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
		_, err = io.ReadFull(conn, body[:lens])
		if err != nil {
			log.Fatalln(err)
			return
		}

		cprep.FromByteArr(body[:lens])

		log.Println(time.Now().UnixNano()/int64(time.Millisecond), ":", cprep.Identifier, " got response from provider")
		ans.connId = c.connId
		ans.id = cprep.Identifier
		ans.reply = cprep.Reply
		go func(ans answer) {
			log.Println(time.Now().UnixNano()/int64(time.Millisecond), ":", cprep.Identifier, " write response to answer")
			c.consumer.chanAnswer <- ans
			log.Println(time.Now().UnixNano()/int64(time.Millisecond), ":", cprep.Identifier, " write response to answer done")
		}(ans)
	}
}

func (c *connection) writeToProvider(conn net.Conn) {
	header := make([]byte, headerMaxSize)
	var lens uint32
	for cpreq := range c.consumer.chanRequest {
		cbreq, err := cpreq.ToByteArr()
		if err != nil {
			log.Fatalln(err)
			return
		}

		lens = uint32(len(cbreq))
		binary.BigEndian.PutUint32(header, lens)
		fullp := append(header, cbreq...)
		log.Println(time.Now().UnixNano()/int64(time.Millisecond), ":", cpreq.Identifier, " write request to provider")
		go func(fullp []byte) {
			conn.Write(fullp)
			log.Println(time.Now().UnixNano()/int64(time.Millisecond), ":", cpreq.Identifier, " write request to provider done")
		}(fullp)
	}

}
