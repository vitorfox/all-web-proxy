package proto

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

type Message struct {
	Data []byte
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

func (m *Message) XXX_Size() int {
	return len(m.Data)
}

func (m *Message) XXX_DiscardUnknown() {
	fmt.Println(m)
}
