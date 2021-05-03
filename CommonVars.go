package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"sync"

	"github.com/go-redis/redis"
)

// Config Values
const (
	GlobalDBMaxConnCount  = 50
	GlobalDBIdleConnCount = 10
	GameDBMaxConnCount    = 50
	GameDBIdleConnCount   = 30
)

// Enums
const (
	EnumLogLevel_Info = EnumLogLevelValue(0)
	EnumLogLevel_Err  = EnumLogLevelValue(1)

	EnumErr_System = EnumErrValue(0)

	EnumDBType_Byte   = EnumDBTypeValue(0)
	EnumDBType_Int    = EnumDBTypeValue(1)
	EnumDBType_Long   = EnumDBTypeValue(2)
	EnumDBType_String = EnumDBTypeValue(3)
	EnumDBType_Time   = EnumDBTypeValue(4)
	EnumDBType_Float  = EnumDBTypeValue(5)
	EnumDBType_Err    = EnumDBTypeValue(99)
)

var Config ConfigInfo

var (
	LogLock        sync.Mutex
	CurLogFileName string
	CurLogFile     *os.File

	LogInfoLogger  *log.Logger
	LogErrorLogger *log.Logger

	GlobalDB  *sql.DB
	GameDBMap map[int]*sql.DB

	RedisClient  *redis.Client
	RedisContext context.Context
)
