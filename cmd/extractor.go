package main

import (
	"github.com/lookstar/video-sap-com-extractor/cmd/app"
)

// input
// ENV: NFS_HOME 10.58.34.199:/hypercd
// ENV: MQ_URL amqp://root:xxxx@10.58.116.110:45268/
// ENV: REDIS_URL 127.0.0.1:6379
// ENV: REDIS_PORT 29
func main() {
	cmd := app.NewCommandRunCollector()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}