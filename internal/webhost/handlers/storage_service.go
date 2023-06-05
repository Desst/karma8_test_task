package handlers

import (
	"errors"
	"fmt"
	"io"
	"karma/internal/service"
	"log"
	"net/http"
)

type StorageService struct {
	service service.Service
}

func NewStorageService(service service.Service) *StorageService {
	return &StorageService{service: service}
}

const (
	filePath         = "/file"
	filenameUrlParam = "filename"
)

func (s *StorageService) UploadFile(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get(filenameUrlParam)
	if filename == "" || r.ContentLength <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := s.service.Store(r.Context(), filename, uint64(r.ContentLength), r.Body)
	switch {
	case errors.Is(err, service.ErrEmptyFile):
		w.WriteHeader(http.StatusBadRequest)
		return
	case errors.Is(err, service.ErrAlreadyExists):
		w.WriteHeader(http.StatusConflict)
		return
	case err != nil:
		log.Println(fmt.Errorf("unable to store file %s: %w", filename, err).Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *StorageService) GetFile(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get(filenameUrlParam)
	if filename == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pr, size, err := s.service.Load(r.Context(), filename)
	switch {
	case errors.Is(err, service.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
		return
	case err != nil:
		log.Println(fmt.Errorf("unable to load file %s: %w", filename, err).Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	copied, err := io.Copy(w, pr)
	if err != nil {
		log.Println(fmt.Errorf("copy error: %w", err).Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if uint64(copied) != size { //sanity check
		log.Println("size mismatch")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
