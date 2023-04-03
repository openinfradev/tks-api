package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
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
				if converter, ok := converterMap[compositeKey{srcType: srcField.Type(), dstType: dstField.Type()}]; ok {
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
		{srcType: reflect.TypeOf((*uuid.UUID)(nil)), dstType: reflect.TypeOf("")}: func(i interface{}) (interface{}, error) {
			return i.(uuid.UUID).String(), nil
		},
		{srcType: reflect.TypeOf(""), dstType: reflect.TypeOf((*uuid.UUID)(nil))}: func(i interface{}) (interface{}, error) {
			val, _ := uuid.Parse(i.(string))
			return val, nil
		},
		{srcType: reflect.TypeOf((*domain.OrganizationStatus)(nil)), dstType: reflect.TypeOf("")}: func(i interface{}) (interface{}, error) {
			return string(i.(domain.OrganizationStatus)), nil
		},
		{srcType: reflect.TypeOf(""), dstType: reflect.TypeOf((*domain.OrganizationStatus)(nil))}: func(i interface{}) (interface{}, error) {
			return i.(domain.OrganizationStatus).String(), nil
		},
	})
}
