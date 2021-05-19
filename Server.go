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
			WriteLog(EnumLogLevel_Err, "Panic Recovered !!! - ", r)
			res.Header().Set(HeaderErrCode, strconv.Itoa(int(EnumErr_System)))
		}
	}()

	if req.Method == "POST" {
		// Header 정보
		pt, err := strconv.Atoi(req.Header.Get(HeaderPacketType))
		HandleErr(err)
		packetType := EnumPacketValue(pt)

		var cSeq int
		var uid int64
		var resVal []byte
		if packetType > PACKET_LOGIN {
			cSeq, err = strconv.Atoi(req.Header.Get(HeaderSequence))
			HandleErr(err)
			uid, err = strconv.ParseInt(req.Header.Get(HeaderUID), 10, 64)
			HandleErr(err)
			if uid == 0 {
				HandleErr(errors.New("Can't find uid from header."))
			}

			// Sequence 체크
			resVal, seqErrCode := CheckSequence(RedisClient, uid, cSeq)
			if seqErrCode > 0 {
				if seqErrCode == EnumSeqErr_InProgress {
					return
				} else if seqErrCode == EnumSeqErr_DuplRequest {
					_, err = res.Write(resVal)
					HandleErr(err)
					return
				}
			}
			defer UpdateSequence(RedisClient, uid, cSeq, &resVal)
		}

		// Body 정보
		b, err := ioutil.ReadAll(req.Body)
		HandleErr(err)
		defer req.Body.Close()

		csType, result := CSPacketMap[packetType]
		if !result {
			HandleErr(errors.New("ERR_PACKET_TYPE: " + strconv.Itoa(int(packetType))))
			return
		}

		csPacket := reflect.New(csType).Interface().(ICSPacket)
		csPacket.ReadPacket(b)

		// Route 함수 처리
		var scPacket ISCPacket
		var errCode EnumErrValue
		switch packetType {
		case PACKET_CHECK_SERVER:
			response, errorCode := PacketHandleCheckServer(csPacket.(*CSEmptyStruct))

			errCode = errorCode
			scPacket = ISCPacket(&response)
		case PACKET_LOGIN:
			response, errorCode := PacketHandleLogin(csPacket.(*CSLoginStruct))

			errCode = errorCode
			scPacket = ISCPacket(&response)
		}

		// Header Set
		res.Header().Set(HeaderPacketType, strconv.Itoa(int(packetType)))
		res.Header().Set(HeaderErrCode, strconv.Itoa(int(errCode)))
		res.Header().Set(HeaderSequence, "")
		// Body Set
		resVal = scPacket.Marshal()
		_, err = res.Write(resVal)
		HandleErr(err)
		// scPacket.WritePacket(res)
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
