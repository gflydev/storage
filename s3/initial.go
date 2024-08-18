package s3

import "github.com/gflydev/storage"

// ========================================================================================
// 										Initial
// ========================================================================================

// Auto initial S3 storage and register to storage manager
func init() {
	storage.Register(string(Type), New())
}
