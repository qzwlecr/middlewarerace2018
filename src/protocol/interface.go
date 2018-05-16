package protocol

type Converter interface {
	HTTPToCustom(buf []byte) (req Request)
	CustomToDubbo(buf []byte) (buf2 []byte)
	DubboToCustom(buf []byte) (buf2 []byte)
	CustomToHTTP(buf []byte) (rep Reply)
}

type Request struct {
	outBuf   []byte
	someInfo string
}

type Reply struct {
	outBuf   []byte
	someInfo string
}
