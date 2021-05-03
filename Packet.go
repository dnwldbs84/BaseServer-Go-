package main

import (
	"encoding/json"
	"net/http"
	"reflect"
)

const (
	PACKET_LOGIN = iota
)

var (
	CSPacketMap map[int]reflect.Type
	SCPacketMap map[int]reflect.Type
)

func InitPackets() {
	CSPacketMap = make(map[int]reflect.Type)
	SCPacketMap = make(map[int]reflect.Type)

	CSPacketMap[PACKET_LOGIN] = reflect.TypeOf(CSLoginStruct{})
	SCPacketMap[PACKET_LOGIN] = reflect.TypeOf(SCLoginStruct{})
}

type ICSPacket interface {
	ReadPacket(b []byte)
}
type ISCPacket interface {
	WritePacket(res http.ResponseWriter)
}

type CSLoginStruct struct {
	CSTest int
}

type SCLoginStruct struct {
	SCTest int
}

func (p *CSLoginStruct) ReadPacket(b []byte) {
	UnmarshalPacket(b, p)
}

func (p *SCLoginStruct) WritePacket(res http.ResponseWriter) {
	WritePacket(res, p)
}

// ICS 공용 함수
func UnmarshalPacket(b []byte, p interface{}) {
	err := json.Unmarshal(b, p)
	HandleErr(err)
}

// ISC 공용 함수
func WritePacket(res http.ResponseWriter, p interface{}) {
	b, err := json.Marshal(p)
	HandleErr(err)
	_, err = res.Write(b)
	HandleErr(err)
}
