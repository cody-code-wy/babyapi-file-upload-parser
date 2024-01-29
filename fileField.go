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

func (ff *FileField) GetFileName() string {
	filename := ff.FormContext
	if ff.FormKey != "" {
		filename += "." + ff.FormKey
	}
	return filename
}

func (ff *FileField) GetFileDir(fileStorePath string, id string) string {
	return fmt.Sprintf("%s/%s", fileStorePath, id)
}

func (ff *FileField) GetFilePath(fileStorePath string, id string) string {
	return fmt.Sprintf("%s/%s", ff.GetFileDir(fileStorePath, id), ff.GetFileName())
}

func (ff *FileField) WriteFile(fileStorePath string, id string) error {
	if !ff.HasContent() {
		return fmt.Errorf("attempt writing empty file")
	}
	directory := ff.GetFileDir(fileStorePath, id)
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return err
	}
	filename := ff.GetFilePath(fileStorePath, id)
	err = os.WriteFile(filename, ff.rawFile.Bytes(), 0644)
	if err != nil {
		return err
	}
	ff.rawFile = bytes.Buffer{}
	return err
}

func (ff *FileField) GetFile(fileStorePath string, id string) (*os.File, error) {
	return os.Open(ff.GetFilePath(fileStorePath, id))
}

func (ff *FileField) DeleteFile(fileStorePath string, id string) error {
	return os.Remove(ff.GetFilePath(fileStorePath, id))
}

func (ff *FileField) HasContent() bool {
	return ff.rawFile.Available() > 0
}
