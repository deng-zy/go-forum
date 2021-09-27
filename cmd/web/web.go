package main

import (
	"forum/internal/app"
)

func main() {
	go app.RunBus()
	app.Run()
}
