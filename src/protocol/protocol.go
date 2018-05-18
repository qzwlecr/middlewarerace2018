package protocol

import "bytes"

// func (pack *HttpPacks) toByteArr() (buffer []byte) {

// }

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
