package babyapiFileUploadParser

import (
	"fmt"
	"github.com/calvinmclean/babyapi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"net/http"
	"os"
	"time"
)

type FileStore[T babyapi.Resource] struct {
	api           *babyapi.API[T]
	FileStorePath string
}

func NewFileStore[T babyapi.Resource](api *babyapi.API[T], BaseFileStorePath string) FileStore[T] {
	fs := FileStore[T]{api: api, FileStorePath: fmt.Sprintf("%s/%s", BaseFileStorePath, api.Name())}
	return fs
}

func (fs FileStore[T]) ServeFilesRoute(basePattern string) chi.Route {
	return chi.Route{
		Pattern: fmt.Sprintf("%s/{fileId}", basePattern),
		Handlers: map[string]http.Handler{
			http.MethodGet: babyapi.Handler(fs.ServeFile),
		},
	}
}

func (fs FileStore[T]) ServeFile(w http.ResponseWriter, r *http.Request) render.Renderer {
	fmt.Println("Attempting to serve static file")
	resource, err := fs.api.GetResourceFromContext(r.Context())
	if err != nil {
		return babyapi.ErrRender(err)
	}
	fileName := chi.URLParam(r, "fileId")
	ff, err := FindByFileName(fileName, resource)
	if err != nil {
		return babyapi.ErrRender(err)
	}
	file, err := ff.GetFile(fs.FileStorePath, resource.GetID())
	if err != nil {
		return babyapi.ErrRender(err)
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+ff.FileName)
	http.ServeContent(w, r, ff.FileName, time.Now(), file)
	return nil
}

func (fs FileStore[T]) DeleteHook(r *http.Request) *babyapi.ErrResponse {
	id := fs.api.GetIDParam(r)
	err := fs.DeleteResourceFiles(id)
	if err != nil {
		fmt.Println(err)
		return babyapi.ErrRender(err)
	}
	return nil
}

func (fs FileStore[T]) CreateUpdateHook(_ *http.Request, resource T) *babyapi.ErrResponse {
	WriteAllFileFields(fs.FileStorePath, resource.GetID(), resource)
	return nil
}

func (fs FileStore[T]) AutoAddHooks() {
	fs.api.SetAfterDelete(fs.DeleteHook)
	fs.api.SetOnCreateOrUpdate(fs.CreateUpdateHook)
}

func (fs FileStore[T]) DeleteResourceFiles(id string) error {
	return os.RemoveAll(fmt.Sprintf("%s/%s", fs.FileStorePath, id))
}
