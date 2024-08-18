# gFly Storage - File system

    Copyright © 2023, gFly
    https://www.gfly.dev
    All rights reserved.

# gFly Storage

### Usage

Install
```bash
# Storage
go get -u github.com/gflydev/storage@v1.0.1

# Local Storage
go get -u github.com/gflydev/storage/local@v1.0.1
```

Quick usage `main.go`
```go
import (
    "github.com/gflydev/core"
    "github.com/gflydev/storage"
    // Auto initial local storage
    _ "github.com/gflydev/storage/local"	
)

func main() {
    // Create file storage with default
    fs := storage.Instance()

	// Create folder `foo/bar` and add file `hello.txt`
    if ok := fs.MakeDir("foo/bar"); ok {
        fs.Put("foo/bar/hello.txt", "Hello world")
    }
}
```
