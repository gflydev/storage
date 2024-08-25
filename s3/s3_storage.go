package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gflydev/core"
	"github.com/gflydev/core/errors"
	"github.com/gflydev/core/log"
	"github.com/gflydev/core/utils"
	"github.com/gflydev/storage"
	"github.com/gflydev/storage/local"
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
	Type = storage.Type("s3")
)

var (
	bucket = utils.Getenv("AWS_S3_BUCKET", "gfly")
	region = utils.Getenv("AWS_S3_REGION", "us-west-1")
)

// New Create S3 Storage with basics info.
func New() *Storage {
	// Load the Shared AWS Configuration (~/.aws/config). Note: Also load combine .env file.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	return &Storage{
		S3Client: s3.NewFromConfig(cfg),
	}
}

type Storage struct {
	S3Client *s3.Client
}

// ========================================================================================
// 									Implement IStorage
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
	_, err := s.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		Body:   fileSource,
	})
	if err != nil {
		log.Errorf("Unable to write file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) PutFilepath(path, filePath string, options ...interface{}) bool {
	fileSource, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		log.Errorf("Unable to read file %q. Here's why: %v\n", filePath, err)

		return false
	}

	_, err = s.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		Body:   fileSource,
	})
	if err != nil {
		log.Errorf("Unable to write file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) Delete(path string) bool {
	_, err := s.S3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		log.Errorf("Unable to delete file %q. Here's why: %v\n", path, err)

		return false
	}

	return true
}

func (s *Storage) Copy(from, to string) bool {
	_, err := s.S3Client.CopyObject(context.TODO(), &s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s", bucket, from)),
		Key:        aws.String(to),
	})
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

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("Unable to close body. Here's why: %v\n", err)
		}
	}(result.Body)

	body, err := io.ReadAll(result.Body)
	if err != nil {
		log.Errorf("Unable read object body from %v. Here's why: %v\n", path, err)
	}

	return body, nil
}

func (s *Storage) Size(path string) int64 {
	result, err := s.S3Client.GetObjectAttributes(context.TODO(), &s3.GetObjectAttributesInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		ObjectAttributes: []types.ObjectAttributes{
			types.ObjectAttributesObjectSize,
		},
	})
	if err != nil {
		return 0
	}

	return *result.ObjectSize
}

func (s *Storage) LastModified(path string) time.Time {
	result, err := s.S3Client.GetObjectAttributes(context.TODO(), &s3.GetObjectAttributesInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		ObjectAttributes: []types.ObjectAttributes{
			types.ObjectAttributesObjectParts,
		},
	})

	if err != nil {
		log.Errorf("Unable to get info of %s. Here's why: %v\n", path, err)

		return time.Time{}
	}

	return *result.LastModified
}

// Url Get public URL of an object via path
//
//	Pattern URL (Use it) `https://<bucket-name>.s3.<region>.amazonaws.com/<key>`
//	Pattern URL `https://<region>.amazonaws.com/<bucket-name>/<key>`
func (s *Storage) Url(path string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		bucket,
		region,
		strings.TrimPrefix(filepath.ToSlash(path), "/"),
	)
}

func (s *Storage) MakeDir(dir string) bool {
	return s.Put(fmt.Sprintf("%s/.info", dir), "Info")
}

func (s *Storage) DeleteDir(dir string) bool {
	// Get all objects in dir
	// Note: Can not delete a dir have children object.
	result, err := s.S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(dir),
	})
	var contents []types.Object
	if err != nil {
		log.Errorf("Unable to list objects from dir %v. Here's why: %v\n", dir, err)
	} else {
		contents = result.Contents
	}

	var objectIds []types.ObjectIdentifier

	// Collect children objects
	for _, object := range contents {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: object.Key})
	}

	// Append current object
	objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(dir)})

	// Delete objects
	_, err = s.S3Client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{Objects: objectIds},
	})

	if err != nil {
		log.Errorf("Unable to delete object from bucket %v. Here's why: %v\n", dir, err)

		return false
	}

	return true
}

func (s *Storage) Append(path, data string) bool {
	log.Errorf("Unable to append data %s into %s. Here's why: %v\n", path, data, errors.NotYetImplemented.Error())

	return false
}

func (s *Storage) getObject(path string) (*s3.GetObjectOutput, error) {
	result, err := s.S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		log.Errorf("Unable to get object %s. Here's why: %v\n", path, err)

		return nil, err
	}

	return result, nil
}
