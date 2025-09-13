package controllers

import (
	"context"
	"fmt"
	"github.com/gflydev/core"
	"github.com/gflydev/core/errors"
	"github.com/gflydev/core/log"
	"github.com/gflydev/core/utils"
	"github.com/gflydev/storage"
	"github.com/gflydev/storage/ws3"
	"github.com/minio/minio-go/v7"
	"github.com/valyala/fasthttp"
	"os"
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
	//return readFile(c)
	return readObject(c)
	//return readAndCache(c)
}

// Handle is process incoming data from Validate to perform delete a content by id
//
//	Forward Proxy:
//		From http://localhost:7789/objects/one/two/hello.txt
//		To https://s3.ap-southeast-1.wasabisys.com/jivecode/one/two/hello.txt
func readFile(c *core.Ctx) error {
	file, err := os.Open("/Users/vinh/Working/gFlyDev/gflydev-storage/local/local_storage.go")
	if err != nil {
		return errors.New("Error opening file %d", fasthttp.StatusInternalServerError)
	}
	// Note: Do not close the file here with defer as c.Stream() needs the file to remain open
	// The streaming framework will handle closing the file when the stream is complete

	c.Root().SetContentType("application/octet-stream")
	return c.Stream(file)
}

func readAndCache(c *core.Ctx) error {
	// Remove /objects prefix and get the remaining path
	objectPath := strings.TrimPrefix(c.Path(), "/objects/")

	// Create S3 storage with default
	fs := storage.Instance(ws3.Type).(*ws3.Storage)

	byteData, err := fs.Get(objectPath)
	if err != nil {
		log.Errorf("Unable to get object info for %s: %v", objectPath, err)
		return c.Status(core.StatusNotFound).String("Object not found")
	}

	// Create a temporary file
	tempFile, err := os.CreateTemp(core.TempDir, "proxy_*")
	if err != nil {
		c.Status(core.StatusInternalServerError)
		return errors.New("Failed to create temp file: %v", err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tempFile.Name()) // Clean up
	defer func(tempFile *os.File) {
		_ = tempFile.Close()
	}(tempFile)

	// Write the retrieved data to temp file
	_, err = tempFile.Write(byteData)
	if err != nil {
		c.Status(core.StatusInternalServerError)
		return errors.New("Failed to write data to temp file: %v", err)
	}

	// Serve the temp file
	return c.File(tempFile.Name())
}

// Handle is process incoming data from Validate to perform delete a content by id
//
//	Forward Proxy:
//		From http://localhost:7789/objects/one/two/hello.txt
//		To https://s3.ap-southeast-1.wasabisys.com/jivecode/one/two/hello.txt
func readObject(c *core.Ctx) error {
	// Remove /objects prefix and get the remaining path
	objectPath := strings.TrimPrefix(c.Path(), "/objects/")

	// Create S3 storage with default
	fs := storage.Instance(ws3.Type).(*ws3.Storage)

	// Get object info first to determine size and content type
	bucketName := utils.Getenv("WS_BUCKET", "")
	objectInfo, err := fs.S3Client.StatObject(context.TODO(), bucketName, objectPath, minio.StatObjectOptions{})

	if err != nil {
		log.Errorf("Unable to get object info for %s: %v", objectPath, err)
		return c.Status(core.StatusNotFound).String("Object not found")
	}

	// Set appropriate headers
	c.SetHeader("Content-Type", objectInfo.ContentType)
	c.SetHeader("Content-Length", fmt.Sprintf("%d", objectInfo.Size))
	c.SetHeader("Last-Modified", objectInfo.LastModified.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	c.SetHeader("Content-Range", fmt.Sprintf("bytes 0-%d/%d", objectInfo.Size, objectInfo.Size))

	log.Infof("Streaming object: %s (size: %d, type: %s)", objectPath, objectInfo.Size, objectInfo.ContentType)

	// Get the object as a stream instead of loading into memory
	objectStream, err := fs.GetStream(objectPath)
	if err != nil {
		log.Errorf("Unable to get object stream for %s: %v", objectPath, err)
		return c.Status(core.StatusInternalServerError).String("Internal server error")
	}

	return c.Stream(objectStream)
}

// Handle is process incoming data from Validate to perform delete a content by id
//
//	Forward Proxy:
//		From http://localhost:7789/objects/one/two/hello.txt
//		To https://s3.ap-southeast-1.wasabisys.com/jivecode/one/two/hello.txt
func readStream(c *core.Ctx) error {
	// Remove /objects prefix and get the remaining path
	objectPath := strings.TrimPrefix(c.Path(), "/objects/")

	// Create S3 storage with default
	fs := storage.Instance(ws3.Type).(*ws3.Storage)

	// Get object info first to determine size and content type
	bucketName := utils.Getenv("WS_BUCKET", "")
	objectInfo, err := fs.S3Client.StatObject(context.TODO(), bucketName, objectPath, minio.StatObjectOptions{})

	if err != nil {
		log.Errorf("Unable to get object info for %s: %v", objectPath, err)
		return c.Status(core.StatusNotFound).String("Object not found")
	}

	// Set appropriate headers
	c.SetHeader("Content-Type", objectInfo.ContentType)
	c.SetHeader("Content-Length", fmt.Sprintf("%d", objectInfo.Size))
	c.SetHeader("Last-Modified", objectInfo.LastModified.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	c.SetHeader("Content-Range", fmt.Sprintf("bytes 0-%d/%d", objectInfo.Size, objectInfo.Size))

	// Get the object as a stream instead of loading into memory
	objectStream, err := fs.GetStream(objectPath)
	if err != nil {
		log.Errorf("Unable to get object stream for %s: %v", objectPath, err)
		return c.Status(core.StatusInternalServerError).String("Internal server error")
	}

	// Note: Do not close the objectStream here with defer as c.Stream() needs the stream to remain open
	// The streaming framework will handle closing the stream when the streaming operation is complete

	log.Infof("Streaming object: %s (size: %d, type: %s)", objectPath, objectInfo.Size, objectInfo.ContentType)

	// WORK : Read the object data as streaming
	// Refer https://www.youtube.com/watch?v=hpGfLeg9LVM
	/*buf := make([]byte, 1024)
	for {
		n, err := objectStream.Read(buf)
		if n > 0 {
			log.Infof("Data %v", string(buf[:n])) // Process the chunk of data
		}
		if err == io.EOF {
			break // End of object stream
		}
		if err != nil {
			log.Fatalf(err.Error())
		}
	}*/

	// TODO checking
	// 		Refer it https://groups.google.com/g/golang-nuts/c/16cWqMIqvSc?pli=1
	//		Refer https://stackoverflow.com/questions/72628570/stream-media-file-from-minio-object
	/*c.Root().SetBodyStreamWriter(func(w *bufio.Writer) {
		log.Infof("Streaming object: %s (size: %d, type: %s)", objectPath, objectInfo.Size, objectInfo.ContentType)
		//defer func(w *bufio.Writer) {
		//	err := w.Flush()
		//	if err != nil {
		//		log.Errorf("Error flushing response: %v", err)
		//	}
		//}(w)
		buf := make([]byte, 1024)
		for {
			log.Info("read data")
			n, err := objectStream.Read(buf)
			if err != nil {
				log.Infof("read data %d %s", n, err.Error())
			}
			if n > 0 {
				log.Infof("Data %v", string(buf[:n]))
				_, writeErr := w.Write(buf[:n])
				if writeErr != nil {
					log.Errorf("Error writing to response: %v", writeErr)
					return
				}
			}
			if err == io.EOF {
				break // End of object stream
			}
			if !errors.Is(err, io.ErrClosedPipe) || !errors.Is(err, io.ErrUnexpectedEOF) {
				log.Errorf("Error reading from object stream: %v", err)
				return
			}
		}
	})*/

	/*stats, err := objectStream.Stat()
	if err != nil {
		return err
	}*/

	//return c.Stream(objectStream, int(objectInfo.Size))
	return c.Stream(objectStream)
}
