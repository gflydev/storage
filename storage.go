package storage

import (
	"github.com/gflydev/core/utils"
	"os"
	"time"
)

// ========================================================================================
// 										Structure
// ========================================================================================

type Type string

// ========================================================================================
// 										Manipulation
// ========================================================================================

type poolType map[string]IStorage

var defaultType = utils.Getenv("FILESYSTEM_TYPE", "local")
var pool = make(poolType)

// Register add storage instance via unique name
//
// Each storage type which implement IStorage interface should be registered by calling method
func Register(name string, storage IStorage) {
	pool[name] = storage
}

// Instance receive a storage instance. Get default storage for NONE `name` argument
func Instance(name ...string) IStorage {
	if len(name) == 0 {
		return pool[defaultType]
	}

	return pool[name[0]]
}

// ========================================================================================
// 										Interfaces
// ========================================================================================

// IStorage Storage interface
type IStorage interface {
	// -- Main actions

	// Put Create a file with content string
	Put(path, contents string, options ...interface{}) bool
	// PutFile Create a file from another file source
	PutFile(path string, fileSource *os.File, options ...interface{}) bool
	// Delete Remove a file
	Delete(path string) bool
	// Copy Clone file to another location
	Copy(from, to string) bool
	// Move Switch file location to new place
	Move(from, to string) bool

	//-- File manipulation

	// Exists Check existed file
	Exists(path string) bool
	// Get Receive a file content
	Get(path string) ([]byte, error)
	// Size Get file size
	Size(path string) int64
	// LastModified Obtains last modified of file
	LastModified(path string) time.Time
	// Url Absolute URL (Public URL)
	Url(path string) string

	//-- Directory

	// MakeDir Create new directory
	MakeDir(dir string) bool
	// DeleteDir Remove empty directory
	DeleteDir(dir string) bool

	//-- Data manipulation

	// Append Add string content to bottom file
	Append(path, data string) bool
}
