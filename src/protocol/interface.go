package protocol

// Protocol : make go happy
type Protocol interface {
	toByteArr() (buffer []byte)
	fromByteArr(buffer []byte)
}

// Converter : make go happy
type Converter interface {
	HTTPToCustom(httpreq Protocol) (req Protocol)
	CustomToDubbo(custreq Protocol) (dubboreq Protocol)
	DubboToCustom(dubboresp Protocol) (custresp Protocol)
	CustomToHTTP(resp Protocol) (httpresp Protocol)
}

// CustRequest : make go happy
type CustRequest struct {
	content []byte
}

// CustResponse : make go happy
type CustResponse struct {
	delay uint32
	reply []byte
}
