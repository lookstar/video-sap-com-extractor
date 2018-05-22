package app

import (
	"github.com/lookstar/video-sap-com-extractor/pkg/provider"
	"github.com/spf13/cobra"
	"os/exec"
	"strings"
)

type DataCollectorOptions struct {
}

func NewCommandRunCollector() *cobra.Command {

	option := DataCollectorOptions{
	}

	cmdRoot := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			option.RunMount()
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
	return err
}

func (option *DataCollectorOptions) RunDataCollector() {

	provider := provider.NewCollectorProvider()

	if err := provider.DoWork(); err != nil {
		panic(err)
	}
}
