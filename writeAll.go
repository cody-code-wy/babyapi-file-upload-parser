package babyapiFileUploadParser

import (
	"fmt"
	"reflect"
)

type fileFieldSearcherContent struct {
	writePath string
	id        string
	handler   func(fileField FileField, context *fileFieldSearcherContent)
	v         interface{}
}

func (c *fileFieldSearcherContent) checkArray(v reflect.Value) {
	for i := 0; i < v.Len(); i++ {
		c.recursiveChecker(v.Index(i))
	}
}

func (c *fileFieldSearcherContent) checkStruct(v reflect.Value) {
	switch fileField := v.Interface().(type) {
	case FileField:
		c.handler(fileField, c)
		return
	default:
		dest := v.Type()
		for i := 0; i < dest.NumField(); i++ {
			c.recursiveChecker(v.Field(i))
		}
	}
}

func (c *fileFieldSearcherContent) checkMap(v reflect.Value) {
	for _, k := range v.MapKeys() {
		c.recursiveChecker(k)
		c.recursiveChecker(v.MapIndex(k))
	}
}

func (c *fileFieldSearcherContent) recursiveChecker(v reflect.Value) {
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
	context := fileFieldSearcherContent{writePath: writePath, id: id, v: v, handler: func(fileField FileField, context *fileFieldSearcherContent) {
		if fileField.HasContent() {
			err := fileField.WriteFile(context.writePath, context.id)
			if err != nil {
				fmt.Println(err)
			}
		}
	}}
	context.recursiveChecker(reflect.ValueOf(v))
}
