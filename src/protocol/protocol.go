package protocol

import "bytes"
import (
	"encoding/binary"
	"fmt"
	"net/http"
	MemoryPool "utility/mempool"
)

var mmp = MemoryPool.NewZhwkMemoryPool(20)

// ToByteArr : make go happy
func (cur *CustRequest) ToByteArr() (buffer []byte, err error) {
	//defer timing.Since(time.Now(), "PROT CustReqToByteArr")
	var buf bytes.Buffer
	bid, idbuf, err := mmp.Acquire()
	defer mmp.Release(bid)
	binary.BigEndian.PutUint64(idbuf, cur.Identifier)
	buf.Write(idbuf[:8])
	buf.Write(cur.Content)
	return buf.Bytes(), nil
}

// FromByteArr: make go happy
func (cur *CustRequest) FromByteArr(buffer []byte) (err error) {
	//defer timing.Since(time.Now(), "PROT CustReqFromByteArr")
	cur.Identifier = binary.BigEndian.Uint64(buffer[:8])
	cur.Content = make([]byte, len(buffer[8:]))
	copy(cur.Content, buffer[8:])
	return nil
}

// ToByteArr : make go happy
func (cus *CustResponse) ToByteArr() (buffer []byte, err error) {
	//defer timing.Since(time.Now(), "PROT CustRespToByteArr")
	var pbuf bytes.Buffer
	bid, u64buf, err := mmp.Acquire()
	defer mmp.Release(bid)
	binary.BigEndian.PutUint64(u64buf, cus.Identifier)
	pbuf.Write(u64buf[:8])
	binary.BigEndian.PutUint64(u64buf, cus.Delay)
	pbuf.Write(u64buf[:8])
	//if cus.Delay != CUST_MAGIC {
		pbuf.Write(cus.Reply)
	//}
	return pbuf.Bytes(), nil
}

// FromByteArr: make go happy
func (cus *CustResponse) FromByteArr(buffer []byte) (err error) {
	//defer timing.Since(time.Now(), "PROT CustRespFromByteArr")
	cus.Identifier = binary.BigEndian.Uint64(buffer[:8])
	cus.Delay = binary.BigEndian.Uint64(buffer[8:16])
	//if cus.Delay != CUST_MAGIC {
	cus.Reply = make([]byte, len(buffer[16:]))
	copy(cus.Reply, buffer[16:])
	//} else {
	//	cus.Reply = make([]byte, 0)
	//}
	return nil
}

// ToByteArr : make go happy
func (dp *DubboPacks) ToByteArr() (buffer []byte, err error) {
	//defer timing.Since(time.Now(), "PROT DubbToByteArr")
	var pbuf bytes.Buffer
	// u16buf := make([]byte, 2)
	// u32buf := make([]byte, 4)
	// u64buf := make([]byte, 8)
	bid, buf, _ := mmp.Acquire()
	defer mmp.Release(bid)
	// first, the Magic
	binary.BigEndian.PutUint16(buf, dp.Magic)
	pbuf.Write(buf[:2])
	pbuf.WriteByte(byte(dp.ReqType))
	pbuf.WriteByte(byte(dp.Status))
	binary.BigEndian.PutUint64(buf, dp.ReqId)
	pbuf.Write(buf[:8])
	binary.BigEndian.PutUint32(buf, uint32(len(dp.Payload)))
	pbuf.Write(buf[:4])
	pbuf.Write(dp.Payload)
	//if LOGGING {
	//	log.Println("DUBB as bytes:")
	//	log.Println(pbuf.Bytes())
	//	log.Println("DUBB as string:")
	//	log.Println(string(pbuf.Bytes()))
	//}
	return pbuf.Bytes(), nil
}

// FromByteArr: make go happy
func (dp *DubboPacks) FromByteArr(buffer []byte) (err error) {
	//defer timing.Since(time.Now(), "PROT DubbFromByteArr")
	//if FORCE_ASSERTION {
	//	assert(len(buffer) > 16, "Too short in dubbo.")
	if len(buffer) <= 16 {
		return fmt.Errorf("Too short in dubbo frombytearr()")
	}
	dp.Magic = binary.BigEndian.Uint16(buffer[0:2])
	//if FORCE_ASSERTION {
	//	assert(dp.Magic == DUBBO_MAGIC, "Not so magic in dubbo.")
	if dp.Magic != DUBBO_MAGIC {
		return fmt.Errorf("Magic mismatch in dubbo frombytearr()")
	}
	dp.ReqType = uint8(buffer[2])
	dp.Status = uint8(buffer[3])
	dp.ReqId = binary.BigEndian.Uint64(buffer[4:12])
	dp.Payload = make([]byte, len(buffer[16:]))
	copy(dp.Payload, buffer[16:])
	return nil
}

// CheckFormat check if the pack format is correct.
func (dp *DubboPacks) CheckFormat(buffer []byte) (err error) {
	//defer timing.Since(time.Now(), "DEBG DubbCheckFmt")
	//assert(len(buffer) > 16, "Too short in dubbo.")
	dp.Magic = binary.BigEndian.Uint16(buffer[0:2])
	//assert(dp.Magic == DUBBO_MAGIC, "Not so magic in dubbo.")
	dp.ReqType = uint8(buffer[2])
	//assert(dp.ReqType&DUBBO_REQUEST != 0, "Not a request.")
	//assert(dp.ReqType&DUBBO_NEEDREPLY != 0, "Not 2-way.")
	//assert(dp.ReqType&DUBBO_EVENT == 0, "Is a event.")
	//assert(dp.ReqType&0x6 != 0, "Serialization type gg.")
	dp.Status = uint8(buffer[3])
	dp.ReqId = binary.BigEndian.Uint64(buffer[4:12])
	//payloadlen := binary.BigEndian.Uint32(buffer[12:16])
	dp.Payload = buffer[16:]
	//assert(len(dp.Payload) == int(payloadlen), "Dynamic length part mismatched.")
	return nil
}

// ToByteArr : make go happy
func (httpack *HttpPacks) ToByteArr() (buffer []byte, err error) {
	//defer timing.Since(time.Now(), "PROT HttpToByteArr <- WARNING ->")
	//if FORCE_ASSERTION {
	//	assert((len(httpack.Direct) ^ len(httpack.Payload)) != 0, "HTTP packs Direct & Payload both exist or both non-exist.")
	if (len(httpack.Direct) ^ len(httpack.Payload)) == 0 {
		return buffer, fmt.Errorf("Http direct and payload both exist or non-exist")
	}
	var l2buf bytes.Buffer
	if len(httpack.Direct) != 0 {
		l2buf.WriteString(httpack.Direct)
	} else {
		//cnt := 0
		//for k, v := range httpack.Payload {
		//	for kk, vv := range v {
		//		l2buf.WriteString(k)
		//		l2buf.WriteString("=")
		//		l2buf.WriteString(vv)
		//		if cnt != len(httpack.Payload)-1 || kk != len(vv)-1 {
		//			l2buf.WriteString("&")
		//		}
		//	}
		//	cnt = cnt + 1
		//}
		l2buf.Write([]byte(httpack.Payload["body"][0]))
	}
	return l2buf.Bytes(), nil
}

// FromByteArr make go happy
func (pack *HttpPacks) FromByteArr(buffer []byte) (err error) {
	//if FORCE_ASSERTION {
	//	assert(false, "Static asserted.")
	//} else {
	//	assert(false, "Static asserted")
	//}
	// log.Panic("Maybe using unusable code.")
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
	return nil
}

func (pack *HttpPacks) FromRequests(r *http.Request) (err error) {
	r.ParseForm()
	pack.Payload = r.Form
	return nil
}

func (pack *HttpPacks) FromFasthttpRequests(key, value []byte) {
	pack.Payload[string(key)] = []string{string(value)}
}
