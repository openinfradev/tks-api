package domain

import (
	"fmt"
	"reflect"
)

func Map(src interface{}, dst interface{}) error {

	srcVal := reflect.ValueOf(src)
	srcType := srcVal.Type()

	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return fmt.Errorf("dst must be a non-nil pointer")
	}
	dstElem := dstVal.Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		fieldName := srcType.Field(i).Name
		dstField := dstElem.FieldByName(fieldName)
		if dstField.IsValid() && dstField.CanSet() {
			dstField.Set(srcVal.Field(i))
		}
	}

	return nil

}
