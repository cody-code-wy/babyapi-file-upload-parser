package babyapiFileUploadParser

import (
	"fmt"
	"reflect"
)

type writeAllFileFieldsContext struct {
	writePath string
	id        string
	v         interface{}
}

func (c *writeAllFileFieldsContext) checkArray(v reflect.Value) {
	for i := 0; i < v.Len(); i++ {
		c.recursiveChecker(v.Index(i))
	}
}

func (c *writeAllFileFieldsContext) checkStruct(v reflect.Value) {
	switch fileField := v.Interface().(type) {
	case FileField:
		if fileField.HasContent() {
			err := fileField.WriteFile(c.writePath, c.id)
			if err != nil {
				fmt.Println(err)
			}
		}
		return
	default:
		dest := v.Type()
		for i := 0; i < dest.NumField(); i++ {
			c.recursiveChecker(v.Field(i))
		}
	}
}

func (c *writeAllFileFieldsContext) checkMap(v reflect.Value) {
	for _, k := range v.MapKeys() {
		c.recursiveChecker(k)
		c.recursiveChecker(v.MapIndex(k))
	}
}

func (c *writeAllFileFieldsContext) recursiveChecker(v reflect.Value) {
	if !v.IsValid() {
		return
	}
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		c.checkArray(v)
	case reflect.Map:
		c.checkMap(v)
	case reflect.Struct:
		c.checkStruct(v)
	default: // Nothing to do
		//fmt.Println("Nothing to do for type", v.Kind())
	}
}

func WriteAllFileFields(writePath string, id string, v interface{}) {
	context := writeAllFileFieldsContext{writePath: writePath, id: id, v: v}
	context.recursiveChecker(reflect.ValueOf(v))
}
