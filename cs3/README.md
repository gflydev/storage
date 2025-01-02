# gFly Contabo S3 Storage

### Usage

Install
```bash
# Storage
go get -u github.com/gflydev/storage@v1.1.0

# S3 Storage
go get -u github.com/gflydev/storage/cs3@v1.1.0
```

Quick usage `main.go`
```go
import (
    "github.com/gflydev/core"
    "github.com/gflydev/storage"
    storageCS3 "github.com/gflydev/storage/cs3"	
)

func main() {
    // Register CS3 storage
    storage.Register(storageS3.Type, storageCS3.New())

    // Create S3 storage with default
    fs := storage.Instance(storageS3.Type)

	// Create folder `foo/bar` and add file `hello.txt`
    if ok := fs.MakeDir("foo/bar"); ok {
        fs.Put("foo/bar/hello.txt", "Hello world")
    }
}
```

### CS3 setting

Make sure CS3 below setting:
