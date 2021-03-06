package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

func InitSystem() {
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		err = os.Mkdir("./logs", os.FileMode(0644))
		HandleErr(err)
	}
}
func InitConfig() {
	file, err := ioutil.ReadFile("ServerConfig.json")
	HandleErr(err)

	err = json.Unmarshal([]byte(file), &Config)
	HandleErr(err)
}

func InitCommonVars() {
}

func InitDB() {
	var err error

	// Global DB
	dbInfo := Config.GlobalDB
	connStr := dbInfo.User + ":" + dbInfo.Password + "@(" + dbInfo.IP + ":" + strconv.Itoa(dbInfo.Port) + ")/" + dbInfo.DBName + "?parseTime=true"
	GlobalDB, err = GetDBConnection(connStr, GlobalDBMaxConnCount, GlobalDBIdleConnCount)
	HandleErr(err)

	// Game DB
	GameDBMap = make(map[int]*sql.DB)
	resultList := DBQuery(GlobalDB, "SELECT idx, ip, port, db_name, accept_user FROM tbl_shard_info;", &Global_ShardInfo{})
	for _, shardInfo := range resultList {
		info := *(shardInfo).(*Global_ShardInfo)
		connStr := dbInfo.User + ":" + dbInfo.Password + "@(" + info.IP + ":" + strconv.Itoa(info.Port) + ")/" + info.DBName + "?parseTime=true"
		GameDB, err := GetDBConnection(connStr, GameDBMaxConnCount, GameDBIdleConnCount)
		HandleErr(err)
		GameDBMap[info.Idx] = GameDB
	}
}

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     Config.Redis.Addr,
		Password: Config.Redis.Password,
		DB:       0,
	})

	RedisContext = context.Background()
}

// 공용 오류 처리 함수
func HandleErr(err error) {
	if err != nil {
		stack := string(debug.Stack())
		WriteLog(EnumLogLevel_Err, err.Error(), "\n\n", stack)
		if GlobalDB != nil {
			// DB 로그
			var queryStr []string
			queryStr = append(queryStr,
				"INSERT INTO tbl_error_log(err_msg, err_stack) VALUES('",
				err.Error(),
				"','",
				stack,
				"');",
			)
			_, dbErr := GlobalDB.Exec(strings.Join(queryStr, ""))
			if dbErr != nil {
				WriteLog(EnumLogLevel_Err, dbErr.Error())
			}
		}
		panic(err)
	}
}

// 파일 오류 로그
func WriteLog(level EnumLogLevelValue, args ...interface{}) {
	t := time.Now()
	fileName := t.Format("./logs/2006-01-02_15_Log.log")

	LogLock.Lock()
	defer LogLock.Unlock()

	if fileName != CurLogFileName {
		if CurLogFile != nil {
			CurLogFile.Close()
		}

		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Println("os.OpenFile Error: ", err.Error())
			return
		}

		LogInfoLogger = log.New(io.MultiWriter(file, os.Stdout), "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
		LogErrorLogger = log.New(io.MultiWriter(file, os.Stdout), "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

		CurLogFile = file
		CurLogFileName = fileName
	}

	if level == EnumLogLevel_Info {
		LogInfoLogger.Println(args...)
	} else if level == EnumLogLevel_Err {
		LogErrorLogger.Println(args...)
	}
}

func GetDBConnection(connStr string, maxConnCount int, idleConnCount int) (db *sql.DB, err error) {
	db, err = sql.Open("mysql", connStr)
	HandleErr(err)

	db.SetMaxOpenConns(maxConnCount)
	db.SetMaxIdleConns(idleConnCount)

	return
}

// 다중 row DB 처리 함수
func DBQuery(db *sql.DB, query string, dbModel IDBModel) (result []IDBModel) {
	rows, err := db.Query(query)
	HandleErr(err)
	defer rows.Close()

	colTypes, err := rows.ColumnTypes()
	HandleErr(err)
	// for _, colType := range colTypes {
	// 	fmt.Println("colType: ", colType.DatabaseTypeName())
	// }

	colNames, err := rows.Columns()
	HandleErr(err)

	colPtrs := make([]interface{}, len(colNames))

	for i := 0; i < len(colNames); i++ {
		// if isNull, isOk := colTypes[i].Nullable(); isNull && isOk {
		// 	// 필요한 경우 추가
		// }
		colType := colTypes[i].DatabaseTypeName()
		dbType := DBTypeToGoType(colType)
		switch dbType {
		case EnumDBType_Byte:
			var data byte
			colPtrs[i] = &data
		case EnumDBType_Int:
			var data int
			colPtrs[i] = &data
		case EnumDBType_Long:
			var data int64
			colPtrs[i] = &data
		case EnumDBType_String:
			var data string
			colPtrs[i] = &data
		case EnumDBType_Time:
			var data time.Time
			colPtrs[i] = &data
		case EnumDBType_Float:
			var data float32
			colPtrs[i] = &data
		}
	}

	for rows.Next() {
		var resultMap = make(map[string]interface{})
		err = rows.Scan(colPtrs...)
		HandleErr(err)

		for i, col := range colPtrs {
			colType := colTypes[i].DatabaseTypeName()
			dbType := DBTypeToGoType(colType)
			switch dbType {
			case EnumDBType_Byte:
				resultMap[colNames[i]] = *col.(*byte)
			case EnumDBType_Int:
				resultMap[colNames[i]] = *col.(*int)
			case EnumDBType_Long:
				resultMap[colNames[i]] = *col.(*int64)
			case EnumDBType_String:
				resultMap[colNames[i]] = *col.(*string)
			case EnumDBType_Time:
				resultMap[colNames[i]] = *col.(*time.Time)
			case EnumDBType_Float:
				resultMap[colNames[i]] = *col.(*float32)
			}
		}

		dbModelType := reflect.New(reflect.TypeOf(dbModel).Elem())
		dbStruct, _ := dbModelType.Interface().(IDBModel)
		dbStruct.MapToStruct(resultMap)

		result = append(result, dbStruct)
	}
	return result
}

// 단일 row DB 처리 함수
func DBQueryRow(db *sql.DB, query string, dbModel IDBModel) {
	rows, err := db.Query(query)
	HandleErr(err)
	defer rows.Close()

	colTypes, err := rows.ColumnTypes()
	HandleErr(err)
	colNames, err := rows.Columns()
	HandleErr(err)

	colPtrs := make([]interface{}, len(colNames))

	for i := 0; i < len(colNames); i++ {
		colType := colTypes[i].DatabaseTypeName()
		dbType := DBTypeToGoType(colType)
		switch dbType {
		case EnumDBType_Byte:
			var data byte
			colPtrs[i] = &data
		case EnumDBType_Int:
			var data int
			colPtrs[i] = &data
		case EnumDBType_Long:
			var data int64
			colPtrs[i] = &data
		case EnumDBType_String:
			var data string
			colPtrs[i] = &data
		case EnumDBType_Time:
			var data time.Time
			colPtrs[i] = &data
		case EnumDBType_Float:
			var data float32
			colPtrs[i] = &data
		}
	}

	var rowsCount int
	for rows.Next() {
		rowsCount++

		var resultMap = make(map[string]interface{})
		err = rows.Scan(colPtrs...)
		HandleErr(err)

		for i, col := range colPtrs {
			colType := colTypes[i].DatabaseTypeName()
			dbType := DBTypeToGoType(colType)
			switch dbType {
			case EnumDBType_Byte:
				resultMap[colNames[i]] = *col.(*byte)
			case EnumDBType_Int:
				resultMap[colNames[i]] = *col.(*int)
			case EnumDBType_Long:
				resultMap[colNames[i]] = *col.(*int64)
			case EnumDBType_String:
				resultMap[colNames[i]] = *col.(*string)
			case EnumDBType_Time:
				resultMap[colNames[i]] = *col.(*time.Time)
			case EnumDBType_Float:
				resultMap[colNames[i]] = *col.(*float32)
			}
		}

		dbModel.MapToStruct(resultMap)
	}
	if rowsCount == 0 {
		HandleErr(errors.New("No result [" + query + "]"))
	} else if rowsCount > 1 {
		HandleErr(errors.New("Too many results [" + query + "]"))
	}
}

func DBTypeToGoType(dbType string) EnumDBTypeValue {
	if dbType == "BIGINT" {
		return EnumDBType_Long
	} else if dbType == "TINYINT" {
		return EnumDBType_Byte
	} else if dbType == "SMALLINT" || dbType == "MEDIUMINT" || dbType == "INT" {
		return EnumDBType_Int
	} else if dbType == "VARCHAR" || dbType == "JSON" || dbType == "TEXT" {
		return EnumDBType_String
	} else if dbType == "DATE" || dbType == "DATETIME" || dbType == "TIMESTAMP" {
		return EnumDBType_Time
	} else if dbType == "FLOAT" || dbType == "DECIMAL" || dbType == "DOUBLE" {
		return EnumDBType_Float
	} else {
		HandleErr(errors.New("Can`t find proper DBType: " + dbType))
		return EnumDBType_Err
	}
}

func GetRedisSequenceKey(uid int64) string {
	return RedisUserSeqKey + strconv.FormatInt(uid, 64)
}

// 로그인 패킷에서 최초로 Sequence 값 Set
func InitUserSequence(redisClient *redis.Client, uid int64, initNum int, packetType EnumPacketValue) (seq int, resVal []byte, seqErrCode EnumSeqErrValue) {
	redisKey := GetRedisSequenceKey(uid)
	seq = (initNum % 100) * 1000000

	beforeSeqInfo := RedisSeqInfo{}
	curSeqInfo := RedisSeqInfo{seq, []byte{}}

	beforeSeqVal, err := redisClient.GetSet(RedisContext, redisKey, RedisSeqInfo{seq + 1, []byte{}}).Result()
	if err == redis.Nil {
		curSeqInfo.Seq = -1
	} else {
		HandleErr(err)
	}
	err = json.Unmarshal([]byte(beforeSeqVal), &beforeSeqInfo)
	HandleErr(err)

	if beforeSeqInfo.Seq == curSeqInfo.Seq+1 { // 진행 도중인 요청
		err = redisClient.Expire(RedisContext, redisKey, time.Second*10).Err()
		HandleErr(err)
		seqErrCode = EnumSeqErr_InProgress
		return
	} else if beforeSeqInfo.Seq == curSeqInfo.Seq+2 { // 완료된 요청
		err = redisClient.Set(RedisContext, redisKey, beforeSeqVal, time.Second*10).Err()
		HandleErr(err)
		resVal = beforeSeqInfo.ResVal
		seqErrCode = EnumSeqErr_DuplRequest
	}
	return
}

func CheckSequence(redisClient *redis.Client, uid int64, cSeq int) (resVal []byte, seqErrCode EnumSeqErrValue) {
	redisKey := GetRedisSequenceKey(uid)

	var seqInfo RedisSeqInfo
	beforeSeqVal, err := redisClient.GetSet(RedisContext, redisKey, RedisSeqInfo{cSeq + 1, []byte{}}).Result()
	HandleErr(err)

	err = json.Unmarshal([]byte(beforeSeqVal), &seqInfo)
	HandleErr(err)
	if seqInfo.Seq+1 == cSeq { // 정상
		return
	}

	if seqInfo.Seq == cSeq+2 { // 요청이 완료된 상태에서의 재요청
		// 기존 값 되돌림
		err = redisClient.Set(RedisContext, redisKey, beforeSeqVal, 0).Err()
		HandleErr(err)
		// 응답 값 셋팅
		seqErrCode = EnumSeqErr_DuplRequest
		resVal = seqInfo.ResVal
	} else if seqInfo.Seq == cSeq+1 { // 요청이 진행 중인 상태에서의 재요청
		seqErrCode = EnumSeqErr_InProgress
		return
	} else {
		HandleErr(fmt.Errorf("%s - [seqInfo]%+v - [cSeq]%d", "Check sequence error ", seqInfo, cSeq))
	}
	return
}

func UpdateSequence(redisClient *redis.Client, uid int64, cSeq int, resVal *[]byte) {
	redisKey := GetRedisSequenceKey(uid)

	seqInfo := RedisSeqInfo{cSeq + 2, *resVal}
	err := redisClient.Set(RedisContext, redisKey, seqInfo, RedisSeqTTL).Err()
	HandleErr(err)
}
