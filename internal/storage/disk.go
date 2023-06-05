package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"sync"
)

const defaultDirectory = "storage"

type DiskStorage struct {
	id        int
	directory string

	statMutex  *sync.RWMutex
	usedSpace  uint64
	totalSpace uint64
}

func NewDiskStorage(id int, usedSpace, totalSpace uint64) (*DiskStorage, error) {
	if err := os.MkdirAll(path.Join(defaultDirectory, strconv.FormatInt(int64(id), 10)), os.ModePerm); err != nil {
		return nil, err
	}

	return &DiskStorage{
		id:         id,
		directory:  defaultDirectory,
		statMutex:  &sync.RWMutex{},
		usedSpace:  usedSpace,
		totalSpace: totalSpace,
	}, nil
}

func (s *DiskStorage) ID() int {
	return s.id
}

func (s *DiskStorage) Store(ctx context.Context, name string, size uint64, obj io.Reader) error {
	filepath := s.buildPath(name)
	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		return errors.New("file already exists")
	}

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("unable to create file: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("closing file error: %s\n", err.Error())
		}
	}()

	if _, err := io.CopyN(f, obj, int64(size)); err != nil { // should use ctx controlled IO here to utilize passed context
		return fmt.Errorf("copy error: %w", err)
	}

	s.statMutex.Lock()
	defer s.statMutex.Unlock()
	s.usedSpace += size

	return nil
}

func (s *DiskStorage) Get(ctx context.Context, name string) (io.ReadCloser, uint64, error) {
	filepath := s.buildPath(name)
	stats, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return nil, 0, errors.New("file not found")
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to open file: %w", err)
	}

	return file, uint64(stats.Size()), nil
}

func (s *DiskStorage) Remove(ctx context.Context, name string) error {
	filepath := s.buildPath(name)
	stat, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return errors.New("file not found")
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("unable to remove file: %w", err)
	}

	s.statMutex.Lock()
	defer s.statMutex.Unlock()
	s.usedSpace -= uint64(stat.Size())

	return nil
}

func (s *DiskStorage) FreeSpace() float64 {
	s.statMutex.RLock()
	defer s.statMutex.RUnlock()

	if s.totalSpace == 0 {
		return 0.0
	}

	return 1.0 - float64(s.usedSpace)/float64(s.totalSpace)
}

func (s *DiskStorage) buildPath(filename string) string {
	return path.Join(s.directory, strconv.FormatInt(int64(s.id), 10), filename)
}
