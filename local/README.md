# gFly Local Storage

### Usage

Install
```bash
# Storage
go get -u github.com/gflydev/storage@v1.1.0

# Local Storage
go get -u github.com/gflydev/storage/local@v1.1.0
```

Quick usage `main.go`
```go
import (
    "github.com/gflydev/core"
    "github.com/gflydev/storage"
    storageLocal "github.com/gflydev/storage/local"	
)

func main() {
    // Register Local storage
    storage.Register(storageLocal.Type, storageLocal.New())

    // Create file storage with default
    fs := storage.Instance()

	// Create folder `foo/bar` and add file `hello.txt`
    if ok := fs.MakeDir("foo/bar"); ok {
        fs.Put("foo/bar/hello.txt", "Hello world")
    }
}
```
