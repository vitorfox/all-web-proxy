package main

import (
	"encoding/binary"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"google.golang.org/grpc/metadata"
	"log"

	internalConfig "github.com/vitorfox/all-web-proxy/internal/config"
	internalProto "github.com/vitorfox/all-web-proxy/internal/proto"

	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
)

var (
	maxRecvSize = 1024 * 1024 * 1024
	grpcUpstreamConnections = make(map[string]*grpc.ClientConn)
	config internalConfig.All
)

type GrpcWebData struct {
	Header, Body []byte
}

func NewGrpcWebData(in []byte) *GrpcWebData {
	return &GrpcWebData{in[:5], in[5:]}
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

func ProtoHandler(ctx *fasthttp.RequestCtx, upstream *internalConfig.Upstream) {

	m := string(ctx.Method())

	fmt.Println("ProtoHandler method", m)
	switch m {
	case "OPTIONS":
		OptionsProto(ctx)
	case "POST":
		if conn, ok := grpcUpstreamConnections[upstream.Name]; ok {
			PostProto(ctx, conn)
		}
	}
}

func OptionsProto(ctx *fasthttp.RequestCtx) {

	method := string(ctx.Path())
	log.Println(string(ctx.Method()), method)

	ctx.Response.Header.Set("Access-Control-Allow-Origin","*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods","GET, POST, OPTIONS")
	ctx.Response.Header.Set("Access-Control-Allow-Headers","DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Content-Transfer-Encoding,Custom-Header-1,X-Accept-Content-Transfer-Encoding,X-Accept-Response-Streaming,X-User-Agent,X-Grpc-Web")
	ctx.Response.Header.Set("Access-Control-Max-Age","1728000")
	ctx.Response.Header.SetContentType("text/plain charset=UTF-8")
	ctx.Response.Header.SetContentLength(0)
	ctx.Response.SetStatusCode(204)
}

func PostProto(ctx *fasthttp.RequestCtx, conn *grpc.ClientConn) {

	grpcWebData := NewGrpcWebData(ctx.PostBody())
	method := string(ctx.Path())
	log.Println(string(ctx.Method()), method)
	log.Println(grpcWebData)

	var header, trailer metadata.MD // variable to store header and trailer

	in := internalProto.Message{grpcWebData.Body}
	out := internalProto.Message{}

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

	ctx.Response.Header.Reset()
	ctx.Response.Header.Del("Content-Length")
	ctx.Response.Header.Set("access-control-allow-origin","*")
	ctx.Response.Header.Set("access-control-expose-headers","custom-header-1,grpc-status,grpc-message")
	ctx.Response.Header.SetContentType("application/grpc-web+proto")

	ctx.Response.SetBody(WithHeader(out.Data))
	status := fmt.Sprintf("grpc-status:%d\r\ngrpc-message:%s\r\n", grpcStatus, grpcMessage)
	ctx.Response.AppendBody(WithHeaderStatus([]byte(status)))

}

func GenericHttp(ctx *fasthttp.RequestCtx, upstream *internalConfig.Upstream) {

	resp := fasthttp.AcquireResponse()
	method := string(ctx.Path())

	log.Println(string(ctx.Method()), method)

	req := ctx.Request

	req.SetHost(upstream.Address)

	err := fasthttp.Do(&req, resp)

	if err != nil {
		fmt.Println(err)
		return
	}

	ctx.Response.SetBody(resp.Body())
}

func setupUpstreams(config internalConfig.All) {
	for _, up := range config.Upstream {
		if up.Type == internalConfig.UPSTREAM_GRPC {
			conn, err := grpc.Dial(up.Address, grpc.WithInsecure())
			if err != nil {
				log.Println(err)
				continue
			}
			grpcUpstreamConnections[up.Name] = conn
		}
	}
}

func handlerWithListener(listener *internalConfig.Listener) func (ctx *fasthttp.RequestCtx) {

	funcc := func (ctx *fasthttp.RequestCtx) {
		var route *internalConfig.Route

		for _, vh := range listener.Virtualhosts {
			if route != nil {
				break
			}
			for _, vhr := range vh.Routes {
				fmt.Println(vhr)
				if vhr.Match.CompiletedRegex != nil {
					if vhr.Match.CompiletedRegex.Match(ctx.Path()) {
						route = &vhr
						break
					}
				}
			}
		}

		if route != nil {
			for _, up := range config.Upstream {
				if up.Name == route.Upstream {
					if up.Type == internalConfig.UPSTREAM_GRPC {
						ProtoHandler(ctx, &up)
					}
					if up.Type == internalConfig.UPSTREAM_HTTP {
						GenericHttp(ctx, &up)
					}
					break
				}
			}
		}
	}

	return funcc
}

func main() {

	data, err := ioutil.ReadFile("config.yaml")

	if err != nil {
		log.Fatal(err)
	}

	tmpConfig := internalConfig.All{}

	err = yaml.Unmarshal(data, &tmpConfig)

	if err != nil {
		log.Fatal("Could not load config.", err)
	}

	config = tmpConfig
	fmt.Println(config)

	setupUpstreams(config)

	for _, li := range config.Listeners {
		log.Fatal(fasthttp.ListenAndServe(li.Address, handlerWithListener(&li)))
	}
}
