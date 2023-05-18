package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

type DNSHeader struct {
    ID int 
    Flags int 
    NumQuestions int 
    NumAnswers int 
    NumAuthorities int 
    NumAdditionals int 
}

func (h *DNSHeader) toBytes() []byte {
    buf := new(bytes.Buffer)
    _ = binary.Write(buf, binary.BigEndian, uint16(h.ID))
    _ = binary.Write(buf, binary.BigEndian, uint16(h.Flags))
    _ = binary.Write(buf, binary.BigEndian, uint16(h.NumQuestions))
    _ = binary.Write(buf, binary.BigEndian, uint16(h.NumAnswers))
    _ = binary.Write(buf, binary.BigEndian, uint16(h.NumAuthorities))
    _ = binary.Write(buf, binary.BigEndian, uint16(h.NumAdditionals))
    return buf.Bytes()
}

type DNSQuestion struct {
    name []byte
    Type int 
    Class int 
}

func (q *DNSQuestion) toBytes() []byte {
    b := make([]byte, 0)
    buf := new(bytes.Buffer)
    _ = binary.Write(buf, binary.BigEndian, uint16(q.Type))
    _ = binary.Write(buf, binary.BigEndian, uint16(q.Class))
    b = append(b, q.name...)
    b = append(b, buf.Bytes()...)
    return b
}

func encodeName(domain string) []byte {
    buf := new(bytes.Buffer) 
    for _, part := range strings.Split(domain, ".") {
        binary.Write(buf, binary.BigEndian, uint8(len(part)))
        buf.Write([]byte(part))
    }
    binary.Write(buf, binary.BigEndian, uint8(0))
    return buf.Bytes() 
}


func main() {
    h := DNSHeader{43690, 1 << 8, 1, 0, 0, 0}
    fmt.Printf("%x\n", h.toBytes())
    db := encodeName("example.com")
    fmt.Printf("%x", db)
}

