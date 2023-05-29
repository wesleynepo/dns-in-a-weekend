package main

import (
	"encoding/binary"
	"fmt"
)

const (
    UINT16_SIZE = 2
    UINT32_SIZE = 4
)

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

func findNameSize(b []byte) (int) {
    for i, b := range b {
        if b == 0 {
            return i
        }
    }
    return 0
}

func ipString(data []byte) string {
    return fmt.Sprintf("%v.%v.%v.%v", data[0], data[1], data[2], data[3])
}


func getNameFromByte(response []byte, offset int) []byte {
    nameStart := response[offset:]
    nameSize := findNameSize(nameStart)
    return response[offset:offset+nameSize]
}

func getNameFromResponse(response *ResponseRead, nameSize int) []byte {
    var name []byte
    index := 0
    name = append(name, response.currentSlice()[:nameSize+1]...)
    for spot, b := range name {
        if b == 192 {
            index = int(name[spot+1])
        }
    }
    if index != 0 {
        name = append(name[:len(name)-3], getNameFromByte(response.data, index)...)
    }
    return name
}

func (r *ResponseRead) parseRecord() DNSRecord {
    var data []byte
    record := r.currentSlice()
    nameSize := findNameSize(record)
    name := getNameFromResponse(r, nameSize)
    r.movePointer(nameSize)
    
    type_ := r.readInt() 
    class := r.readInt() 
    ttl := r.readInt32() 
    byteLen := r.readInt() 
    if type_ == TYPE_NS {
        data = getNameFromResponse(r, int(byteLen)-1)
    } else {
        data =record[nameSize+10:nameSize+10+int(byteLen)]
    }
    r.movePointer(int(byteLen))
    return DNSRecord{name, type_, class, ttl, data} 
}

func (r *ResponseRead) parseHeader() DNSHeader {
    return DNSHeader{
        ID: r.readInt(), 
        Flags: r.readInt(),
        NumQuestions: r.readInt(),
        NumAnswers: r.readInt(), 
        NumAuthorities: r.readInt(),
        NumAdditionals: r.readInt(),
    }
}

func (r *ResponseRead) readInt() int {
    b := r.getSlice(UINT16_SIZE) 
    return int(binary.BigEndian.Uint16(b))
}

func (r *ResponseRead) readInt32() int {
    b := r.getSlice(UINT32_SIZE) 
    return int(binary.BigEndian.Uint16(b))
}

func (r *ResponseRead) parseQuestion() DNSQuestion { 
    currentSlice := r.currentSlice()
    nameSize := findNameSize(currentSlice)
    name := r.getSlice(nameSize+1) 
    type_ := r.readInt() 
    class := r.readInt() 

    return DNSQuestion{name, type_, class}
}


func (r *ResponseRead) Parse() DNSPacket {
    packet := DNSPacket{}
    packet.header = r.parseHeader()

    for i := 0; i < packet.header.NumQuestions; i++ {
        q := r.parseQuestion()
        packet.questions = append(packet.questions, q)
    }
    
    for i := 0; i < packet.header.NumAnswers; i++ {
        a := r.parseRecord()
        packet.answers = append(packet.answers, a)
    }

    for i := 0; i < packet.header.NumAuthorities; i++ {
        a := r.parseRecord()
        packet.authorities = append(packet.authorities, a)
    }

    for i := 0; i < packet.header.NumAdditionals; i++ {
        a := r.parseRecord()
        packet.additionals = append(packet.additionals, a)
    }

    return packet
}
