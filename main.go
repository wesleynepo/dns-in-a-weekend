package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
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

func pb(b []byte, s string) {
    fmt.Printf("%s : %x\n",s, b)
}

func newHeaderFromBytes(header []byte) DNSHeader {
    pb(header, "header")
    fields :=  make([]int, 0)
    for i := 0; i < len(header); i+=2 {
        value := binary.BigEndian.Uint16(header[i:i+2])
        fields = append(fields, int(value))
    }

    return DNSHeader{fields[0], fields[1], fields[2], fields[3], fields[4], fields[5]}
}

type DNSQuestion struct {
    name []byte
    Type int 
    Class int 
}

func newQuestionFromBytes(question []byte, nameSize int) DNSQuestion { 
    pb(question, "response question")
    name := question[:nameSize+1]
    type_ := binary.BigEndian.Uint16(question[nameSize+1:nameSize+3])
    class := binary.BigEndian.Uint16(question[nameSize+3:nameSize+5])

    return DNSQuestion{name, int(type_), int(class)}
}

func (q *DNSQuestion) toBytes() []byte {
    b := make([]byte, 0)
    buf := new(bytes.Buffer)
    _ = binary.Write(buf, binary.BigEndian, uint16(q.Type))
    _ = binary.Write(buf, binary.BigEndian, uint16(q.Class))
    b = append(b, q.name...)
    b = append(b, buf.Bytes()...)
    pb(b, "question request")
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

type DNSRecord struct {
    name []byte
    Type int
    Class int
    ttl int
    data []byte
}

func newRecordFromBytes(record []byte, nameSize int) DNSRecord {
    pb(record, "record")
    println(nameSize)
    name := record[:nameSize]
    pb(name, "name")
    type_ := binary.BigEndian.Uint16(record[nameSize:nameSize+2])
    class := binary.BigEndian.Uint16(record[nameSize+2:nameSize+4])
    ttl := binary.BigEndian.Uint32(record[nameSize+4:nameSize+8])
    byteLen := binary.BigEndian.Uint16(record[nameSize+8:nameSize+10])
    println(byteLen)
    pb(record[nameSize+10:], "data")
    data :=record[nameSize+10:nameSize+10+int(byteLen)]
    pb(data, "data") 
    fmt.Printf("%v", data)

    return DNSRecord{name, int(type_), int(class), int(ttl), data} 
}

func parseRecord(response []byte) {
    // heade + question + response ?
    pb(response, "response")
    _ = newHeaderFromBytes(response[:12])
    nameSize, _ := findNameSize(response[12:])
    _ = newQuestionFromBytes(response[12:12+nameSize+5], nameSize)
    nameSize2, _ := findNameSize(response[12+nameSize+5:])
    _ = newRecordFromBytes(response[12+nameSize+5:], nameSize2)
}

func findNameSize(b []byte) (int, error) {
    for i, b := range b {
        if b == 0 {
            return i, nil
        }
    }
    return 0, nil
}


func main() {
    query := buildQuery("www.example.com", 1)

    con, err := net.Dial("udp", "8.8.8.8:53")
    
    checkErr(err)

    con.Write(query)

    reply := make([]byte, 1024)

    _, err = con.Read(reply)

    checkErr(err)

    parseRecord(reply)
}

func checkErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

