go-chi decoder for http-multipart with some helper structure and functions designed to easily add file uploading support to [BabyAPI](https://github.com/calvinmclean/babyapi)

### Usage

The example file is the best place to check for usage, and all the currently supported tested types and structures, but the basics are simple.

1. Overritde the default decoder `render.Decode = babyapiFileUploadParser.Decoder` somewhere before starting your APIs
2. Add a `babyapiFileUploadParser.FileField` field to your API struct (or a slice of them maybe a map, whatever you need)
3. After create or update, write your files somewhere.
   * Call `WriteFile(uploadPath string, id string)` on each FileField item
   * Call `babyapiFileUploadParser.WriteAllFileFields(uploadPath string, id string, api interface{})` on your api struct to write every FileField inside
4. On delete ensure you clean up files

there is also a FileStore helper that helps implementing the needed hooks on the api to save and delete files along with providing a go-chi route that can be used for serving the files back. Here is an example of it's use

```
myFileStore := babyapiFileUploadParser.NewFileStore[*myResource](myAPI, "./UploadPath")
// using AutoAddHooks will automate writing files on create and update, and deleting
myFileStore.AutoAddHooks()
// Or you can add only the hooks you want
myAPI.SetAfterDelete(myFileStore.DeleteHook)
myAPI.SetOnCreateOrUpdate(myFileStore.CreateUpdateHook)
// Finally add the custom ID route to the myResource
myApi.AddCustomIDRoute(myFileStore.ServeFilesRoute("/files"))
```
