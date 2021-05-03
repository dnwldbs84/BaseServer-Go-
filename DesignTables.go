package main

import (
	"encoding/json"
	"io/ioutil"
)

var (
	Test_Table01Data map[int]Test_Table01
	Test_Table02Data map[int]Test_Table02
	Test_Table03Data map[int]Test_Table03
)

type Test_Table01 struct {
	Key	int
	Test02	float64
	Test04	int
	Test06	int
	Test08	[]float64
	Test10	[]int
	Test11	int
}

type Test_Table02 struct {
	Key	int
	Test02	float64
	Test04	int
	Test06	int
}

type Test_Table03 struct {
	Key	int
	Test02	float64
	Test04	int
	Test06	int
}

func InitDesignTables() {
	var bytes []byte
	var err error

	Test_Table01Data = make(map[int]Test_Table01)
	var Test_Table01Val []Test_Table01
	Test_Table02Data = make(map[int]Test_Table02)
	var Test_Table02Val []Test_Table02
	Test_Table03Data = make(map[int]Test_Table03)
	var Test_Table03Val []Test_Table03

	bytes, err = ioutil.ReadFile("./Table/Test_Table01.json")
	HandleErr(err)
	err = json.Unmarshal(bytes, &Test_Table01Val)
	HandleErr(err)
	bytes, err = ioutil.ReadFile("./Table/Test_Table02.json")
	HandleErr(err)
	err = json.Unmarshal(bytes, &Test_Table02Val)
	HandleErr(err)
	bytes, err = ioutil.ReadFile("./Table/Test_Table03.json")
	HandleErr(err)
	err = json.Unmarshal(bytes, &Test_Table03Val)
	HandleErr(err)

	for _, data := range Test_Table01Val {
		Test_Table01Data[data.Key] = data
	}
	for _, data := range Test_Table02Val {
		Test_Table02Data[data.Key] = data
	}
	for _, data := range Test_Table03Val {
		Test_Table03Data[data.Key] = data
	}
}

