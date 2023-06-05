package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"karma/internal/models"
	"karma/internal/storage"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=service

type Service interface {
	Store(ctx context.Context, name string, size uint64, obj io.Reader) error
	Load(ctx context.Context, name string) (io.ReadCloser, uint64, error)
	Stats() []float64 // free space percentage [0.0, 1.0] per node
	AddNode() error
}

type Config struct {
	InitialNumNodes int
}

func DefaultConfig() Config {
	return Config{InitialNumNodes: 6}
}

type StorageService struct {
	config Config

	nodesMutex *sync.RWMutex
	nodes      []storage.Server

	objsMutex     *sync.Mutex
	storedObjects map[string]models.ObjectMeta
}

const (
	defaultMinUsed  = 5 * 1024 * 1024  // 5 MB
	defaultMinTotal = 20 * 1024 * 1024 // 20 MB
)

var (
	ErrNotFound      = errors.New("object not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrEmptyFile     = errors.New("empty file")
)

func NewStorageService(config Config) (*StorageService, error) {
	nodes := make([]storage.Server, config.InitialNumNodes)

	for i := 0; i < config.InitialNumNodes; i++ {
		node, err := storage.NewDiskStorage(i, uint64(defaultMinUsed+rand.Intn(10*1024*1024)+1),
			uint64(defaultMinTotal+rand.Intn(30*1024*1024)+1))
		if err != nil {
			return nil, fmt.Errorf("unable to create new disk storage: %w", err)
		}
		nodes[i] = node
	}

	return &StorageService{
		config:        config,
		nodesMutex:    &sync.RWMutex{},
		nodes:         nodes,
		objsMutex:     &sync.Mutex{},
		storedObjects: make(map[string]models.ObjectMeta, 0),
	}, nil
}

func (s *StorageService) Store(ctx context.Context, name string, size uint64, obj io.Reader) error {
	if size == 0 {
		return ErrEmptyFile
	}

	s.objsMutex.Lock()
	if _, exists := s.storedObjects[name]; exists {
		s.objsMutex.Unlock()
		return ErrAlreadyExists
	}
	s.storedObjects[name] = models.ObjectMeta{Name: name} // only name in meta = "uploading" status
	s.objsMutex.Unlock()

	stats := s.Stats()
	if len(stats) == 0 { //sanity check
		return errors.New("empty stats")
	}

	split := splitToParts(size, stats)
	nodes := make([]int, 0)
	for idx, chunkSize := range split {
		if chunkSize == 0 {
			continue
		}

		s.nodesMutex.RLock()
		node := s.nodes[idx]
		s.nodesMutex.RUnlock()

		if err := node.Store(ctx, name, chunkSize, obj); err != nil {
			if idx > 0 { //remove stored parts on previous nodes upon failure
				s.cleanupOnErr(ctx, name, idx)
			}
			return fmt.Errorf("unable to store on node %d: %w", idx, err)
		}
		nodes = append(nodes, idx)
	}

	s.objsMutex.Lock()
	s.storedObjects[name] = models.ObjectMeta{Name: name, TotalSize: size, Nodes: nodes}
	s.objsMutex.Unlock()

	return nil
}

func (s *StorageService) cleanupOnErr(ctx context.Context, name string, numFailedNodes int) {
	for i := 0; i < numFailedNodes; i++ {
		s.nodesMutex.RLock()
		node := s.nodes[i]
		s.nodesMutex.RUnlock()

		if err := node.Remove(ctx, name); err != nil {
			log.Println(fmt.Errorf("failed to remove %s on node %d: %w", name, i, err).Error())
		}
	}
}

func (s *StorageService) Load(ctx context.Context, name string) (io.ReadCloser, uint64, error) {
	s.objsMutex.Lock()
	meta, exists := s.storedObjects[name]
	s.objsMutex.Unlock()

	if !exists || meta.TotalSize == 0 {
		return nil, 0, ErrNotFound
	}

	pr, pw := io.Pipe()

	go func() {
		var err error
		for _, idx := range meta.Nodes {
			s.nodesMutex.RLock()
			node := s.nodes[idx]
			s.nodesMutex.RUnlock()

			if err = s.writeChunkToPipe(ctx, node, name, pw); err != nil {
				err = fmt.Errorf("write chunk to pipe error: %w", err)
				break
			}
		}

		//if err == nil - EOF close, otherwise Read will return err
		if closeErr := pw.CloseWithError(err); closeErr != nil {
			log.Println(fmt.Errorf("pipe writer close error: %w", closeErr).Error())
		}
	}()

	return pr, meta.TotalSize, nil
}

func (s *StorageService) writeChunkToPipe(ctx context.Context, node storage.Server, name string, pw *io.PipeWriter) error {
	chunkReadCloser, chunkSize, err := node.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("unable to get %s chunk from node %d: %w", name, node.ID(), err)
	}

	defer func() {
		if err := chunkReadCloser.Close(); err != nil {
			log.Printf("chunk readCloser close error: %s", err.Error())
		}
	}()

	written, err := io.Copy(pw, chunkReadCloser)
	if err != nil {
		return fmt.Errorf("copy error of %s chunk from node %d: %w", name, node.ID(), err)
	}

	//sanity check
	if uint64(written) != chunkSize {
		return errors.New("file size mismatch")
	}

	return nil
}

func (s *StorageService) Stats() []float64 {
	s.nodesMutex.RLock()
	defer s.nodesMutex.RUnlock()

	res := make([]float64, len(s.nodes))
	for i, n := range s.nodes {
		res[i] = n.FreeSpace()
	}

	return res
}

func (s *StorageService) AddNode() error {
	s.nodesMutex.Lock()
	defer s.nodesMutex.Unlock()

	const newNodeSpace = 30 * 1024 * 1024
	newNodeID := len(s.nodes)
	newNode, err := storage.NewDiskStorage(newNodeID, 0, newNodeSpace)
	if err != nil {
		err = fmt.Errorf("unable to create new disk storage: %w", err)
		log.Println(err.Error())
		return err
	}

	s.nodes = append(s.nodes, newNode)
	log.Printf("New storage %d added\n", newNodeID)
	return nil
}

// SplitToParts - split filesize into N (N= = nodes count) parts. Calculate the sizes of those parts according to FreeSpace.
func splitToParts(totalSize uint64, stats []float64) []uint64 {
	sum := 0.0
	idWithMostFreeSpace, idWithLeastFreeSpace := 0, 0
	for i, s := range stats {
		if s < stats[idWithLeastFreeSpace] {
			idWithLeastFreeSpace = i
		}
		if s > stats[idWithMostFreeSpace] {
			idWithMostFreeSpace = i
		}

		sum += s
		//		fmt.Printf("%.3f ", s)
	}

	//	fmt.Printf("\nleast: %d, most: %d\n", idWithLeastFreeSpace, idWithMostFreeSpace)

	meanFreePercentage := sum / float64(len(stats))
	if equal(meanFreePercentage, 0.0) { // sanity check
		log.Println("zero mean")
		return []uint64{}
	}

	//	fmt.Printf("\nmeanFree: %.3f", meanFreePercentage)

	meanSize := float64(totalSize) / float64(len(stats))
	//	fmt.Printf("\nmeanSize: %f\n", meanSize)

	split := make([]uint64, len(stats))
	var check uint64
	for i, nodeFreePercentage := range stats {
		split[i] = uint64(math.Round(meanSize * nodeFreePercentage / meanFreePercentage))
		check += split[i]
	}

	roundingError := int64(totalSize - check)
	if roundingError >= 0 {
		split[idWithMostFreeSpace] += uint64(roundingError)
	} else { // negative
		split[idWithLeastFreeSpace] -= uint64(math.Abs(float64(roundingError)))
	}

	var checkRoundingAdjusted uint64
	for _, s := range split {
		checkRoundingAdjusted += s
	}

	fmt.Printf("\ntotalSize = %d, check = %d, checkRoundingAdjusted = %d\n", totalSize, check, checkRoundingAdjusted)
	fmt.Println("Split:", split)
	return split
}

const float64EqualityThreshold = 1e-9

func equal(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
