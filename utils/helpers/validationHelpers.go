package helpers

import "reflect"

func ValidateStruct(s any) {
	t := reflect.TypeOf(s)

	for i := 0; i < t.NumField(); i++ {
		sfield := t.Field(i)

		sfieldTag := sfield.Tag

		sfieldTag.Get("i9validate")
	}
}
