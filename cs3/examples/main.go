package main

import (
	"fmt"
	"github.com/gflydev/core"
	"github.com/gflydev/core/log"
	"github.com/gflydev/core/utils"
	"github.com/gflydev/storage"
	"github.com/gflydev/storage/cs3"
	_ "github.com/gflydev/storage/cs3"
	"github.com/gflydev/view/pongo"
	_ "github.com/joho/godotenv/autoload"
)

// =========================================================================================
//                                     Default API
// =========================================================================================

// NewDefaultApi As a constructor to create new API.
func NewDefaultApi() *DefaultApi {
	return &DefaultApi{}
}

// DefaultApi API struct.
type DefaultApi struct {
	core.Api
}

func (h *DefaultApi) Handle(c *core.Ctx) error {
	return c.JSON(core.Data{
		"name":   core.AppName,
		"server": core.AppURL,
	})
}

// =========================================================================================
//                                     Home page
// =========================================================================================

// NewHomePage As a constructor to create a Home Page.
func NewHomePage() *HomePage {
	return &HomePage{}
}

type HomePage struct {
	core.Page
}

func (m *HomePage) Handle(c *core.Ctx) error {
	checkCS3()

	return c.String("Hello world")
}

func checkCS3() {
	// Create S3 storage with default
	fs := storage.Instance(cs3.Type)

	// Make Dir and Put Object
	if ok := fs.MakeDir("one/two"); ok {
		fs.Put("one/two/hello.txt", "Hello world")
	}

	// Path
	log.Infof("URL object %s", fs.Url("one/two/hello.txt"))

	// Copy & Move
	fs.Copy("one/two/hello.txt", "one/two/world.txt")
	fs.Move("one/two/world.txt", "one/world.txt")

	// Get Object
	data, _ := fs.Get("one/world.txt")
	log.Infof("Read object %s", utils.UnsafeStr(data))

	log.Infof("Size %d", fs.Size("one/two/hello.txt"))

	// Last Modify
	log.Infof("Last Modify %v", fs.LastModified("one/two/hello.txt"))

	// Delete Objects
	fs.Delete("one/two/hello.txt")

	// Delete Dir
	fs.DeleteDir("one")
}

// =========================================================================================
//                                     Routers
// =========================================================================================

func router(g core.IFly) {
	prefixAPI := fmt.Sprintf(
		"/%s/%s",
		utils.Getenv("API_PREFIX", "api"),
		utils.Getenv("API_VERSION", "v1"),
	)

	// API Routers
	g.Group(prefixAPI, func(apiRouter *core.Group) {
		apiRouter.GET("/info", NewDefaultApi())
	})

	// Web Routers
	g.GET("/home", NewHomePage())
}

// =========================================================================================
//                                     Application
// =========================================================================================

func main() {
	app := core.New()

	// Register view
	core.RegisterView(pongo.New())

	// Register storages
	storage.Register(cs3.Type, cs3.New())

	// Register router
	app.RegisterRouter(router)

	app.Run()
}
