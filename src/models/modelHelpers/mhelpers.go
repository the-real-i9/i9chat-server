package modelHelpers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func neo4jMapResToStruct(val any, dest any) {
	destType, destVal := reflect.TypeOf(dest).Elem(), reflect.ValueOf(dest).Elem()

	if reflect.TypeOf(val).Kind() != reflect.Map {
		panic("expected 'val' to be a map[string]any")
	}

	if destType.Kind() != reflect.Struct {
		panic("expected 'dest' to be a pointer to struct")
	}

	valMap, ok := val.(map[string]any)
	if !ok {
		panic("expected 'val' to be a map[string]any")
	}

	for i := range destType.NumField() {
		structField := destType.Field(i)

		mapKey, ok := structField.Tag.Lookup("db")
		if !ok {
			mapKey = strings.ToLower(structField.Name)
		}

		if mapKey == "" || mapKey == "-" {
			continue
		}

		if _, ok := valMap[mapKey]; !ok {
			continue
		}

		mapKeyT := reflect.TypeOf(valMap[mapKey])

		if !mapKeyT.AssignableTo(structField.Type) {
			panic(fmt.Sprintf("cannot map key %s of type %s to struct field %s of type %s", mapKey, mapKeyT.Kind(), structField.Name, structField.Type.Kind()))
		}

		destVal.Field(i).Set(reflect.ValueOf(valMap[mapKey]))
	}
}

func RKeyGet[T any](r []*neo4j.Record, key string) (res T) {
	var nilRes T

	if len(r) == 0 {
		return nilRes
	}

	switch reflect.TypeOf(nilRes).Kind() {
	case reflect.Struct:
		resAny, ok := r[0].Get(key)
		if !ok {
			return nilRes
		}

		neo4jMapResToStruct(resAny, &res)

		return res
	default:
		resAny, ok := r[0].Get(key)
		if !ok {
			return nilRes
		}

		return resAny.(T)
	}
}
