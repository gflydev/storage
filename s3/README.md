# gFly S3 Storage

### Usage

Install
```bash
# Storage
go get -u github.com/gflydev/storage@v1.0.1

# S3 Storage
go get -u github.com/gflydev/storage/s3@v1.0.0
```

Quick usage `main.go`
```go
import (
    "github.com/gflydev/core"
    "github.com/gflydev/storage"
    // Auto initial S3 storage
    _ "github.com/gflydev/storage/s3"	
)

func main() {
    // Create S3 storage with default
    fs := storage.Instance(strin(s3.Type))

	// Create folder `foo/bar` and add file `hello.txt`
    if ok := fs.MakeDir("foo/bar"); ok {
        fs.Put("foo/bar/hello.txt", "Hello world")
    }
}
```

### S3 setting

Make sure S3 below setting:

Section `Bucket policy`
```bash
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicRead",
            "Effect": "Allow",
            "Principal": "*",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject",
                "s3:PutObjectAcl",
                "s3:GetObjectAcl",
                "s3:GetObjectAttributes"
            ],
            "Resource": "arn:aws:s3:::gfly-local/*"
        }
    ]
}
```

Section `Cross-origin resource sharing (CORS)`

```bash
[
    {
        "AllowedHeaders": [
            "*"
        ],
        "AllowedMethods": [
            "PUT",
            "POST",
            "DELETE"
        ],
        "AllowedOrigins": [
            "*"
        ],
        "ExposeHeaders": []
    }
]
```