package protocol

// Protocol : make go happy
type Protocol interface {
	toByteArr() (buffer []byte)
	fromByteArr(buffer []byte)
}

// Converter : make go happy
type Converter interface {
	HTTPToCustom(httpreq HttpPacks) (req CustRequest)
	CustomToDubbo(custreq CustRequest) (dubboreq DubboPacks)
	DubboToCustom(extrainfo uint64, dubboresp DubboPacks) (custresp CustResponse)
	CustomToHTTP(resp CustResponse) (httpresp HttpPacks)
}

// CustRequest : make go happy
type CustRequest struct {
	content []byte
}

// CustResponse : make go happy
type CustResponse struct {
	delay uint64
	reply []byte
}

// CUST_MAGIC : the ultimate magic (TuM)
const CUST_MAGIC uint64 = 7234316346692625778

// HttpPacks : make Go HAPPY!
type HttpPacks struct {
	url     string
	method  string
	direct  string
	payload map[string]string
}

// DubboPacks : make Go Happy too!
type DubboPacks struct {
	magic   uint16
	reqType uint8
	status  uint8
	reqId   uint64
	payload []byte // => length(uint32) + payload_content(byte[])
}

// DUBBO_VERSION : the very version. Very magic.
const DUBBO_VERSION = "2.6.0"

// API_VERSION : the next very version. Magic again.
const API_VERSION = "0.0.0"

// DUBBO_MAGIC : hmmmmm, really interpreted like that?
const DUBBO_MAGIC = 0xdabb

// DUBBO_REQUEST : for reqType, use bitwise-OR
const DUBBO_REQUEST = 128

// DUBBO_NEEDREPLY : also
const DUBBO_NEEDREPLY = 64

// DUBBO_EVENT : woo-hoo!
const DUBBO_EVENT = 32

// DUBBO_OK : script generated, make go happy
const DUBBO_OK = 20

// DUBBO_CLIENT_TIMEOUT : script generated, make go happy
const DUBBO_CLIENT_TIMEOUT = 30

// DUBBO_SERVER_TIMEOUT : script generated, make go happy
const DUBBO_SERVER_TIMEOUT = 31

// DUBBO_BAD_REQUEST : script generated, make go happy
const DUBBO_BAD_REQUEST = 40

// DUBBO_BAD_RESPONSE : script generated, make go happy
const DUBBO_BAD_RESPONSE = 50

// DUBBO_SERVICE_NOT_FOUND : script generated, make go happy
const DUBBO_SERVICE_NOT_FOUND = 60

// DUBBO_SERVICE_ERROR : script generated, make go happy
const DUBBO_SERVICE_ERROR = 70

// DUBBO_SERVER_ERROR : script generated, make go happy
const DUBBO_SERVER_ERROR = 80

// DUBBO_CLIENT_ERROR : script generated, make go happy
const DUBBO_CLIENT_ERROR = 90

// DUBBO_SERVER_THREADPOOL_EXHAUSTED_ERROR : script generated, make go happy
const DUBBO_SERVER_THREADPOOL_EXHAUSTED_ERROR = 100
