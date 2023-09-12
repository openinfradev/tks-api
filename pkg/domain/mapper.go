package domain

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/log"
	"gorm.io/datatypes"
)

type ConverterMap map[compositeKey]func(interface{}) (interface{}, error)

type compositeKey struct {
	srcType reflect.Type
	dstType reflect.Type
}

func recursiveMap(src interface{}, dst interface{}, converterMap ConverterMap) error {
	// not support type
	if _, ok := src.(datatypes.JSON); ok {
		return nil
	}

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
				if functionExists(srcField.Interface(), "String") &&
					functionExists(srcField.Interface(), "FromString") &&
					dstField.Type().Kind() == reflect.String {
					mthd := reflect.ValueOf(srcField.Interface()).MethodByName("String").Call([]reflect.Value{})
					if len(mthd) > 0 {
						dstField.Set(mthd[0])
						continue
					}
				}
				if functionExists(dstField.Interface(), "String") &&
					functionExists(dstField.Interface(), "FromString") &&
					srcField.Type().Kind() == reflect.String {
					mthd := reflect.ValueOf(dstField.Interface()).MethodByName("FromString").Call([]reflect.Value{srcField})
					if len(mthd) > 0 {
						dstField.Set(mthd[0])
						continue
					}
				}

				converterKey := compositeKey{srcType: srcField.Type(), dstType: dstField.Type()}
				if converter, ok := converterMap[converterKey]; ok {
					if converted, err := converter(srcField.Interface()); err != nil {
						return err
					} else {
						dstField.Set(reflect.ValueOf(converted))
					}
				} else {
					log.Debugf("no converter found for %s -> %s", srcField.Type(), dstField.Type())
					continue
				}
			}

			/*
				 else if srcField.Type().Kind() == reflect.Ptr && dstField.Type().Kind() == reflect.Ptr {
					log.Info("AAA ", dstField.Type())
					ptr := reflect.New(dstField.Elem().Type())
					if err := recursiveMap(srcField.Elem().Interface(), ptr.Elem().Interface(), converterMap); err != nil {
						return err
					}
				}
			*/

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
		{srcType: reflect.TypeOf((*Role)(nil)).Elem(), dstType: reflect.TypeOf("")}: func(i interface{}) (interface{}, error) {
			return i.(Role).Name, nil
		},
		{srcType: reflect.TypeOf(""), dstType: reflect.TypeOf((*Role)(nil)).Elem()}: func(i interface{}) (interface{}, error) {
			return Role{Name: i.(string)}, nil
		},
	})
}

func functionExists(obj interface{}, funcName string) bool {
	mthd := reflect.ValueOf(obj).MethodByName(funcName)
	return mthd.IsValid()
}
