package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/examples/helloworld/helloworld"
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
	Data map[string]string
}

func (m Message) Reset() {}

func (m Message) ProtoMessage() {}

func (m Message) String() string {
	return proto.CompactTextString(m)
}

func Hello(ctx *fasthttp.RequestCtx) {
	fmt.Println(ctx, "hello, %s!\n", ctx.UserValue("service"), ctx.UserValue("method"))
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		fmt.Println(err)
		return
	}

	client := helloworld.NewGreeterClient(conn)

	reply, err := client.SayHello(context.Background(), &helloworld.HelloRequest{})
	
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(reply)

	grpc.
	//err = conn.Invoke(context.Background(), "/proto.Greeter/sayHello", &Message{}, &Reply{})
	//
	//if err != nil {
	//	fmt.Println(err)
	//}
}

func main() {
	router := fasthttprouter.New()
	router.GET("/", Index)
	router.OPTIONS("/proto.:service/:method", Hello)

	log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}