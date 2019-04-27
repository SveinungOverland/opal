package router

import (
	"io/ioutil"
	"strings"
)

// FileHandler handles the reading of the file
type FileHandler struct {
	relativePath string
	filePath     string
	MimeType	 string
}

func newFileHandler(relativePath string, filePath string) *FileHandler {
	fs := &FileHandler{
		relativePath: relativePath,
		filePath:     filePath,
	}

	// Check if filePath is actually a directory, if so, add index.html
	splitFilePath := strings.Split(filePath, ".")
	if len(splitFilePath) == 1 {
		fs.filePath = strings.TrimRight(fs.filePath, "/") + "/index.html"
		fs.MimeType = Mimes[".html"]
	} else {
		fileExtension := splitFilePath[len(splitFilePath) - 1]
		fs.MimeType = Mimes["." + fileExtension]
	}

	return fs
}

// FullPath returns the full path of the file.
func (fh *FileHandler) FullPath() string {
	return strings.TrimRight(fh.relativePath, "/") + "/" + strings.TrimLeft(fh.filePath, "/")
}

// ReadFile reads the file of the provided path. Returns an error if it does not exist.
func (fh *FileHandler) ReadFile() ([]byte, error) {
	path := fh.FullPath()
	return ioutil.ReadFile(path)
}
