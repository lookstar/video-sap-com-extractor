package app

import (
	//"github.com/lookstar/video-sap-com-extractor/pkg/collector"
	"github.com/lookstar/video-sap-com-extractor/pkg/queue"
	"github.com/spf13/cobra"
	"os/exec"
	"strings"
	"fmt"
)

type DataCollectorOptions struct {
}

func NewCommandRunCollector() *cobra.Command {

	option := DataCollectorOptions{
	}

	cmdRoot := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			err := option.RunMount()
			if err != nil {
				panic(err) 
			}
			option.RunDataCollector()
		},
	}

	return cmdRoot
}

func (option *DataCollectorOptions) RunMount() error {
	cmd := exec.Command("mount", "-t", "nfs", "10.58.34.199:/hypercd", "/hypercd")
	out, err := cmd.CombinedOutput()
	if strings.Contains(string(out), "already mounted") {
		return nil
	}
	fmt.Println(err)
	return err
}

func (option *DataCollectorOptions) RunDataCollector() {
	handler := queue.NewQueueHandler()
	handler.Run()
}
