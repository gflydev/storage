package local

import (
	"fmt"
	"github.com/gflydev/core/log"
	"github.com/gflydev/core/utils"
	"github.com/gflydev/storage"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ========================================================================================
// 										Structure
// ========================================================================================

const (
	Type = storage.Type("local")
)

// New Create local storage.
func New() *Storage {
	return &Storage{
		BaseDir: utils.Getenv("STORAGE_DIR", "storage"),
	}
}

type Storage struct {
	BaseDir string
}

// Path Create file by content
func (s *Storage) Path(path string) string {
	fullPath := fmt.Sprintf("%s/%s", s.BaseDir, path)

	if strings.HasPrefix(path, s.BaseDir) {
		fullPath = path
	}

	return filepath.Clean(fullPath)
}

// ========================================================================================
// 									Implement IStorage
// ========================================================================================

// Put Create file by content
func (s *Storage) Put(path, contents string) bool {
	return s.PutData(path, []byte(contents))
}

// PutData Create file by content
func (s *Storage) PutData(path string, contents []byte) bool {
	// Create file
	file, err := os.Create(s.Path(path))
	if err != nil {
		log.Errorf("Unable create file %q. Here's why: %v\n", path, err)

		return false
	}

	// Defer to close file
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf("Unable to close file %q. Here's why: %v\n", path, err)
		}
	}(file)

	_, err = file.Write(contents)

	if err != nil {
		log.Errorf("Unable to write file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) PutFile(path string, fileSource *os.File) bool {
	// Create file
	file, err := os.Create(s.Path(path))
	if err != nil {
		log.Errorf("Unable create file %q. Here's why: %v\n", path, err)

		return false
	}

	// Defer to close file
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Errorf("Unable to close file %q. Here's why: %v\n", path, err)
		}
	}(file)

	_, err = io.Copy(file, fileSource)
	if err != nil {
		log.Errorf("Unable to write file %q. Here's why: %v\n", path, err)

		return false
	}

	err = file.Sync()
	if err != nil {
		log.Errorf("Unable to sync file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) Delete(path string) bool {
	err := os.Remove(s.Path(path))
	if err != nil {
		log.Errorf("Unable delete file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) Copy(from, to string) bool {
	// Open original file
	originalFile, err := os.Open(s.Path(from))
	if err != nil {
		log.Errorf("Unable open file %q. Here's why: %v\n", from, err)

		return false
	}
	// Defer to close file
	defer func(originalFile *os.File) {
		err = originalFile.Close()
		if err != nil {
			log.Errorf("Unable to close file %q. Here's why: %v\n", from, err)
		}
	}(originalFile)

	// Create new file
	newFile, err := os.Create(s.Path(to))
	if err != nil {
		log.Errorf("Unable create file %q. Here's why: %v\n", to, err)

		return false
	}
	// Defer to close file
	defer func(newFile *os.File) {
		err = newFile.Close()
		if err != nil {
			log.Errorf("Unable to close file %q. Here's why: %v\n", to, err)
		}
	}(newFile)

	// Copy the bytes to destination from source
	_, err = io.Copy(newFile, originalFile)
	if err != nil {
		log.Errorf("Unable copy file %q. Here's why: %v\n", to, err)

		return false
	}

	// Commit the file contents
	// Flushes memory to disk
	err = newFile.Sync()
	if err != nil {
		log.Errorf("Unable to sync file %q. Here's why: %v\n", to, err)

		return false
	}

	return true
}

func (s *Storage) Move(from, to string) bool {
	// Move file
	err := os.Rename(s.Path(from), s.Path(to))
	if err != nil {
		log.Errorf("Unable move file %q. Here's why: %v\n", to, err)

		return false
	}

	return true
}

func (s *Storage) Exists(path string) bool {
	// Stat returns file info. It will return an error if there is no file.
	_, err := os.Stat(s.Path(path))
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("File %v does not exist. Here's why: %v\n", path, err)
		}

		return false
	}

	return true
}

func (s *Storage) Get(path string) ([]byte, error) {
	// Stat returns file info. It will return an error if there is no file.
	fileData, err := os.ReadFile(s.Path(path))
	if err != nil {
		log.Errorf("Unable read file %q. Here's why: %v\n", path)

		return nil, err
	}

	return fileData, nil
}

func (s *Storage) Size(path string) int64 {
	// Stat returns file info. It will return an error if there is no file.
	fileInfo, err := os.Stat(s.Path(path))
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("File %v does not exist. Here's why: %v\n", path, err)
		}

		return 0
	}

	return fileInfo.Size()
}

func (s *Storage) LastModified(path string) time.Time {
	// Stat returns file info. It will return an error if there is no file.
	fileInfo, err := os.Stat(s.Path(path))
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("File %v does not exist. Here's why: %v\n", path, err)
		}

		return time.Time{}
	}

	return fileInfo.ModTime()
}

// Url Get public URL of an object via path
func (s *Storage) Url(path string) string {
	appURL := os.Getenv("APP_URL")

	return strings.TrimSuffix(appURL, "/") + "/" + strings.TrimPrefix(filepath.ToSlash(path), "/")
}

func (s *Storage) MakeDir(dir string) bool {
	err := os.MkdirAll(s.Path(dir), os.ModePerm)
	if err != nil {
		log.Errorf("Unable make dir %q. Here's why: %v\n", dir, err)

		return false
	}

	return true
}

func (s *Storage) DeleteDir(dir string) bool {
	err := os.Remove(s.Path(dir))
	if err != nil {
		log.Errorf("Unable delete dir %q. Here's why: %v\n", dir, err)

		return false
	}

	return true
}

func (s *Storage) Append(path, data string) bool {
	// Open original file
	file, err := os.OpenFile(s.Path(path), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		log.Errorf("Unable open file %q. Here's why: %v\n", path, err)

		return false
	}
	// Defer to close file
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Errorf("Unable to close file %q. Here's why: %v\n", path, err)
		}
	}(file)

	if _, err := file.WriteString(data); err != nil {
		log.Errorf("Unable write file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}
