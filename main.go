package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
)

func Index(ctx *fasthttp.RequestCtx) {
	fmt.Println(ctx, "Welcome!\n")
}

type Message struct {}
type Reply struct {
	Data []byte
}

func (m Reply) Reset() {}

func (m Reply) ProtoMessage() {}

func (m Reply) String() string {
	return fmt.Sprintln(m.Data)
}

func (m *Reply) XXX_Unmarshal(b []byte) error {
	m.Data = b
	return nil
}

func (m Message) Reset() {}

func (m Message) ProtoMessage() {}

func (m Message) String() string {
	return proto.CompactTextString(m)
}

func OptionsProto(ctx *fasthttp.RequestCtx) {

	method := fmt.Sprintf("/proto.%s/%s", ctx.UserValue("service"), ctx.UserValue("method"))

	log.Println(string(ctx.Method()), method)
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		fmt.Println(err)
		return
	}

	in := Message{}
	out := Reply{}

	err = conn.Invoke(context.Background(), method, &in, &out)
	
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(out)

	ctx.Response.Header.Set("Access-Control-Allow-Origin","*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods","GET, POST, OPTIONS")
	ctx.Response.Header.Set("Access-Control-Allow-Headers","DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Content-Transfer-Encoding,Custom-Header-1,X-Accept-Content-Transfer-Encoding,X-Accept-Response-Streaming,X-User-Agent,X-Grpc-Web")
	ctx.Response.Header.Set("Access-Control-Max-Age","1728000")
	ctx.Response.Header.SetContentType("text/plain charset=UTF-8")
	ctx.Response.Header.SetContentLength(0)
	ctx.Response.SetStatusCode(204)
}

func PostProto(ctx *fasthttp.RequestCtx) {

	method := fmt.Sprintf("/proto.%s/%s", ctx.UserValue("service"), ctx.UserValue("method"))

	log.Println(string(ctx.Method()), method)
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		fmt.Println(err)
		return
	}

	in := Message{}
	out := Reply{}

	err = conn.Invoke(ctx, method, &in, &out)

	desc := &grpc.StreamDesc{}

	stream, err := conn.NewStream(context.Background(), desc, method)

	if err != nil {
		fmt.Println(err)
	}

	err = stream.SendMsg(&in)

	if err != nil {
		fmt.Println(err)
	}

	err = stream.RecvMsg(&out)

	if err != nil {
		fmt.Println(err)
	}

	meta := stream.Trailer()

	fmt.Println(meta)

	fmt.Println(string(out.Data))

	ctx.Response.Header.Reset()
	ctx.Response.Header.Del("Content-Length")
	ctx.Response.Header.Set("access-control-allow-origin","*")
	ctx.Response.Header.Set("access-control-expose-headers","custom-header-1,grpc-status,grpc-message")
	ctx.Response.Header.SetContentType("application/grpc-web+proto")

	ctx.Response.SetBodyStream(bytes.NewReader(out.Data), len(out.Data))

	/*

	2
	BTCETHBTC *ETH0
	BTCUSDTBTC *USDT0grpc-status:0
	grpc-message:


	4
	2
	BTCETHBTC *ETH0
	BTCUSDTBTC *USDT0grpc-status:0
	grpc-message:

	*/

	//ctx.Response.AppendBodyString("grpc-status:0\ngrpc-message:\n")
}

func main() {
	router := fasthttprouter.New()
	router.GET("/", Index)
	router.OPTIONS("/proto.:service/:method", OptionsProto)
	router.POST("/proto.:service/:method", PostProto)

	log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}
