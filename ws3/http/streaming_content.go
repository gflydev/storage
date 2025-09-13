package http

import (
	"github.com/gflydev/core"
	"github.com/gflydev/core/log"
	"github.com/gflydev/storage"
	"github.com/gflydev/storage/ws3"
)

// StreamObject Stream object from Wasabi S3
func StreamObject(c *core.Ctx, objectPath string) error {
	// Create S3 storage with default
	fs := storage.Instance(ws3.Type)

	// Get the object as a stream instead of loading into memory
	objectStream, err := fs.GetStream(objectPath)
	if err != nil {
		log.Errorf("Unable to get object stream for %s: %v", objectPath, err)
		return c.Status(core.StatusInternalServerError).String("Internal server error")
	}

	return c.Stream(objectStream)
}
