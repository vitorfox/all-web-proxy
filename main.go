package main

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/metadata"
	"log"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
)

var (
	connDatasource *grpc.ClientConn
	connDataloader *grpc.ClientConn
	maxRecvSize = 1024 * 1024 * 1024
)

func Index(ctx *fasthttp.RequestCtx) {
	fmt.Println(ctx, "Welcome!\n")
}

type GrpcWebData struct {
	Header, Body []byte
}

func NewGrpcWebData(in []byte) *GrpcWebData {
	return &GrpcWebData{in[:5], in[5:]}
}

type Message struct {
	Data []byte
}

type Reply struct {
	Data []byte
}

func (m Reply) Reset() {
	m = Reply{}
}

func (m Reply) ProtoMessage() {}

func (m Reply) String() string {
	return fmt.Sprintln(m.Data)
}

func (m *Reply) XXX_Unmarshal(b []byte) error {
	m.Data = b
	return nil
}

func (m *Message) Reset()         {
	*m = Message{}
}
func (m *Message) String() string {
	return proto.CompactTextString(m)
}
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return []byte{0}, []int{0}
}

func (m *Message) XXX_Unmarshal(b []byte) error {
	m.Data = b
	return nil
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return m.Data, nil
}
func (m *Message) XXX_Merge(src proto.Message) {
	//xxx_messageInfo_DataloaderMeta.Merge(m, src)
	fmt.Println(src)
}
func (m *Message) XXX_Size() int {
	return len(m.Data)
}
func (m *Message) XXX_DiscardUnknown() {
	fmt.Println(m)
}

func WithHeader(in []byte) []byte{
	fmt.Println("actual len", len(in))
	header := make([]byte, 5)
	binary.BigEndian.PutUint32(header[1:], uint32(len(in)))

	fmt.Println("actual header", header)
	return append(header, in...)
}

func WithHeaderStatus(in []byte) []byte{
	return append([]byte{128,0,0,0,uint8(len(in))}, in...)
}

func OptionsProto(ctx *fasthttp.RequestCtx) {

	method := fmt.Sprintf("/proto.%s/%s", ctx.UserValue("service"), ctx.UserValue("method"))

	log.Println(string(ctx.Method()), method)

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
	grpcWebData := NewGrpcWebData(ctx.PostBody())

	log.Println(string(ctx.Method()), method)
	log.Println(grpcWebData)

	conn := &grpc.ClientConn{}
	switch ctx.UserValue("service") {
	case "DatasourceService":
		conn = connDatasource
	case "Dataloader":
		conn = connDataloader
	default:
		return
	}

	var header, trailer metadata.MD // variable to store header and trailer

	in := Message{grpcWebData.Body}
	out := Reply{}

	err := conn.Invoke(ctx, method, &in, &out,
		grpc.Header(&header),
		grpc.Trailer(&trailer),
		grpc.MaxCallRecvMsgSize(maxRecvSize))

	if err != nil {
		fmt.Println(err)
		return
	}

	var grpcMessage string
	var grpcStatus int

	fmt.Println(header)
	fmt.Println(trailer)
	fmt.Println(out.Data)

	ctx.Response.Header.Reset()
	ctx.Response.Header.Del("Content-Length")
	ctx.Response.Header.Set("access-control-allow-origin","*")
	ctx.Response.Header.Set("access-control-expose-headers","custom-header-1,grpc-status,grpc-message")
	ctx.Response.Header.SetContentType("application/grpc-web+proto")

	ctx.Response.SetBody(WithHeader(out.Data))
	status := fmt.Sprintf("grpc-status:%d\r\ngrpc-message:%s\r\n", grpcStatus, grpcMessage)
	ctx.Response.AppendBody(WithHeaderStatus([]byte(status)))

}

func main() {

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	connDatasource = conn

	conn, err = grpc.Dial("localhost:50055", grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	connDataloader = conn

	router := fasthttprouter.New()
	router.GET("/", Index)
	router.OPTIONS("/proto.:service/:method", OptionsProto)
	router.POST("/proto.:service/:method", PostProto)

	log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}
