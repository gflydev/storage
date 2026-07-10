package cs3

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
	Type = storage.Type("cs3")
)

// Endpoint returns the endpoint host without protocol scheme for minio client initialization.
// The function strips both "https://" and "http://" prefixes from the endpoint string.
func endpointURL() string {
	endPoint := utils.Getenv("CS_ENDPOINT", "sin1.contabostorage.com")

	// Strip protocol scheme from endpoint for minio client
	return strings.TrimPrefix(strings.TrimPrefix(endPoint, "https://"), "http://")
}

// New Create S3 Storage with basics info.
func New() *Storage {
	accessKey := utils.Getenv("CS_ACCESS_KEY_ID", "")
	secretKey := utils.Getenv("CS_SECRET_ACCESS_KEY", "")

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
//                                     Implement IStorage
// ========================================================================================

func (s *Storage) Put(path, contents string) bool {
	localStorage := local.New()

	// Put content to temporary dir at local.
	tempPath := tempFilePath(path)
	localStorage.Put(tempPath, contents)
	// Remove the temporary file once the upload is done.
	defer removeTempFile(tempPath)

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
	tempPath := tempFilePath(path)
	localStorage.PutData(tempPath, contents)
	// Remove the temporary file once the upload is done.
	defer removeTempFile(tempPath)

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
	bucket := utils.Getenv("CS_BUCKET", "")

	_, err := s.S3Client.PutObject(context.TODO(), bucket, path, fileSource, -1, minio.PutObjectOptions{ContentType: core.MIMEOctetStream})

	if err != nil {
		log.Errorf("Unable to write file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) PutFilepath(path, filePath string, options ...interface{}) bool {
	bucket := utils.Getenv("CS_BUCKET", "")
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
	bucket := utils.Getenv("CS_BUCKET", "")
	err := s.S3Client.RemoveObject(context.TODO(), bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		log.Errorf("Unable to delete file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) Copy(from, to string) bool {
	bucket := utils.Getenv("CS_BUCKET", "")

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

		return nil, err
	}

	return body, nil
}

func (s *Storage) Size(path string) int64 {
	bucket := utils.Getenv("CS_BUCKET", "")

	result, err := s.S3Client.StatObject(context.TODO(), bucket, path, minio.StatObjectOptions{})
	if err != nil {
		log.Errorf("Unable to get object size from %v. Here's why: %v\n", path, err)

		return 0
	}

	return result.Size
}

func (s *Storage) LastModified(path string) time.Time {
	bucket := utils.Getenv("CS_BUCKET", "")

	result, err := s.S3Client.StatObject(context.TODO(), bucket, path, minio.StatObjectOptions{})

	if err != nil {
		log.Errorf("Unable to get info of %s. Here's why: %v\n", path, err)

		return time.Time{}
	}

	return result.LastModified
}

// Url Get public URL of an object via path
// Pattern URL `https://<region>.contabostorage.com/<bucket_code>:<bucket>/<key>`
func (s *Storage) Url(path string) string {
	bucket := utils.Getenv("CS_BUCKET", "")
	region := utils.Getenv("CS_REGION", "")
	bucketCode := utils.Getenv("CS_BUCKET_CODE", "")

	return fmt.Sprintf("https://%s.contabostorage.com/%s:%s/%s",
		region,
		bucketCode,
		bucket,
		strings.TrimPrefix(filepath.ToSlash(path), "/"),
	)
}

func (s *Storage) MakeDir(dir string) bool {
	return s.Put(fmt.Sprintf("%s/%s", dir, storage.DirFileHolder), "Place holder")
}

func (s *Storage) DeleteDir(dir string) bool {
	bucket := utils.Getenv("CS_BUCKET", "")

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

// GetStream returns a stream (io.ReadCloser) for the object at the given path.
// This allows for efficient streaming without loading the entire file into memory.
// The caller is responsible for closing the returned reader.
func (s *Storage) GetStream(path string) (io.ReadCloser, error) {
	return s.getObject(path)
}

// tempFilePath builds a collision-resistant local path used to buffer an object
// before uploading it. Flattening the full object path (instead of using only its
// base name) prevents two different objects that share a file name from clobbering
// each other's temporary file during concurrent uploads.
func tempFilePath(path string) string {
	safe := strings.ReplaceAll(strings.TrimPrefix(filepath.ToSlash(path), "/"), "/", "_")

	return fmt.Sprintf("%s/%d-%s", core.TempDir, os.Getpid(), safe)
}

// removeTempFile deletes a temporary upload buffer, ignoring a missing file and
// logging any other failure so temp files are not leaked on disk.
func removeTempFile(tempPath string) {
	if err := os.Remove(filepath.Clean(tempPath)); err != nil && !os.IsNotExist(err) {
		log.Errorf("Unable to remove temporary file %q. Here's why: %v\n", tempPath, err)
	}
}

func (s *Storage) getObject(path string) (*minio.Object, error) {
	bucket := utils.Getenv("CS_BUCKET", "")

	result, err := s.S3Client.GetObject(context.TODO(), bucket, path, minio.GetObjectOptions{})
	if err != nil {
		log.Errorf("Unable to get object %s. Here's why: %v\n", path, err)

		return nil, err
	}

	return result, nil
}
