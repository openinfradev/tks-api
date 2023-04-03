package domain

import (
	"fmt"
	"github.com/google/uuid"
	"reflect"
)

type ConverterMap map[compositeKey]func(interface{}) (interface{}, error)

type compositeKey struct {
	srcType reflect.Type
	dstType reflect.Type
}

func recursiveMap(src interface{}, dst interface{}, converterMap ConverterMap) error {
	srcVal := reflect.ValueOf(src)
	srcType := srcVal.Type()

	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return fmt.Errorf("dst must be a non-nil pointer")
	}
	dstElem := dstVal.Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		fieldName := srcType.Field(i).Name
		srcField := srcVal.Field(i)
		dstField := dstElem.FieldByName(fieldName)

		if dstField.IsValid() && dstField.CanSet() {
			if dstField.Type() == srcField.Type() {
				dstField.Set(srcField)
				continue
			} else if srcField.Type().Kind() == reflect.Struct && dstField.Type().Kind() == reflect.Struct {
				if err := recursiveMap(srcField.Interface(), dstField.Addr().Interface(), converterMap); err != nil {
					return err
				}
			} else {
				converterKey := compositeKey{srcType: srcField.Type(), dstType: dstField.Type()}
				if converter, ok := converterMap[converterKey]; ok {
					if converted, err := converter(srcField.Interface()); err != nil {
						return err
					} else {
						dstField.Set(reflect.ValueOf(converted))
					}
				} else {
					return fmt.Errorf("no converter found for %s -> %s", srcField.Type(), dstField.Type())
				}
			}
		}
	}

	return nil
}
func Map(src interface{}, dst interface{}) error {
	return recursiveMap(src, dst, ConverterMap{
		{srcType: reflect.TypeOf((*uuid.UUID)(nil)).Elem(), dstType: reflect.TypeOf("")}: func(i interface{}) (interface{}, error) {
			return i.(uuid.UUID).String(), nil
		},
		{srcType: reflect.TypeOf(""), dstType: reflect.TypeOf((*uuid.UUID)(nil)).Elem()}: func(i interface{}) (interface{}, error) {
			val, _ := uuid.Parse(i.(string))
			return val, nil
		},
		{srcType: reflect.TypeOf((*OrganizationStatus)(nil)).Elem(), dstType: reflect.TypeOf("")}: func(i interface{}) (interface{}, error) {
			return string(i.(OrganizationStatus)), nil
		},
		{srcType: reflect.TypeOf(""), dstType: reflect.TypeOf((*OrganizationStatus)(nil)).Elem()}: func(i interface{}) (interface{}, error) {
			return i.(OrganizationStatus).String(), nil
		},
		{srcType: reflect.TypeOf((*Role)(nil)).Elem(), dstType: reflect.TypeOf("")}: func(i interface{}) (interface{}, error) {
			return i.(Role).Name, nil
		},
		{srcType: reflect.TypeOf(""), dstType: reflect.TypeOf((*Role)(nil)).Elem()}: func(i interface{}) (interface{}, error) {
			return Role{Name: i.(string)}, nil
		},
	})
}
