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
    TYPE_NS = 2
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
    fields :=  make([]int, 0)
    for i := 0; i < len(header); i+=2 {
        value := binary.BigEndian.Uint16(header[i:i+2])
        fields = append(fields, int(value))
    }

    return DNSHeader{fields[0], fields[1], fields[2], fields[3], fields[4], fields[5]}
}

func newHeaderFromResponse(response *ResponseRead) DNSHeader {
    fields :=  make([]int, 0)
    for i := 0; i < 12; i+=2 {
        value := binary.BigEndian.Uint16(response.getSlice(2))
        fields = append(fields, int(value))
    }
    return DNSHeader{fields[0], fields[1], fields[2], fields[3], fields[4], fields[5]}
}

type DNSQuestion struct {
    name []byte
    Type int 
    Class int 
}

func newQuestionFromResponse(response *ResponseRead) DNSQuestion { 
    currentSlice := response.currentSlice()
    nameSize, _ := findNameSize(currentSlice)
    name := response.getSlice(nameSize+1) 
    type_ := binary.BigEndian.Uint16(response.getSlice(2))
    class := binary.BigEndian.Uint16(response.getSlice(2))

    return DNSQuestion{name, int(type_), int(class)}
}

func newQuestionFromBytes(question []byte, nameSize int) DNSQuestion { 
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

func decodeName(encoded []byte) string {
    println(string(encoded))
    return ""
}

func randomId() int {
    return rand.Intn(65535)
}

func buildQuery(domain string, recordType int) []byte {
    name := encodeName(domain)
    header := DNSHeader{
        ID:randomId(),
        Flags: 0,
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

func getNameFromByte(response []byte, offset int) []byte {
    nameStart := response[offset:]
    nameSize , _ := findNameSize(nameStart)
    return response[offset:offset+nameSize]
}

func newRecordFromBytes(response []byte, nameSize int, start int) DNSRecord {
    record := response[start:]
    nameIndicaton := int(record[:nameSize][1])
    name := getNameFromByte(response, nameIndicaton)
    type_ := binary.BigEndian.Uint16(record[nameSize:nameSize+2])
    class := binary.BigEndian.Uint16(record[nameSize+2:nameSize+4])
    ttl := binary.BigEndian.Uint32(record[nameSize+4:nameSize+8])
    byteLen := binary.BigEndian.Uint16(record[nameSize+8:nameSize+10])
    data :=record[nameSize+10:nameSize+10+int(byteLen)]
    return DNSRecord{name, int(type_), int(class), int(ttl), data} 
}

func newRecordFromResponse(response *ResponseRead) DNSRecord {
    var name []byte
    record := response.currentSlice()
    nameSize, _ := findNameSize(record)

    if (nameSize == 2) {
      nameIndicaton := int(record[:nameSize][1])
      name = getNameFromByte(response.data, nameIndicaton)
    } else {
      name = record[:nameSize+1]
    }
    type_ := binary.BigEndian.Uint16(record[nameSize:nameSize+2])
    class := binary.BigEndian.Uint16(record[nameSize+2:nameSize+4])
    ttl := binary.BigEndian.Uint32(record[nameSize+4:nameSize+8])
    byteLen := binary.BigEndian.Uint16(record[nameSize+8:nameSize+10])
    data :=record[nameSize+10:nameSize+10+int(byteLen)]
    response.movePointer(nameSize+10+int(byteLen))
    // move the pointer throught the data after parsing the record
    return DNSRecord{name, int(type_), int(class), int(ttl), data} 
}



func parseRecord(response []byte) {
    header := newHeaderFromBytes(response[:12])
    fmt.Println(header)
    nameSize, _ := findNameSize(response[12:])
    _ = newQuestionFromBytes(response[12:12+nameSize+5], nameSize)
    nameSize2, _ := findNameSize(response[12+nameSize+5:])
    start := 12+nameSize+5
    record := newRecordFromBytes(response, nameSize2, start)
    println(ipString(record.data))

}

func ipString(data []byte) string {
    return fmt.Sprintf("%v.%v.%v.%v", data[0], data[1], data[2], data[3])
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
    var packet DNSPacket
    domain := "google.com"
    ip := "198.41.0.4"

    for true {
        packet = run(domain, ip)
        
        if (len(packet.answers) != 0) {
            break
        }
        authority := getFirstIPV4(packet.additionals) 
        ip = ipString(authority.data)
        println(ip + " ->")
    }

    println(ipString(packet.answers[0].data))
    println("Finish")
}

func getFirstIPV4(additionals []DNSRecord) DNSRecord {
    for _, a := range additionals {
        if (a.Type == 1) {
            return a
        }
    }

    return DNSRecord{}
}

func run(domain string, ip string) DNSPacket {
    query := buildQuery(domain, TYPE_A)

    con, err := net.Dial("udp", ip + ":53")
    
    checkErr(err)

    con.Write(query)

    reply := make([]byte, 1024)

    _, err = con.Read(reply)

    checkErr(err)
    con.Close()

    return parse(reply)
}

func checkErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

type DNSPacket struct {
    header DNSHeader
    questions []DNSQuestion
    answers []DNSRecord
    authorities []DNSRecord
    additionals []DNSRecord
}

type ResponseRead struct {
    data []byte
    pointer int
}

func (r *ResponseRead) movePointer(sizes int) {
    r.pointer += sizes
}

func (r *ResponseRead) currentSlice() []byte {
    return r.data[r.pointer:]
}

func (r *ResponseRead) getSlice(width int) []byte {
    defer r.movePointer(width)
    return r.data[r.pointer:r.pointer+width]
}

func parse(data []byte) DNSPacket {
    response := &ResponseRead{data, 0}
    packet := DNSPacket{}
    packet.header = newHeaderFromResponse(response)

    for i := 0; i < packet.header.NumQuestions; i++ {
        q := newQuestionFromResponse(response)
        packet.questions = append(packet.questions, q)
    }
    
    for i := 0; i < packet.header.NumAnswers; i++ {
        a := newRecordFromResponse(response)
        packet.answers = append(packet.answers, a)
    }

    for i := 0; i < packet.header.NumAuthorities; i++ {
        a := newRecordFromResponse(response)
        packet.authorities = append(packet.authorities, a)
    }

    for i := 0; i < packet.header.NumAdditionals; i++ {
        a := newRecordFromResponse(response)
        packet.additionals = append(packet.additionals, a)
    }

    return packet
}
