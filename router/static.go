package router

import (
	"io/ioutil"
	"strings"
)

// FileHandler handles the reading of the file
type FileHandler struct {
	relativePath string
	filePath     string
}

func newFileHandler(relativePath string, filePath string) *FileHandler {
	return &FileHandler{
		relativePath: relativePath,
		filePath:     filePath,
	}
}

// FullPath returns the full path of the file.
func (fh *FileHandler) FullPath() string {
	return strings.TrimRight(fh.relativePath, "/") + "/" + fh.filePath
}

// ReadFile reads the file of the provided path. Returns an error if it does not exist.
func (fh *FileHandler) ReadFile() ([]byte, error) {
	path := fh.FullPath()
	return ioutil.ReadFile(path)
}
