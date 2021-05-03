package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func main() {
	InitSystem()
	InitConfig()
	InitCommonVars()
	InitDesignTables()
	InitPackets()
	InitDB()
	InitRedis()

	http.HandleFunc("/", CommonHandler)

	WriteLog(EnumLogLevel_Info, " Server Starting !!!")

	err := http.ListenAndServe(Config.Server.IP+":"+strconv.Itoa(Config.Server.Port), nil)
	HandleErr(err)
}

func CommonHandler(res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			WriteLog(EnumLogLevel_Err, "	Panic Recovered !!! - ", r)
		}
	}()
	if req.Method == "POST" {
		packetType, err := strconv.Atoi(req.Header.Get("PT"))
		HandleErr(err)

		b, err := ioutil.ReadAll(req.Body)
		HandleErr(err)
		defer req.Body.Close()

		csType, result := CSPacketMap[packetType]
		if !result {
			HandleErr(errors.New("ERR_PACKET_TYPE: " + strconv.Itoa(packetType)))
			return
		}

		csPacket := reflect.New(csType).Interface().(ICSPacket)
		csPacket.ReadPacket(b)
		var scPacket ISCPacket

		switch packetType {
		case PACKET_LOGIN:
			response := PacketHandleLogin(csPacket.(*CSLoginStruct))
			scPacket = ISCPacket(&response)
		}

		scPacket.WritePacket(res)
	} else if req.Method == "GET" {
		var response []string
		for key := range CSPacketMap {
			response = append(response, CSPacketMap[key].Name())
			response = append(response, SCPacketMap[key].Name())
		}
		_, err := res.Write([]byte(strings.Join(response, "\n")))
		HandleErr(err)
	}
}
