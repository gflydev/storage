module examples

go 1.24.0

toolchain go1.24.3

replace github.com/gflydev/storage => ../../

replace github.com/gflydev/storage/s3 => ../

require (
	github.com/gflydev/core v1.17.11
	github.com/gflydev/storage v1.1.6
	github.com/gflydev/storage/s3 v1.1.6
	github.com/gflydev/view/pongo v1.0.3
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.39.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.31.9 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.13 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.8.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.88.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.34.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.4 // indirect
	github.com/aws/smithy-go v1.23.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/flosch/pongo2/v6 v6.0.0 // indirect
	github.com/gflydev/storage/local v1.1.6 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.68.0 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
)
