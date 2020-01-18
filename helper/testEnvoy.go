package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
)

type ByteEqual struct {
	Byte byte
	Equal bool
}

func (b *ByteEqual) String() string {
	if b.Equal {
		return Green(b.Byte)
	}

	return Red(b.Byte)
}

var (
	Black   = Color("\033[1;30m%s\033[0m")
	Red     = Color("\033[1;31m%s\033[0m")
	Green   = Color("\033[1;32m%s\033[0m")
	Yellow  = Color("\033[1;33m%s\033[0m")
	Purple  = Color("\033[1;34m%s\033[0m")
	Magenta = Color("\033[1;35m%s\033[0m")
	Teal    = Color("\033[1;36m%s\033[0m")
	White   = Color("\033[1;37m%s\033[0m")

	oneTag = "localhost:8088"
	twoTag = "localhost:8080"
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func post(host, uri string, body []byte) ([]byte, error) {

	req := fasthttp.AcquireRequest()
	req.SetBody(body)
	req.Header.SetHost(host)
	req.Header.SetMethodBytes([]byte("POST")) // or "DELETE"
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Grpc-Web", "1")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Content-Type", "application/grpc-web+proto")
	req.Header.Set("Grpc-Insecure", "true")
	req.URI().Update(uri)
	resp := fasthttp.AcquireResponse()
	error := fasthttp.Do(req, resp)

	if error != nil {
		return []byte{}, error
	}

	return resp.Body(), nil
}

func diff(one, two []byte, oneTag, twoTag string) {

	resp1 := one
	resp2 := two

	max := len(resp1)
	if len(resp2) > max {
		max = len(resp2)
	}

	be1 := make([]*ByteEqual,len(resp1))
	be2 := make([]*ByteEqual,len(resp2))

	for i := 0; i < max; i++ {
		if i < len(resp1) && i < len(resp2) {
			isEqualFalse := false
			if resp1[i] == resp2[i] {
				isEqualFalse = true
			}
			be1[i] = &ByteEqual{Byte:resp1[i], Equal: isEqualFalse}
			be2[i] = &ByteEqual{Byte:resp2[i], Equal: isEqualFalse}
			continue
		}

		if i < len(resp1) {
			be1[i] = &ByteEqual{Byte:resp1[i], Equal: false}
		} else {
			be2[i] = &ByteEqual{Byte:resp2[i], Equal: false}
		}
	}

	fmt.Println(oneTag, be1[:30])
	fmt.Println(twoTag, be2[:30])
}

func GetExchangeInfo() {

	resp1, err1 := post(oneTag,"/proto.DatasourceService/GetExchangeInfo", []byte{0,0,0,0,0})
	resp2, err2 := post(twoTag,"/proto.DatasourceService/GetExchangeInfo", []byte{0,0,0,0,0})

	if err1 != nil {
		fmt.Println("Error", err1)
		return
	}

	if err2 != nil {
		fmt.Println("Error", err2)
		return
	}

	diff(resp1, resp2, oneTag, twoTag)
}

func GetEvent() {

	resp1, err1 := post(oneTag,"/proto.Dataloader/getEvent", []byte{0,0,0,0,9,10,7,18,5,65,69,66,78,66})
	resp2, err2 := post(twoTag,"/proto.Dataloader/getEvent", []byte{0,0,0,0,9,10,7,18,5,65,69,66,78,66})

	if err1 != nil {
		fmt.Println("Error", err1)
		return
	}

	if err2 != nil {
		fmt.Println("Error", err2)
		return
	}

	diff(resp1, resp2, oneTag, twoTag)
}

func GetMeta() {

	resp1, err1 := post(oneTag,"/proto.Dataloader/getMeta", []byte{0,0,0,0,9,10,7,18,5,65,69,66,78,66})
	resp2, err2 := post(twoTag,"/proto.Dataloader/getMeta", []byte{0,0,0,0,9,10,7,18,5,65,69,66,78,66})

	if err1 != nil {
		fmt.Println("Error", err1)
		return
	}

	if err2 != nil {
		fmt.Println("Error", err2)
		return
	}

	diff(resp1, resp2, oneTag, twoTag)
}

func GetTicker() {

	resp1, err1 := post(oneTag,"/proto.Dataloader/getTicker", []byte{0,0,0,0,9,10,7,18,5,65,69,66,78,66})
	resp2, err2 := post(twoTag,"/proto.Dataloader/getTicker", []byte{0,0,0,0,9,10,7,18,5,65,69,66,78,66})

	if err1 != nil {
		fmt.Println("Error", err1)
		return
	}

	if err2 != nil {
		fmt.Println("Error", err2)
		return
	}

	diff(resp1, resp2, oneTag, twoTag)
}

func main() {

	//GetExchangeInfo()
	//GetEvent()
	GetMeta()
	GetTicker()

}