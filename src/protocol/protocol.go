package protocol

import "bytes"
import (
	"encoding/binary"
	"log"
	"net/http"
)

// ToByteArr : make go happy
func (cur *CustRequest) ToByteArr() (buffer []byte) {
	return cur.Content
}

// FromByteArr: make go happy
func (cur *CustRequest) FromByteArr(buffer []byte) {
	cur.Content = buffer
}

// ToByteArr : make go happy
func (cus *CustResponse) ToByteArr() (buffer []byte) {
	var pbuf bytes.Buffer
	u64buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(u64buf, cus.Delay)
	pbuf.Write(u64buf)
	if cus.Delay != CUST_MAGIC {
		pbuf.Write(cus.Reply)
	}
	return pbuf.Bytes()
}

// FromByteArr: make go happy
func (cus *CustResponse) FromByteArr(buffer []byte) {
	cus.Delay = binary.LittleEndian.Uint64(buffer[0:8])
	if cus.Delay != CUST_MAGIC {
		cus.Reply = buffer[8:]
	} else {
		cus.Reply = make([]byte, 0)
	}
}

// ToByteArr : make go happy
func (dp *DubboPacks) ToByteArr() (buffer []byte) {
	var pbuf bytes.Buffer
	u16buf := make([]byte, 2)
	u32buf := make([]byte, 4)
	u64buf := make([]byte, 8)
	// first, the Magic
	binary.LittleEndian.PutUint16(u16buf, dp.Magic)
	pbuf.Write(u16buf)
	pbuf.WriteByte(byte(dp.ReqType))
	pbuf.WriteByte(byte(dp.Status))
	binary.LittleEndian.PutUint64(u64buf, dp.ReqId)
	pbuf.Write(u64buf)
	binary.LittleEndian.PutUint32(u32buf, uint32(len(dp.Payload)))
	pbuf.Write(u32buf)
	pbuf.Write(dp.Payload)
	return pbuf.Bytes()
}

// FromByteArr: make go happy
func (dp *DubboPacks) FromByteArr(buffer []byte) {
	dp.Magic = binary.LittleEndian.Uint16(buffer[0:2])
	dp.ReqType = uint8(buffer[2])
	dp.Status = uint8(buffer[3])
	dp.ReqId = binary.LittleEndian.Uint64(buffer[4:12])
	dp.Payload = buffer[16:]
}

// ToByteArr : make go happy
func (httpack *HttpPacks) ToByteArr() (buffer []byte) {
	assert((len(httpack.Direct)^len(httpack.Payload)) != 0, "HTTP packs Direct & Payload both exist or both non-exist.")
	var l2buf bytes.Buffer
	if len(httpack.Direct) != 0 {
		l2buf.WriteString(httpack.Direct)
	} else {
		cnt := 0
		for k, v := range httpack.Payload {
			for kk, vv := range v {
				l2buf.WriteString(k)
				l2buf.WriteString("=")
				l2buf.WriteString(vv)
				if cnt != len(httpack.Payload)-1 || kk != len(vv)-1 {
					l2buf.WriteString("&")
				}
			}
			cnt = cnt + 1
		}
	}
	return l2buf.Bytes()
}

// FromByteArr: make go happy
func (pack *HttpPacks) FromByteArr(buffer []byte) {
	log.Panic("Maybe using unusable code.")
	//packBuf := bytes.NewBuffer(buffer)
	//
	//line, err := packBuf.ReadBytes('\r')
	//if err != nil {
	//	return
	//}
	//// throw away CR in line
	//line = line[:len(line)-1]
	//
	//// throw away LF
	//packBuf.Next(1)
	//
	//// init map
	//pack.Payload = make(map[string]string)
	//if bytes.Compare(line[0:4], []byte("HTTP")) != 0 {
	//	// http request
	//
	//	// processing the start line
	//	sLnElems := bytes.Split(line, []byte(" "))
	//	if len(sLnElems) != 3 {
	//		return
	//	}
	//
	//	method := sLnElems[0]
	//	pack.Payload["HTTP_method"] = string(method)
	//
	//	target := sLnElems[1]
	//	pack.Payload["request_target"] = string(target)
	//
	//	version := sLnElems[2]
	//	pack.Payload["HTTP_version"] = string(version)
	//} else {
	//	// http response
	//	// processing the Status line
	//	lineBuf := bytes.NewBuffer(line)
	//	version, err := lineBuf.ReadBytes(' ')
	//	if err != nil {
	//		return
	//	}
	//
	//	version = version[:len(version)-1]
	//	pack.Payload["HTTP_version"] = string(version)
	//
	//	status, err := lineBuf.ReadBytes(' ')
	//	if err != nil {
	//		return
	//	}
	//	status = status[:len(version)-1]
	//	pack.Payload["status_code"] = string(status)
	//
	//	text := lineBuf.Bytes()
	//	pack.Payload["status_text"] = string(text)
	//}
	//
	//for line, err = packBuf.ReadBytes('\r'); err != nil && len(line) != 1; {
	//	// throw away LF
	//	packBuf.Next(1)
	//
	//	lineBuf := bytes.NewBuffer(line)
	//	header, err := lineBuf.ReadBytes(':')
	//	if err != nil {
	//		return
	//	}
	//
	//	headerStr := string(header[:len(header)-1])
	//	fieldVal := bytes.TrimLeft(lineBuf.Bytes(), " ")
	//	// throw away CR
	//	fieldVal = fieldVal[:len(fieldVal)-1]
	//	pack.Payload[headerStr] = string(fieldVal)
	//}
	//if err != nil {
	//	return
	//}
	//
	//// throw away LF of the blank line
	//packBuf.Next(1)
	//
	//body := packBuf.Bytes()
	//// excluding CRLF at the end
	//if len(body) > 2 {
	//	pack.Payload["body"] = string(body)
	//}
}

func (pack *HttpPacks) FromRequests(r *http.Request) {
	r.ParseForm()
	pack.Payload = r.Form
}
