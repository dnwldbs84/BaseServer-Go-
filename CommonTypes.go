package main

type EnumLogLevelValue int
type EnumErrValue int
type EnumDBTypeValue int

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
