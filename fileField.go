package babyapiFileUploadParser

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
)

type FileField struct {
	FileName    string
	FileSize    int64
	FormKey     string
	FormContext string
	rawFile     bytes.Buffer
}

var fileFieldType = reflect.TypeOf((*FileField)(nil)).Elem()

func (ff *FileField) GetFileName(fileStorePath string, id string) string {
	directory := fmt.Sprintf("%s/%s", fileStorePath, id)
	filename := fmt.Sprintf("%s/%s", directory, ff.FormContext)
	if ff.FormKey != "" {
		filename += "." + ff.FormKey
	}
	return filename
}

func (ff *FileField) WriteFile(fileStorePath string, id string) error {
	if !ff.HasContent() {
		return fmt.Errorf("attempt writing empty file")
	}
	directory := fmt.Sprintf("%s/%s", fileStorePath, id)
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return err
	}
	filename := ff.GetFileName(fileStorePath, id)
	err = os.WriteFile(filename, ff.rawFile.Bytes(), 0644)
	if err != nil {
		return err
	}
	ff.rawFile = bytes.Buffer{}
	return err
}

func (ff *FileField) GetFile(fileStorePath string, id string) (*os.File, error) {
	return os.Open(ff.GetFileName(fileStorePath, id))
}

func (ff *FileField) HasContent() bool {
	return ff.rawFile.Available() > 0
}
