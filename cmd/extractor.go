package main

import (
	"os"

	"github.com/lookstar/video-sap-com-extractor/cmd/app"
)

func main() {
	cmd := app.NewCommandRunCollector(os.Stdout, os.Stderr)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}