# gFly Wasabi Storage

https://docs.wasabi.com/docs/how-do-i-use-aws-sdk-for-go-golang-with-wasabi

### Usage

Install
```bash
# Storage
go get -u github.com/gflydev/storage@v1.1.0

# S3 Storage
go get -u github.com/gflydev/storage/ws3@v1.0.0
```

Quick usage `main.go`
```go
import (
    "github.com/gflydev/core"
    "github.com/gflydev/storage"
    storageWS3 "github.com/gflydev/storage/ws3"	
)

func main() {
    // Register WS3 storage
    storage.Register(storageWS3.Type, storageWS3.New())

    // Create S3 storage with default
    fs := storage.Instance(storageWS3.Type)

	// Create folder `foo/bar` and add file `hello.txt`
    if ok := fs.MakeDir("foo/bar"); ok {
        fs.Put("foo/bar/hello.txt", "Hello world")
    }
}
```

### WS3 setting

Make sure WS3 below setting:

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