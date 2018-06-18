package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
)

//const LOGGING = false

//const FORCE_ASSERTION = false

// SimpleConverter : the converter that do something great!
type SimpleConverter struct {
	id uint64
	mu sync.Mutex
}

// HTTPToCustom : TODO test.
func (cnvt *SimpleConverter) HTTPToCustom(httpreq HttpPacks) (req CustRequest, err error) {
	//defer timing.Since(time.Now(), "CNVT HTTPToCust")
	cnvt.mu.Lock()
	req.Identifier = cnvt.id
	cnvt.id = cnvt.id + 1
	cnvt.mu.Unlock()
	interf := httpreq.Payload["interface"]
	method := httpreq.Payload["method"]
	pmtpstr := httpreq.Payload["parameterTypesString"]
	param := httpreq.Payload["parameter"]
	att := httpreq.Payload["attachments"]
	var buf bytes.Buffer
	//TODO: Hardcode to get the first element of the array
	buf.WriteString(interf[0])
	buf.WriteByte('\n')
	buf.WriteString(method[0])
	buf.WriteByte('\n')
	buf.WriteString(pmtpstr[0])
	buf.WriteByte('\n')
	buf.WriteString(param[0])
	buf.WriteByte('\n')
	// generate attachments like a boss!
	rattach := make(map[string]string)
	if att != nil {
		err := json.Unmarshal([]byte(att[0]), &rattach)
		//if FORCE_ASSERTION {
		//	assert(err == nil, "The HTTP attach not in JSON format.")
		//} else if err != nil {
		//	return req, fmt.Errorf("the HTTP attach not in JSON format, %s", err.Error())
		//}
		if err != nil {
			return req, fmt.Errorf("the HTTP attach not in JSON format, %s", err.Error())
		}
	}

	// 3 attach elements should be added: dubbov, path and version
	// path is the interface..
	rattach["dubbo"] = DUBBO_VERSION
	rattach["version"] = API_VERSION
	rattach["path"] = interf[0]
	// serialize into it
	rser, err := json.Marshal(rattach)
	//if FORCE_ASSERTION {
	//	assert(err == nil, "Attachments serialization failed.")
	//} else if err != nil {
	//	return req, fmt.Errorf("attachments serialization failed: %s", err.Error())
	//}
	if err != nil {
		return req, fmt.Errorf("attachments serialization failed: %s", err.Error())
	}
	buf.Write(rser)
	req.Content = buf.Bytes()
	return req, nil
}

func marshalHelper(buf *bytes.Buffer, obj interface{}) {
	tmpres, _ := json.Marshal(obj)
	buf.Write(tmpres)
	buf.WriteByte('\n')
}

// CustomToDubbo : TODO test.
func (cnvt *SimpleConverter) CustomToDubbo(custreq CustRequest) (dubboreq DubboPacks, err error) {
	//defer timing.Since(time.Now(), "CNVT CustToDubb")
	// initialize dubbo basic structures
	dubboreq.Magic = DUBBO_MAGIC
	dubboreq.ReqType = 0
	dubboreq.ReqType |= (DUBBO_REQUEST | DUBBO_NEEDREPLY)
	dubboreq.ReqType |= 6 // serialization fastjson(6)
	dubboreq.Status = 0   // no meaning
	dubboreq.ReqId = custreq.Identifier
	strslice := strings.Split(string(custreq.Content), "\n")
	//if LOGGING {
	//	log.Println("SlicedStr:")
	//	log.Println(strslice)
	//}
	//if FORCE_ASSERTION {
	//	assert(len(strslice) > 4, "Attachment refill failed.")
	if len(strslice) <= 4 {
		return dubboreq, fmt.Errorf("attachment refill failed, length not enough")
	}
	var buf bytes.Buffer
	marshalHelper(&buf, DUBBO_VERSION)
	marshalHelper(&buf, strslice[0])
	marshalHelper(&buf, nil)
	marshalHelper(&buf, strslice[1])
	marshalHelper(&buf, strslice[2])
	marshalHelper(&buf, strslice[3])
	// this is the attachments. optional, so..
	/*if len(strslice) >= 5 {
		marshalHelper(&buf, strslice[4])
	}*/
	// the real attachment!
	//if LOGGING {
	//	log.Println("Attachments:")
	//}
	for i := 4; i < len(strslice); i++ {
		buf.WriteString(strslice[i])
		buf.WriteByte('\n')
		//if LOGGING {
		//	log.Println(strslice[i])
		//}
	}
	dubboreq.Payload = buf.Bytes()
	return dubboreq, nil
}

func assert(a bool, pnstr string) {
	if !a {
		log.Panicln("Assertion Failed: " + pnstr)
		panic("Assertion Failed: " + pnstr)
	}
}

// DubboToCustom : TODO test.
func (cnvt *SimpleConverter) DubboToCustom(extrainfo uint64, dubboresp DubboPacks) (custresp CustResponse, err error) {
	//defer timing.Since(time.Now(), "CNVT DubbToCust")
	custresp.Delay = extrainfo
	if extrainfo == CUST_MAGIC {
		custresp.Reply = make([]byte, 0)
		return custresp, nil
	}
	// so there are actual contents
	//if FORCE_ASSERTION {
	//	assert(dubboresp.ReqType&uint8(6) != 0, "Serialization method not supported")
	if dubboresp.ReqType&uint8(6) == 0 {
		return custresp, fmt.Errorf("unsupported serialization method: %d", dubboresp.ReqType)
	}
	strslice := strings.Split(string(dubboresp.Payload), "\n")
	var rettype int
	err = json.Unmarshal([]byte(strslice[0]), &rettype)
	//if FORCE_ASSERTION {
	//	assert(err == nil, "unmarshalling return type: ")
	if err != nil {
		return custresp, fmt.Errorf("unmarshalling failed from dubbo: %s", err.Error())
	}
	//if FORCE_ASSERTION {
	//	assert(rettype == 1, "Unexpected response type: "+strconv.Itoa(rettype))
	if rettype != 1 {
		return custresp, fmt.Errorf("unexpected response type: %d", rettype)
	}
	var retval int64
	err = json.Unmarshal([]byte(strslice[1]), &retval)
	//if FORCE_ASSERTION {
	//	assert(err == nil, "Unable to unmarshal return value: ")
	if err != nil {
		return custresp, fmt.Errorf("unable to unmarshal dubbo return value: %s", err.Error())
	}
	custresp.Reply = []byte(strslice[1])
	custresp.Identifier = dubboresp.ReqId
	return custresp, nil
}

// CustomToHTTP : woo-hoo!
func (cnvt *SimpleConverter) CustomToHTTP(resp CustResponse) (httpresp HttpPacks, err error) {
	//defer timing.Since(time.Now(), "CNVT CustToHTTP")
	//if FORCE_ASSERTION {
	//	assert(resp.Delay != CUST_MAGIC, "Attempt to convert a rejected response to HTTP.")
	if resp.Delay == CUST_MAGIC {
		return httpresp, fmt.Errorf("attempt to convert a rejected custresp to http")
	}
	httpresp.Payload = make(map[string][]string)
	httpresp.Payload["body"] = []string{string(resp.Reply)}
	return httpresp, nil
}
