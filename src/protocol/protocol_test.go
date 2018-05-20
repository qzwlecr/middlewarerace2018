package protocol

import (
	"fmt"
	"testing"
)

func TestHttpByte2Arr(t *testing.T) {
	req := []byte("GET /index.html HTTP/1.1\r\nHost: www.example.com\r\n\r\n")

	var pack HttpPacks
	pack.fromByteArr(req)

	fmt.Printf("%+v\n", pack.payload)
}