package babyapiFileUploadParser

import (
	"bytes"
	"fmt"
	"github.com/go-chi/render"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type MultipartFormDecoder struct {
	request         *http.Request
	contextStack    []string
	baseContext     string
	fullContextKeys bool
	decodeInto      interface{}
}

func (d *MultipartFormDecoder) PushContext(newLayer string) {
	d.contextStack = append(d.contextStack, newLayer)
}

func (d *MultipartFormDecoder) PopContext() string {
	if len(d.contextStack) == 0 {
		return ""
	}
	oldContext := d.contextStack[len(d.contextStack)-1]
	d.contextStack = d.contextStack[:len(d.contextStack)-1]
	return oldContext
}

func (d *MultipartFormDecoder) GetContext() string {
	context := ""
	for _, c := range d.contextStack {
		if context == "" {
			context = c
		} else {
			if c[0] == '[' {
				context = fmt.Sprintf("%s%s", context, c)
			} else {
				context = fmt.Sprintf("%s.%s", context, c)
			}
		}
	}
	return context
}

func (d *MultipartFormDecoder) GetFormKey(field reflect.StructField) (string, []string) {
	formKey := field.Name
	var tags []string
	if tag := field.Tag.Get("form"); tag != "" {
		tags = strings.Split(tag, ",")
		formKey = tags[0]
	}
	return formKey, tags
}

func (d *MultipartFormDecoder) AddContext(formKey string) string {
	context := d.GetContext()
	if !d.fullContextKeys && context == d.baseContext {
		context = ""
	} else if !d.fullContextKeys {
		context = context[len(d.baseContext)+1:]
	}
	if context == "" {
		return formKey
	}
	if formKey == "" {
		return context
	}
	return fmt.Sprintf("%s.%s", context, formKey)
}

func (d *MultipartFormDecoder) isSettable(field reflect.StructField, value reflect.Value) bool {
	if !field.IsExported() {
		return false
	}
	if !value.CanSet() {
		return false
	}
	return true
}

func (d *MultipartFormDecoder) hasNewValue(fieldKind reflect.Kind, formKey string) bool {
	switch fieldKind {
	case reflect.Struct, reflect.Array, reflect.Slice, reflect.Map:
		return true
	default:
		return len(d.request.PostForm[formKey]) > 0
	}
}

func (d *MultipartFormDecoder) setIntFormValue(formValue string, fieldValue reflect.Value, bitSize int) {
	value, err := strconv.ParseInt(formValue, 10, bitSize)
	if err == nil {
		fieldValue.SetInt(value)
	} else {
		fmt.Println(err)
	}
}

func (d *MultipartFormDecoder) setUintFormValue(formValue string, fieldValue reflect.Value, bitSize int) {
	value, err := strconv.ParseUint(formValue, 10, bitSize)
	if err == nil {
		fieldValue.SetUint(value)
	} else {
		fmt.Println(err)
	}
}

func (d *MultipartFormDecoder) setFloatFormValue(formValue string, fieldValue reflect.Value, bitSize int) {
	value, err := strconv.ParseFloat(formValue, bitSize)
	if err == nil {
		fieldValue.SetFloat(value)
	} else {
		fmt.Println(err)
	}
}

func (d *MultipartFormDecoder) setBoolFormValue(formValue string, fieldValue reflect.Value) {
	value, err := strconv.ParseBool(formValue)
	if err == nil {
		fieldValue.SetBool(value)
	} else {
		fmt.Println(err)
	}
}

func (d *MultipartFormDecoder) setArrayFormValue(formKey string, tags []string, fieldValue reflect.Value) {
	for i := 0; i < fieldValue.Len(); i++ {
		if formKey == "" {
			d.PushContext(fmt.Sprintf("[%d]", i))
		} else {
			d.PushContext(fmt.Sprintf("%s[%d]", formKey, i))
		}

		d.CheckDecodeNode(d.request.FormValue(d.GetContext()), "", tags, fieldValue.Index(i))

		d.PopContext()
	}
}

func (d *MultipartFormDecoder) checkSliceElementAvailable(formKey string, index int) bool {
	fullKey := fmt.Sprintf("%s[%d]", d.AddContext(formKey), index)
	if d.hasNewValue(reflect.String, fullKey) {
		return true
	}
	for k := range d.request.PostForm {
		if len(k) >= len(fullKey) && k[:len(fullKey)] == fullKey {
			return true
		}
	}
	if d.request.MultipartForm != nil {
		for k := range d.request.MultipartForm.File {
			if len(k) >= len(fullKey) && k[:len(fullKey)] == fullKey {
				return true
			}
		}
	}
	return false
}

func (d *MultipartFormDecoder) setSliceFormValue(formKey string, tags []string, fieldValue reflect.Value) {
	newSlice := reflect.MakeSlice(reflect.SliceOf(fieldValue.Type().Elem()), 0, 0)
	for i := 0; d.checkSliceElementAvailable(formKey, i); i++ {
		if formKey == "" {
			d.PushContext(fmt.Sprintf("[%d]", i))
		} else {
			d.PushContext(fmt.Sprintf("%s[%d]", formKey, i))
		}

		newSliceElement := reflect.New(fieldValue.Type().Elem()).Elem()
		d.CheckDecodeNode(d.request.FormValue(d.GetContext()), "", tags, newSliceElement)
		newSlice = reflect.Append(newSlice, newSliceElement)

		d.PopContext()
	}
	fieldValue.Set(newSlice)
}

func extractKeys(baseKey string, k string) string {
	if len(k) >= len(baseKey)+2 && k[:len(baseKey)] == baseKey {
		partialKey := k[len(baseKey):]
		closingIndex := strings.IndexByte(partialKey, ']')
		if closingIndex > 0 {
			return partialKey[:closingIndex]
		}
	}
	return ""
}

func (d *MultipartFormDecoder) getMapKeys(formKey string) []string {
	baseKey := fmt.Sprintf("%s[", d.AddContext(formKey))
	var foundKeys []string
	for k := range d.request.PostForm {
		if foundKey := extractKeys(baseKey, k); foundKey != "" {
			foundKeys = append(foundKeys, foundKey)
		}
	}
	if d.request.MultipartForm != nil {
		for k := range d.request.MultipartForm.File {
			if foundKey := extractKeys(baseKey, k); foundKey != "" {
				foundKeys = append(foundKeys, foundKey)
			}
		}
	}
	return foundKeys
}

func (d *MultipartFormDecoder) setMapFormValue(formKey string, tags []string, fieldValue reflect.Value) {
	newMap := reflect.MakeMap(fieldValue.Type())
	for _, k := range d.getMapKeys(formKey) {
		if formKey == "" {
			d.PushContext(fmt.Sprintf("[%s]", k))
		} else {
			d.PushContext(fmt.Sprintf("%s[%s]", formKey, k))
		}

		newKey := reflect.New(fieldValue.Type().Key()).Elem()
		d.DecodeNode(k, "", tags, newKey)
		newValue := reflect.New(fieldValue.Type().Elem()).Elem()
		d.CheckDecodeNode(d.request.FormValue(d.GetContext()), "", tags, newValue)
		newMap.SetMapIndex(newKey, newValue)

		d.PopContext()
	}
	fieldValue.Set(newMap)
}

func (d *MultipartFormDecoder) setFileField(formKey string, tags []string, fieldValue reflect.Value) {
	file, header, err := d.request.FormFile(d.AddContext(formKey))
	if err == nil {
		fileBuf := bytes.Buffer{}
		_, err = io.Copy(&fileBuf, file)
		if err != nil {
			fmt.Println(d.AddContext(formKey), err)
			return
		}
		fileMeta := FileField{
			FileName:    header.Filename,
			FileSize:    header.Size,
			MIMEHeader:  header.Header,
			FormKey:     formKey,
			FormContext: d.GetContext(),
			rawFile:     fileBuf,
		}
		fieldValue.Set(reflect.ValueOf(fileMeta))
	} else {
		fmt.Println(d.AddContext(formKey), err)
	}
}

func (d *MultipartFormDecoder) CheckDecodeNode(formValue string, formKey string, tags []string, value reflect.Value) {
	if !d.hasNewValue(value.Kind(), d.AddContext(formKey)) {
		return
	}
	d.DecodeNode(formValue, formKey, tags, value)
}

func (d *MultipartFormDecoder) DecodeNode(formValue string, formKey string, tags []string, value reflect.Value) {
	switch elementKind := value.Kind(); elementKind {
	case reflect.Struct:
		if value.Type().AssignableTo(fileFieldType) {
			d.setFileField(formKey, tags, value)
		} else {
			if formKey != "" {
				d.PushContext(formKey)
				defer d.PopContext()
			}
			d.RecursiveStructDecoder(value)
		}
	case reflect.Int:
		d.setIntFormValue(formValue, value, 0)
	case reflect.Int8:
		d.setIntFormValue(formValue, value, 8)
	case reflect.Int16:
		d.setIntFormValue(formValue, value, 16)
	case reflect.Int32:
		d.setIntFormValue(formValue, value, 32)
	case reflect.Int64:
		d.setIntFormValue(formValue, value, 64)
	case reflect.Uint:
		d.setUintFormValue(formValue, value, 0)
	case reflect.Uint8:
		d.setUintFormValue(formValue, value, 8)
	case reflect.Uint16:
		d.setUintFormValue(formValue, value, 16)
	case reflect.Uint32:
		d.setUintFormValue(formValue, value, 32)
	case reflect.Uint64:
		d.setUintFormValue(formValue, value, 64)
	case reflect.Float32:
		d.setFloatFormValue(formValue, value, 32)
	case reflect.Float64:
		d.setFloatFormValue(formValue, value, 64)
	case reflect.Bool:
		d.setBoolFormValue(formValue, value)
	case reflect.String:
		value.SetString(formValue)
	case reflect.Array:
		d.setArrayFormValue(formKey, tags, value)
	case reflect.Slice:
		d.setSliceFormValue(formKey, tags, value)
	case reflect.Map:
		d.setMapFormValue(formKey, tags, value)
	default:
		slog.Error("Unsupported kind for element name", "name", value.Type().Name, "kind", elementKind)
	}

}

func (d *MultipartFormDecoder) RecursiveStructDecoder(v reflect.Value) {
	if !v.IsValid() {
		return
	}

	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}

	dest := v.Type()

	for i := 0; i < dest.NumField(); i++ {
		field := dest.Field(i)
		value := v.Field(i)
		formKey, tags := d.GetFormKey(field)
		if !d.isSettable(field, value) {
			continue
		}
		d.CheckDecodeNode(d.request.FormValue(d.AddContext(formKey)), formKey, tags, value)
	}
}

func DecodeMultipartForm(r *http.Request, v interface{}) {
	d := MultipartFormDecoder{
		request:         r,
		baseContext:     reflect.TypeOf(v).Elem().Name(),
		fullContextKeys: true,
		decodeInto:      v,
	}
	d.PushContext(reflect.TypeOf(d.decodeInto).Elem().Name())
	defer d.PopContext()
	d.RecursiveStructDecoder(reflect.ValueOf(d.decodeInto))
}

func Decoder(r *http.Request, v interface{}) error {
	var err error

	fmt.Println(r.Header.Get("Content-Type"))

	if r.Header.Get("Content-Type")[:19] == "multipart/form-data" {
		DecodeMultipartForm(r, v)
	} else {
		err = render.DefaultDecoder(r, v)
	}

	return err
}
