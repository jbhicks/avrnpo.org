package main

import (
	"log"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func testMain() {
	app := pocketbase.New()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/test123", func(e *core.RequestEvent) error {
			return e.String(200, "TEST ROUTE WORKS")
		})

		se.Router.GET("/", func(e *core.RequestEvent) error {
			log.Printf("[HOME] Path: %s", e.Request.URL.Path)
			return e.String(200, "HOME PAGE")
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
