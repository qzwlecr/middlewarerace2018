package protocol

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"testing"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")

func randAscString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func unmarshalHelper(spcont string, t *testing.T) string {
	var buf bytes.Buffer
	buf.WriteString(spcont)
	var tmp string
	err := json.Unmarshal(buf.Bytes(), &tmp)
	if err != nil {
		t.Fatal("Unmarshalling ", spcont, "failed:", err)
	}
	return tmp
}

/*func TestCustDubbo(t *testing.T) {
	var cnvt SimpleConverter
	cnvt.id = 0
	var custreq CustRequest
	//fmt.Println("Rand string is:")
	rndstr := randAscString(10)
	//fmt.Println(rndstr)
	// create custrequest content
	var creqbuf bytes.Buffer
	for i := 0; i < 5; i++ {
		creqbuf.WriteString(rndstr)
		creqbuf.WriteString("\n")
	}
	custreq.Content = creqbuf.Bytes()
	dubboreq := (&cnvt).CustomToDubbo(custreq)
	if dubboreq.ReqId != 0 {
		t.Fatal("ReqId mismatch, correct error = ", 1, dubboreq.ReqId)
	}
	//fmt.Println("Dubbo Request Cont is:")
	//fmt.Println(string(dubboreq.Payload))
	strslice := strings.Split(string(dubboreq.Payload), "\n")
	// there will be total 7 valid parts.
	if len(strslice) != 8 {
		t.Fatal("unslicing error, total parts: correct error = ", 8, len(strslice))
	}
	// check DUBBO_VERSION
	if unmarshalHelper(strslice[0], t) != DUBBO_VERSION {
		t.Fatal("Content mismatch at DUBBO_VERSION: correct error = ", DUBBO_VERSION, unmarshalHelper(strslice[0], t))
	}
	tmpustr := unmarshalHelper(strslice[1], t)
	if tmpustr != rndstr {
		t.Fatal("Content mismatch at 0: correct error = ", rndstr, tmpustr)
	}
	tmpuapi := unmarshalHelper(strslice[2], t)
	if tmpuapi != API_VERSION {
		t.Fatal("Content mismatch at API_VERSION: correct error = ", API_VERSION, tmpuapi)
	}
	for i := 3; i < 7; i++ {
		tmpustr = unmarshalHelper(strslice[i], t)
		if tmpustr != rndstr {
			t.Fatal("Content mismatch at ", i-2, ": correct error = ", rndstr, tmpustr)
		}
	}
	if len(strslice[7]) != 0 {
		t.Fatal("Null mismatch, the content is: ", strslice[7])
	}
}*/
