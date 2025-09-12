package controllers

import (
	"github.com/gflydev/core"
	"github.com/gflydev/core/log"
	"github.com/gflydev/core/utils"
	"github.com/gflydev/storage"
	"github.com/gflydev/storage/ws3"
	"strings"
)

// ====================================================================
// ======================== Controller Creation =======================
// ====================================================================

// NewObjectsProxyPage As a constructor to create Proxy Page.
func NewObjectsProxyPage() *ObjectsProxyPage {
	return &ObjectsProxyPage{}
}

type ObjectsProxyPage struct {
	core.Page
}

// ====================================================================
// ========================= Request Handling =========================
// ====================================================================

// Handle is process incoming data from Validate to perform delete a content by id
//
//	Forward Proxy:
//		From http://localhost:7789/objects/one/two/hello.txt
//		To https://s3.ap-southeast-1.wasabisys.com/jivecode/one/two/hello.txt
func (m *ObjectsProxyPage) Handle(c *core.Ctx) error {
	// Remove /objects prefix and get the remaining path
	remainingPath := strings.TrimPrefix(c.Path(), "/objects/")

	// Create S3 storage with default
	fs := storage.Instance(ws3.Type)

	byteData, err := fs.Get(remainingPath)
	if err != nil {
		log.Error(err)
	}

	log.Infof("Read object: %s", utils.UnsafeStr(byteData))

	// Create S3 storage with default
	//fsLocal := storage.Instance(local.Type)
	//fileName := utils.MD5(remainingPath)
	//fsLocal.PutData(fileName, byteData)
	//defer fsLocal.Delete(fileName)

	/*file, err := os.Open("large_file.txt")
	if err != nil {
		return errors.New("Error opening file %d", fasthttp.StatusInternalServerError)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	fileInfo, err := file.Stat()
	if err != nil {
		return errors.New("Error getting file info %d", fasthttp.StatusInternalServerError)
	}

	c.Root().SetContentType("application/octet-stream")
	return c.Stream(file, int(fileInfo.Size()))*/

	// Serve the temp file
	return c.String(utils.UnsafeStr(byteData))
}
