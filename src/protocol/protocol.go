package protocol

import "bytes"
import "encoding/binary"

// ToByteArr : make go happy
func (cur *CustRequest) ToByteArr() (buffer []byte) {
	return cur.content
}

// FromByteArr: make go happy
func (cur *CustRequest) FromByteArr(buffer []byte) {
	cur.content = buffer
}

// ToByteArr : make go happy
func (cus *CustResponse) ToByteArr() (buffer []byte) {
	var pbuf bytes.Buffer
	u64buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(u64buf, cus.Delay)
	pbuf.Write(u64buf)
	if cus.Delay != CUST_MAGIC {
		pbuf.Write(cus.reply)
	}
	return pbuf.Bytes()
}

// FromByteArr: make go happy
func (cus *CustResponse) FromByteArr(buffer []byte) {
	cus.Delay = binary.LittleEndian.Uint64(buffer[0:4])
	if cus.Delay != CUST_MAGIC {
		cus.reply = buffer[4:]
	} else {
		cus.reply = make([]byte, 0)
	}
}

// ToByteArr : make go happy
func (dp *DubboPacks) ToByteArr() (buffer []byte) {
	var pbuf bytes.Buffer
	u16buf := make([]byte, 2)
	u32buf := make([]byte, 4)
	u64buf := make([]byte, 8)
	// first, the magic
	binary.BigEndian.PutUint16(u16buf, dp.magic)
	pbuf.Write(u16buf)
	pbuf.WriteByte(byte(dp.reqType))
	pbuf.WriteByte(byte(dp.status))
	binary.BigEndian.PutUint64(u64buf, dp.reqId)
	pbuf.Write(u64buf)
	binary.BigEndian.PutUint32(u32buf, uint32(len(dp.payload)))
	pbuf.Write(u32buf)
	pbuf.Write(dp.payload)
	return pbuf.Bytes()
}

// FromByteArr: make go happy
func (dp *DubboPacks) FromByteArr(buffer []byte) {
	dp.magic = binary.BigEndian.Uint16(buffer[0:2])
	dp.reqType = uint8(buffer[2])
	dp.status = uint8(buffer[3])
	dp.reqId = binary.BigEndian.Uint64(buffer[4:12])
	dp.payload = buffer[16:]
}

// ToByteArr : make go happy
func (httpack *HttpPacks) ToByteArr() (buffer []byte) {
	assert((len(httpack.direct)^len(httpack.payload)) != 0, "HTTP packs direct & payload both exist or both non-exist.")
	var l2buf bytes.Buffer
	if len(httpack.direct) != 0 {
		l2buf.WriteString(httpack.direct)
	} else {
		cnt := 0
		for k, v := range httpack.payload {
			l2buf.WriteString(k)
			l2buf.WriteString("=")
			l2buf.WriteString(v)
			if cnt != len(httpack.payload)-1 {
				l2buf.WriteString("&")
			}
		}
	}
	return l2buf.Bytes()
}

// FromByteArr: make go happy
func (pack *HttpPacks) FromByteArr(buffer []byte) {
	packBuf := bytes.NewBuffer(buffer)

	line, err := packBuf.ReadBytes('\r')
	if err != nil {
		return
	}
	// throw away CR in line
	line = line[:len(line)-1]

	// throw away LF
	packBuf.Next(1)

	// init map
	pack.payload = make(map[string]string)
	if bytes.Compare(line[0:4], []byte("HTTP")) != 0 {
		// http request

		// processing the start line
		sLnElems := bytes.Split(line, []byte(" "))
		if len(sLnElems) != 3 {
			return
		}

		method := sLnElems[0]
		pack.payload["HTTP_method"] = string(method)

		target := sLnElems[1]
		pack.payload["request_target"] = string(target)

		version := sLnElems[2]
		pack.payload["HTTP_version"] = string(version)
	} else {
		// http response
		// processing the status line
		lineBuf := bytes.NewBuffer(line)
		version, err := lineBuf.ReadBytes(' ')
		if err != nil {
			return
		}

		version = version[:len(version)-1]
		pack.payload["HTTP_version"] = string(version)

		status, err := lineBuf.ReadBytes(' ')
		if err != nil {
			return
		}
		status = status[:len(version)-1]
		pack.payload["status_code"] = string(status)

		text := lineBuf.Bytes()
		pack.payload["status_text"] = string(text)
	}

	for line, err = packBuf.ReadBytes('\r'); err != nil && len(line) != 1; {
		// throw away LF
		packBuf.Next(1)

		lineBuf := bytes.NewBuffer(line)
		header, err := lineBuf.ReadBytes(':')
		if err != nil {
			return
		}

		headerStr := string(header[:len(header)-1])
		fieldVal := bytes.TrimLeft(lineBuf.Bytes(), " ")
		// throw away CR
		fieldVal = fieldVal[:len(fieldVal)-1]
		pack.payload[headerStr] = string(fieldVal)
	}
	if err != nil {
		return
	}

	// throw away LF of the blank line
	packBuf.Next(1)

	body := packBuf.Bytes()
	// excluding CRLF at the end
	if len(body) > 2 {
		pack.payload["body"] = string(body)
	}
}
