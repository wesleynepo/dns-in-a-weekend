package main

import (
    "bytes"
    "encoding/binary"
    "log"
)

type DNSHeader struct {
    ID int 
    Flags int 
    NumQuestions int 
    NumAnswers int 
    NumAuthorities int 
    NumAdditionals int 
}

type DNSQuestion struct {
    name []byte
    Type int 
    Class int 
}

type DNSRecord struct {
    name []byte
    Type int
    Class int
    ttl int
    data []byte
}

type DNSPacket struct {
    header DNSHeader
    questions []DNSQuestion
    answers []DNSRecord
    authorities []DNSRecord
    additionals []DNSRecord
}

func (h *DNSHeader) toBytes() []byte {
    buf := new(bytes.Buffer)

    err := binary.Write(buf, binary.BigEndian, uint16(h.ID))

    if err != nil {
        log.Fatal(err)
    }

    err = binary.Write(buf, binary.BigEndian, uint16(h.Flags))

    if err != nil {
        log.Fatal(err)
    }

    err = binary.Write(buf, binary.BigEndian, uint16(h.NumQuestions))

    if err != nil {
        log.Fatal(err)
    }
    
    err = binary.Write(buf, binary.BigEndian, uint16(h.NumAnswers))

    if err != nil {
        log.Fatal(err)
    }
    
    err = binary.Write(buf, binary.BigEndian, uint16(h.NumAuthorities))

    if err != nil {
        log.Fatal(err)
    }

    err = binary.Write(buf, binary.BigEndian, uint16(h.NumAdditionals))

    if err != nil {
        log.Fatal(err)
    }

    return buf.Bytes()
}

func (q *DNSQuestion) toBytes() []byte {
    var b []byte

    buf := new(bytes.Buffer)

    err := binary.Write(buf, binary.BigEndian, uint16(q.Type))

    if err != nil {
        log.Fatal(err)
    }

    err = binary.Write(buf, binary.BigEndian, uint16(q.Class))

    if err != nil {
        log.Fatal(err)
    }

    b = append(b, q.name...)
    b = append(b, buf.Bytes()...)
    return b
}

