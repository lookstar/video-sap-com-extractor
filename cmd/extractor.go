package main

import (
	"github.com/lookstar/video-sap-com-extractor/cmd/app"
)

// input
// ENV: NFS_HOME 10.58.34.199:/hypercd
// ENV: MQ_URL amqp://root:xxxx@10.58.116.110:45268/
func main() {
	cmd := app.NewCommandRunCollector()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}