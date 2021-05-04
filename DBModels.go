package main

type IDBModel interface {
	MapToStruct(mapData map[string]interface{})
}

type Global_ShardInfo struct {
	Idx        int
	IP         string
	Port       int
	DBName     string
	AcceptUser byte
}

func (s *Global_ShardInfo) MapToStruct(mapData map[string]interface{}) {
	s.Idx = mapData["idx"].(int)
	s.IP = mapData["ip"].(string)
	s.Port = mapData["port"].(int)
	s.DBName = mapData["db_name"].(string)
	s.AcceptUser = mapData["accept_user"].(byte)
}
