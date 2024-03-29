package main

import (
	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/storage"
	"github.com/cody-code-wy/babyapi-file-upload-parser"
	"github.com/go-chi/render"
	"github.com/madflojo/hord/drivers/hashmap"
)

type TestStruct struct {
	Test string
}

type Types struct {
	// Numbers
	Int    int
	Int8   int8
	Int16  int16
	Int32  int32
	Int64  int64
	Uint   uint
	Uint8  uint8
	Uint16 uint16
	Uint32 uint32
	Uint64 uint64
	Rune   rune
	Byte   byte
	// Uintptr uintptr // why should I deserialize pointers??
	// Floats
	Float32 float32
	Float64 float64
	// Complex not supported by JSON
	// Complex64  complex64
	// Complex128 complex128
	// Others
	Boolean       bool
	String        string
	Struct        TestStruct
	Array         [2]int8
	Array2D       [2][2]int8
	StructArray   [3]TestStruct
	Slice         []string
	Slice2D       [][]float32
	StructSlice   []TestStruct
	SliceArray    [][3]float64
	ArraySlice    [3][]float64
	Image         babyapi_file_upload_parser.FileField
	Images        []babyapi_file_upload_parser.FileField
	Images2D      [][]babyapi_file_upload_parser.FileField
	ImagesArray   [3]babyapi_file_upload_parser.FileField
	ImagesArray2D [2][2]babyapi_file_upload_parser.FileField
	privateInt    int
	Maps          Maps
}

type Maps struct {
	SliceStrStr []map[string]string
	StrStr      map[string]string
	StrStruct   map[string]TestStruct
	IntStr      map[int]string
	StrImage    map[string]babyapi_file_upload_parser.FileField
	// StructStr map[TestStruct]string // struct keys not supported by JSON
}

type Project struct {
	babyapi.DefaultResource

	Name        string `form:"projectName" json:"projectName"`
	Description string
	Test        string
	Image       babyapi_file_upload_parser.FileField
	Image2      babyapi_file_upload_parser.FileField `form:"OtherImage" json:"OtherImage"`
	Types       Types
}

func main() {
	render.Decode = babyapi_file_upload_parser.Decoder

	ProjectApi := babyapi.NewAPI[*Project]("Projects", "/Projects", func() *Project { return &Project{} })
	projectFileStore := babyapi_file_upload_parser.NewFileStore[*Project](ProjectApi, "./Uploads")
	projectFileStore.AutoAddHooks()
	ProjectApi.AddCustomIDRoute(projectFileStore.ServeFilesRoute("/file"))

	db, err := storage.NewFileDB(hashmap.Config{
		Filename: "projects.db.json",
	})
	if err != nil {
		return
	}

	ProjectApi.Storage = storage.NewClient[*Project](db, "User")

	ProjectApi.RunCLI()
}
