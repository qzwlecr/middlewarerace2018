package protocol

import "bytes"
import "strings"
import "strconv"
import "encoding/json"

type SimpleConverter uint64

// HTTPToCustom : TODO test.
func (cnvt *SimpleConverter) HTTPToCustom(httpreq HttpPacks) (req CustRequest) {
	interf := httpreq.payload["interface"]
	method := httpreq.payload["method"]
	pmtpstr := httpreq.payload["parameterTypesString"]
	param := httpreq.payload["parameter"]
	att := httpreq.payload["attachments"]
	var buf bytes.Buffer
	buf.WriteString(interf)
	buf.WriteByte('\n')
	buf.WriteString(method)
	buf.WriteByte('\n')
	buf.WriteString(pmtpstr)
	buf.WriteByte('\n')
	buf.WriteString(param)
	buf.WriteByte('\n')
	buf.WriteString(att)
	req.content = buf.Bytes()
	return req
}

func marshalHelper(buf *bytes.Buffer, obj interface{}) {
	tmpres, _ := json.Marshal(obj)
	buf.Write(tmpres)
	buf.WriteByte('\n')
}

// CustomToDubbo : TODO test.
func (cnvt *SimpleConverter) CustomToDubbo(custreq CustRequest) (dubboreq DubboPacks) {
	// initialize dubbo basic structures
	dubboreq.magic = DUBBO_MAGIC
	dubboreq.reqType = 0
	dubboreq.reqType |= (DUBBO_REQUEST | DUBBO_NEEDREPLY)
	dubboreq.reqType |= 6 // serialization fastjson(6)
	dubboreq.status = 233 // no meaning
	dubboreq.reqId = uint64(*cnvt)
	*cnvt = *cnvt + 1
	strslice := strings.Split(string(custreq.content), "\n")
	var buf bytes.Buffer
	marshalHelper(&buf, DUBBO_VERSION)
	marshalHelper(&buf, strslice[0])
	marshalHelper(&buf, API_VERSION)
	marshalHelper(&buf, strslice[1])
	marshalHelper(&buf, strslice[2])
	marshalHelper(&buf, strslice[3])
	marshalHelper(&buf, strslice[4])
	dubboreq.payload = buf.Bytes()
	return dubboreq
}

func assert(a bool, pnstr string) {
	if !a {
		panic("Assertion Failed: " + pnstr)
	}
}

// DubboToCustom : TODO test.
func (cnvt *SimpleConverter) DubboToCustom(extrainfo uint64, dubboresp DubboPacks) (custresp CustResponse) {
	custresp.delay = extrainfo
	if extrainfo == CUST_MAGIC {
		custresp.reply = make([]byte, 0)
		return custresp
	}
	// so there are actual contents
	assert(dubboresp.reqType&uint8(6) != 0, "Serialization method not supported")
	strslice := strings.Split(string(dubboresp.payload), "\n")
	var rettype int
	err := json.Unmarshal([]byte(strslice[0]), &rettype)
	assert(err != nil, "unmarshalling return type: "+err.Error())
	assert(rettype == 1, "Unexpected response type: "+strconv.Itoa(rettype))
	var retval string
	err = json.Unmarshal([]byte(strslice[1]), &retval)
	assert(err != nil, "Unable to unmarshal return value: "+err.Error())
	custresp.reply = []byte(retval)
	return custresp
}

// CustomToHTTP : WARN response will be put in httpresp.payload['body'], which must be put into the HTTP body.
func (cnvt *SimpleConverter) CustomToHTTP(resp CustResponse) (httpresp HttpPacks) {
	assert(resp.delay != CUST_MAGIC, "Attempt to convert a rejected response to HTTP.")
	httpresp.payload = make(map[string]string)
	httpresp.payload["body"] = string(resp.reply)
	return httpresp
}