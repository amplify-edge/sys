package coredb

import (
	"reflect"
	"strings"
)

func getStructTags(model interface{}) map[string]string {
	fieldMap := map[string]string{}
	fieldsLength := reflect.ValueOf(model).NumField()
	for i := 0; i < fieldsLength; i++ {
		field := reflect.TypeOf(model).Field(i)
		fieldStructType := field.Type.String()
		genjiTag := field.Tag.Get("genji")
		switch fieldStructType {
		case "string":
			fieldMap[genjiTag] = "TEXT"
		case "int", "int64", "uint", "uint64", "int32", "uint32":
			fieldMap[genjiTag] = "INTEGER"
		case "double", "float", "float32", "float64":
			fieldMap[genjiTag] = "DOUBLE"
		case "bool":
			fieldMap[genjiTag] = "BOOL"
		case "map[string]interface {}":
			fieldMap[genjiTag] = "TEXT"
		case "[]byte", "[]uint8":
			fieldMap[genjiTag] = "BLOB"
		case "[]string", "[]int", "[]int32", "[]int64", "[]uint", "[]uint32", "[]uint64", "[]float64", "[]float32", "[]float", "[]double":
			fieldMap[genjiTag] = "ARRAY"
		default:
			fieldMap[genjiTag] = "DOCUMENT"
		}
		coredbTag := field.Tag.Get("coredb")
		if coredbTag == "primary" {
			fieldMap[genjiTag] += " PRIMARY KEY"
		}
	}
	return fieldMap
}

func GetStructTags(model interface{}) map[string]string {
	return getStructTags(model)
}

func GetStructColumns(someStruct interface{}) string {
	structTags := GetStructTags(someStruct)
	var structCols []string
	for k, _ := range structTags {
		structCols = append(structCols, k)
	}
	return strings.Join(structCols, ", ")
}

