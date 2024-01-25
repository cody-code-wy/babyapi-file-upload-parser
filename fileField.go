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
	valid       bool
}

var fileFieldType = reflect.TypeOf((*FileField)(nil)).Elem()

func (ff *FileField) getFileName(fileStorePath string, id string) string {
	directory := fmt.Sprintf("%s/%s", fileStorePath, id)
	filename := fmt.Sprintf("%s/%s", directory, ff.FormContext)
	if ff.FormKey != "" {
		filename += "." + ff.FormKey
	}
	filename += ">" + ff.FileName
	return filename
}

func (ff *FileField) writeFile(fileStorePath string, id string) error {
	if !ff.valid {
		return fmt.Errorf("attempt writing empty file")
	}
	directory := fmt.Sprintf("%s/%s", fileStorePath, id)
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return err
	}
	filename := ff.getFileName(fileStorePath, id)
	return os.WriteFile(filename, ff.rawFile.Bytes(), 0644)
}

func (ff *FileField) getFile(fileStorePath string, id string) (*os.File, error) {
	return os.Open(ff.getFileName(fileStorePath, id))
}
