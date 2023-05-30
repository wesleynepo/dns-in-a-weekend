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

func ipString(data []byte) string {
    return fmt.Sprintf("%v.%v.%v.%v", data[0], data[1], data[2], data[3])
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
    var name string
    var i int 
    for  i < len(encoded) {
        many := int(encoded[i])
        if (i != 0 && many != 0) {
            name += "."
        }
        name += (string(encoded[i+1:i+1+many]))
        i += many + 1
    }
    return name
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

func main() {
    ip := resolveIp("twitter.com")
    print(ip)
}

func resolveIp(domain string) string {
    var packet DNSPacket
    ip := "198.41.0.4"

    print(domain + " -> ")

    for true {
        packet = run(domain, ip)
        
        if (len(packet.answers) != 0) {
            break
        }

        if (len(packet.additionals) == 0) {
            nameServer := decodeName(packet.authorities[0].data)
            ip = resolveIp(nameServer)
        } else {
            authority := getFirstIPV4(packet.additionals) 
            ip = ipString(authority.data)
        }
        print(ip + " -> ")
    }

    return ipString(packet.answers[0].data)
}

func getFirstIPV4(additionals []DNSRecord) DNSRecord {
    for _, a := range additionals {
        if (a.Type == TYPE_A) {
            return a
        }
    }

    return DNSRecord{}
}

func run(domain string, ip string) DNSPacket {
    query := buildQuery(domain, TYPE_A)

    con, err := net.Dial("udp", ip + ":53")
    defer con.Close()

    if err != nil {
        log.Fatal(err)
    }

    con.Write(query)

    reply := make([]byte, 1024)

    _, err = con.Read(reply)

    if err != nil {
        log.Fatal(err)
    }

    response := ResponseRead{reply, 0}

    return response.Parse() 
}
