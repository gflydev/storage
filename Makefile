mod:
	go list -m --versions

mod.local:
	go list -m -versions github.com/gflydev/storage/local

mod.s3:
	go list -m -versions github.com/gflydev/storage/s3

mod.cs3:
	go list -m -versions github.com/gflydev/storage/cs3

mod.ws3:
	go list -m -versions github.com/gflydev/storage/ws3
