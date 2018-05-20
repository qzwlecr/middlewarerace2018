package protocol

import "net/http"
import "bytes"
import "encoding/binary"

func (cur *CustRequest) toByteArr() (buffer []byte) {
	return cur.content
}

func (cur *CustRequest) fromByteArr(buffer []byte) {
	cur.content = buffer
}

func (cus *CustResponse) toByteArr() (buffer []byte) {
	var pbuf bytes.Buffer
	u64buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(u64buf, cus.delay)
	pbuf.Write(u64buf)
	if cus.delay != CUST_MAGIC {
		pbuf.Write(cus.reply)
	}
	return pbuf.Bytes()
}

func (cus *CustResponse) fromByteArr(buffer []byte) {
	cus.delay = binary.LittleEndian.Uint64(buffer[0:4])
	if cus.delay != CUST_MAGIC {
		cus.reply = buffer[4:]
	} else {
		cus.reply = make([]byte, 0)
	}
}

func (dp *DubboPacks) toByteArr() (buffer []byte) {
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

func (dp *DubboPacks) fromByteArr(buffer []byte) {
	dp.magic = binary.BigEndian.Uint16(buffer[0:2])
	dp.reqType = uint8(buffer[2])
	dp.status = uint8(buffer[3])
	dp.reqId = binary.BigEndian.Uint64(buffer[4:12])
	dp.payload = buffer[16:]
}

func (httpack *HttpPacks) toByteArr() (buffer []byte) {
	var pbuf bytes.Buffer
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
	assert((len(httpack.method)&len(httpack.url)) != 0, "Not a valid HTTP pack.")
	req, err := http.NewRequest(httpack.method, httpack.url, &l2buf)
	assert(err == nil, "Error occurred while attempt to create request: ")
	req.Write(&pbuf)
	return pbuf.Bytes()
}

func (pack *HttpPacks) fromByteArr(buffer []byte) {
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
