package main

import (
	"encoding/json"
	"reflect"
)

type EnumPacketValue int

const (
	PACKET_CHECK_SERVER = EnumPacketValue(0)
	PACKET_LOGIN        = EnumPacketValue(1)
	PACKET_TEST         = EnumPacketValue(2)
)

var (
	CSPacketMap map[EnumPacketValue]reflect.Type
	SCPacketMap map[EnumPacketValue]reflect.Type
)

func InitPackets() {
	CSPacketMap = make(map[EnumPacketValue]reflect.Type)
	SCPacketMap = make(map[EnumPacketValue]reflect.Type)

	CSPacketMap[PACKET_CHECK_SERVER] = reflect.TypeOf(CSEmptyStruct{})
	SCPacketMap[PACKET_CHECK_SERVER] = reflect.TypeOf(SCEmptyStruct{})

	CSPacketMap[PACKET_LOGIN] = reflect.TypeOf(CSLoginStruct{})
	SCPacketMap[PACKET_LOGIN] = reflect.TypeOf(SCLoginStruct{})

	CSPacketMap[PACKET_TEST] = reflect.TypeOf(CSTestStruct{})
	SCPacketMap[PACKET_TEST] = reflect.TypeOf(SCTestStruct{})
}

type ICSPacket interface {
	ReadPacket(b []byte)
}
type ISCPacket interface {
	Marshal() []byte
	// WritePacket(res http.ResponseWriter)
}

// ICS 공용 함수
func UnmarshalPacket(b []byte, p interface{}) {
	err := json.Unmarshal(b, p)
	HandleErr(err)
}

// ISC 공용 함수
func MarshalPacket(p interface{}) []byte {
	b, err := json.Marshal(p)
	HandleErr(err)
	return b
}

// func WritePacket(res http.ResponseWriter, p interface{}) {
// 	b, err := json.Marshal(p)
// 	HandleErr(err)
// 	_, err = res.Write(b)
// 	HandleErr(err)
// }

type CSEmptyStruct struct {
}

type SCEmptyStruct struct {
}

func (p *CSEmptyStruct) ReadPacket(b []byte) {
	UnmarshalPacket(b, p)
}
func (p *SCEmptyStruct) Marshal() []byte {
	return MarshalPacket(p)
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
func (p *SCLoginStruct) Marshal() []byte {
	return MarshalPacket(p)
}

type CSTestStruct struct {
}
type SCTestStruct struct {
	UID int64
}

func (p *CSTestStruct) ReadPacket(b []byte) {
	UnmarshalPacket(b, p)
}
func (p *SCTestStruct) Marshal() []byte {
	return MarshalPacket(p)
}
