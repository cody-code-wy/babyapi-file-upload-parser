go-chi decoder for http-multipart with some helper structure and functions designed to easily add file uploading support to [BabyAPI](https://github.com/calvinmclean/babyapi)

### Usage

The example file is the best place to check for usage, and all the currently supported tested types and structures, but the basics are simple.

1. Overritde the default decoder `render.Decode = babyapiFileUploadParser.Decoder` somewhere before starting your APIs
2. Add a `babyapiFileUploadParser.FileField` field to your API struct (or a slice of them maybe a map, whatever you need)
3. After create or update, write your files somewhere.
   * Call `WriteFile(uploadPath string, id string)` on each FileField item
   * Call `babyapiFileUploadParser.WriteAllFileFields(uploadPath string, id string, api interface{})` on your api struct to write every FileField inside
4. On delete ensure you clean up files
