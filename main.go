package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"
	"strings"
)
const (
    RECURSION_DESIRED = 1 << 8
    CLASS_IN = 1
    TYPE_A = 1

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

func randomId() int {
    return rand.Intn(65535)
}

func buildQuery(domain string, recordType int) []byte {
    name := encodeName(domain)
    header := DNSHeader{
        ID:randomId(),
        Flags: RECURSION_DESIRED,
        NumQuestions: 1,
        NumAnswers: 0,
        NumAuthorities: 0,
        NumAdditionals: 0,
    }
    question := DNSQuestion{
        name: name,
        Class: CLASS_IN,
        Type: recordType,
    }

    b := make([]byte, 0)
    b = append(b, header.toBytes()...)
    b = append(b, question.toBytes()...)

    return b
}


func main() {
    query := buildQuery("www.example.com", 1)
    con, _ := net.Dial("udp", "8.8.8.8:53")
    con.Write(query)
}

