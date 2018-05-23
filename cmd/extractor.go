package main

import (
	"github.com/lookstar/video-sap-com-extractor/cmd/app"
)

func main() {
	cmd := app.NewCommandRunCollector()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}