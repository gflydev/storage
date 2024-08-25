package local

import "github.com/gflydev/storage"

// ========================================================================================
// 										Initial
// ========================================================================================

// Auto initial local storage and register to storage manager
func init() {
	storage.Register(Type, New())
}
