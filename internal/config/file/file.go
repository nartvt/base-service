package file

import (
	"base-service/internal/config/source"
	"base-service/internal/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var _ source.Source = (*file)(nil)

type file struct {
	path string
}

// NewSource new a file source.
func NewSource(path string) source.Source {
	return &file{path: path}
}

func (f *file) loadFile(path string) (*source.KeyValue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	return &source.KeyValue{
		Key:    info.Name(),
		Format: utils.Format(info.Name()),
		Value:  data,
	}, nil
}

func (f *file) loadDir(path string) (kvs []*source.KeyValue, err error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		// ignore hidden files
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}
		kv, err := f.loadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv)
	}
	return
}

func (f *file) Load() (kvs []*source.KeyValue, err error) {
	fi, err := os.Stat(f.path)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return f.loadDir(f.path)
	}
	kv, err := f.loadFile(f.path)
	if err != nil {
		return nil, err
	}
	return []*source.KeyValue{kv}, nil
}

func (f *file) Watch() (source.Watcher, error) {
	return newWatcher(f)
}
