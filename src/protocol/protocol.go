package protocol

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
