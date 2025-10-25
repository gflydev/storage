package ws3

import (
	"context"
	"fmt"
	"github.com/gflydev/core"
	"github.com/gflydev/core/errors"
	"github.com/gflydev/core/log"
	"github.com/gflydev/core/utils"
	"github.com/gflydev/storage"
	"github.com/gflydev/storage/local"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ========================================================================================
//                                        Structure
// ========================================================================================

const (
	Type = storage.Type("ws3")
)

// Endpoint returns the endpoint host without protocol scheme for minio client initialization.
// The function strips both "https://" and "http://" prefixes from the endpoint string.
func endpointURL() string {
	endPoint := utils.Getenv("WS_ENDPOINT", "s3.ap-southeast-1.wasabisys.com")

	// Strip protocol scheme from endpoint for minio client
	return strings.TrimPrefix(strings.TrimPrefix(endPoint, "https://"), "http://")
}

// New Create S3 Storage with basics info.
func New() *Storage {
	accessKey := utils.Getenv("WS_ACCESS_KEY_ID", "")
	secretKey := utils.Getenv("WS_SECRET_ACCESS_KEY", "")

	// Initialize minio client object.
	minioClient, err := minio.New(endpointURL(), &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	return &Storage{
		S3Client: minioClient,
	}
}

type Storage struct {
	S3Client *minio.Client
}

// ========================================================================================
//                                   Implement IStorage
// ========================================================================================

func (s *Storage) Put(path, contents string) bool {
	localStorage := local.New()

	// Put content to temporary dir at local.
	fileName := filepath.Base(path)
	tempPath := fmt.Sprintf("%s/%s", core.TempDir, fileName)
	localStorage.Put(tempPath, contents)

	// Open file source
	file, err := os.Open(filepath.Clean(tempPath))
	if err != nil {
		log.Errorf("Unable create file %q. Here's why: %v\n", tempPath, err)

		return false
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf("Unable to close file %q. Here's why: %v\n", tempPath, err)
		}
	}(file)

	return s.PutFile(path, file)
}

// PutData Create file by content
func (s *Storage) PutData(path string, contents []byte) bool {
	localStorage := local.New()

	// Put content to temporary dir at local.
	fileName := filepath.Base(path)
	tempPath := fmt.Sprintf("%s/%s", core.TempDir, fileName)
	localStorage.PutData(tempPath, contents)

	// Open file source
	file, err := os.Open(filepath.Clean(tempPath))
	if err != nil {
		log.Errorf("Unable create file %q. Here's why: %v\n", tempPath, err)

		return false
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf("Unable to close file %q. Here's why: %v\n", tempPath, err)
		}
	}(file)

	return s.PutFile(path, file)
}

func (s *Storage) PutFile(path string, fileSource *os.File) bool {
	bucket := utils.Getenv("WS_BUCKET", "")

	_, err := s.S3Client.PutObject(context.TODO(), bucket, path, fileSource, -1, minio.PutObjectOptions{ContentType: core.MIMEOctetStream})

	if err != nil {
		log.Errorf("Unable to write file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) PutFilepath(path, filePath string, options ...interface{}) bool {
	bucket := utils.Getenv("WS_BUCKET", "")

	fileSource, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		log.Errorf("Unable to read file %q. Here's why: %v\n", filePath, err)

		return false
	}
	defer func(fileSource *os.File) {
		err := fileSource.Close()
		if err != nil {
			log.Errorf("Unable to close file %q. Here's why: %v\n", filePath, err)
		}
	}(fileSource)

	_, err = s.S3Client.PutObject(context.TODO(), bucket, path, fileSource, -1, minio.PutObjectOptions{ContentType: core.MIMEOctetStream})
	if err != nil {
		log.Errorf("Unable to write file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) Delete(path string) bool {
	bucket := utils.Getenv("WS_BUCKET", "")

	err := s.S3Client.RemoveObject(context.TODO(), bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		log.Errorf("Unable to delete file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) Copy(from, to string) bool {
	bucket := utils.Getenv("WS_BUCKET", "")

	srcOpts := minio.CopySrcOptions{
		Bucket: bucket,
		Object: from,
	}
	dstOpts := minio.CopyDestOptions{
		Bucket: bucket,
		Object: to,
	}
	_, err := s.S3Client.CopyObject(context.TODO(), dstOpts, srcOpts)
	if err != nil {
		log.Errorf("Unable to copy file %s to %s. Here's why: %v\n", from, to, err)

		return false
	}

	return true
}

func (s *Storage) Move(from, to string) bool {
	if s.Copy(from, to) {
		return s.Delete(from)
	}

	return false
}

func (s *Storage) Exists(path string) bool {
	return s.Size(path) != 0
}

func (s *Storage) Get(path string) ([]byte, error) {
	result, err := s.getObject(path)

	if err != nil {
		return nil, err
	}

	defer func(obj *minio.Object) {
		err := obj.Close()
		if err != nil {
			log.Errorf("Unable to close object. Here's why: %v\n", err)
		}
	}(result)

	body, err := io.ReadAll(result)
	if err != nil {
		log.Errorf("Unable read object body from %v. Here's why: %v\n", path, err)
	}

	return body, nil
}

func (s *Storage) Size(path string) int64 {
	bucket := utils.Getenv("WS_BUCKET", "")

	result, err := s.S3Client.StatObject(context.TODO(), bucket, path, minio.StatObjectOptions{})
	if err != nil {
		log.Errorf("Unable to get object size from %v. Here's why: %v\n", path, err)

		return 0
	}

	return result.Size
}

func (s *Storage) LastModified(path string) time.Time {
	bucket := utils.Getenv("WS_BUCKET", "")

	result, err := s.S3Client.StatObject(context.TODO(), bucket, path, minio.StatObjectOptions{})

	if err != nil {
		log.Errorf("Unable to get info of %s. Here's why: %v\n", path, err)

		return time.Time{}
	}

	return result.LastModified
}

// Url Get public URL of an object via path
// Pattern URL `https://s3.<region>.wasabisys.com/<bucket>/<key>`
func (s *Storage) Url(path string) string {
	region := utils.Getenv("WS_REGION", "")
	bucket := utils.Getenv("WS_BUCKET", "")

	return fmt.Sprintf("https://%s.wasabisys.com/%s/%s",
		region,
		bucket,
		strings.TrimPrefix(filepath.ToSlash(path), "/"),
	)
}

func (s *Storage) MakeDir(dir string) bool {
	return s.Put(fmt.Sprintf("%s/%s", dir, storage.DirFileHolder), "Place holder")
}

func (s *Storage) DeleteDir(dir string) bool {
	bucket := utils.Getenv("WS_BUCKET", "")

	// Get all objects in dir
	// Note: Can not delete a dir have children object.
	objectCh := s.S3Client.ListObjects(context.TODO(), bucket, minio.ListObjectsOptions{
		Prefix:    dir,
		Recursive: true,
	})

	var objectNames []string

	// Collect children objects
	for object := range objectCh {
		if object.Err != nil {
			log.Errorf("Unable to list objects from dir %v. Here's why: %v\n", dir, object.Err)
			return false
		}
		objectNames = append(objectNames, object.Key)
	}

	// Append current object if not already included
	objectNames = append(objectNames, dir)

	// Delete objects
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for _, objectName := range objectNames {
			objectsCh <- minio.ObjectInfo{Key: objectName}
		}
	}()

	for rErr := range s.S3Client.RemoveObjects(context.TODO(), bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if rErr.Err != nil {
			log.Errorf("Unable to delete object %s from bucket %v. Here's why: %v\n", rErr.ObjectName, dir, rErr.Err)
			return false
		}
	}

	return true
}

func (s *Storage) Append(path, data string) bool {
	log.Errorf("Unable to append data %s into %s. Here's why: %v\n", path, data, errors.NotImplemented)

	return false
}

// GetStream returns a stream (io.ReadCloser) for the object at the given path
// This allows for efficient streaming without loading the entire file into memory
func (s *Storage) GetStream(path string) (io.ReadCloser, error) {
	return s.getObject(path)
}

func (s *Storage) getObject(path string) (*minio.Object, error) {
	bucket := utils.Getenv("WS_BUCKET", "")

	result, err := s.S3Client.GetObject(context.TODO(), bucket, path, minio.GetObjectOptions{})
	if err != nil {
		log.Errorf("Unable to get object %s. Here's why: %v\n", path, err)

		return nil, err
	}

	return result, nil
}
