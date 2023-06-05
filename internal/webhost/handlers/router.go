package handlers

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"runtime/debug"
)

const apiPrefix = "/api/v1"

func CreateRouter(storageServiceCtrl *StorageService) *mux.Router {
	router := mux.NewRouter()
	router.Use(recoverPanic)

	apiRouter := router.PathPrefix(apiPrefix).Subrouter()
	apiRouter.Methods(http.MethodPost).Path(filePath).HandlerFunc(storageServiceCtrl.UploadFile)
	apiRouter.Methods(http.MethodGet).Path(filePath).HandlerFunc(storageServiceCtrl.GetFile)

	return router
}

func recoverPanic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, request *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if err, ok := rec.(error); ok {
					log.Printf("%s: %s", err, string(debug.Stack()))
				}
			}
		}()
		h.ServeHTTP(resp, request)
	})
}
