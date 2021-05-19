package main

type EnumDBTypeValue int
type EnumLogLevelValue int
type EnumErrValue int
type EnumSeqErrValue int

type ConfigInfo struct {
	Server struct {
		IP   string
		Port int
	}
	GlobalDB struct {
		User     string
		Password string
		IP       string
		Port     int
		DBName   string
	}
	Redis struct {
		Addr     string
		Password string
	}
}

type RedisSeqInfo struct {
	Seq    int
	ResVal []byte
}
