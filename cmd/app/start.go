package app

import (
	"io"

	"github.com/lookstar/video-sap-com-extractor/pkg/provider"
	"github.com/spf13/cobra"
)

type DataCollectorOptions struct {
	Url    string
	Output string
}

func NewCommandRunCollector(out, errOut io.Writer) *cobra.Command {

	option := DataCollectorOptions{
		Url:    "",
		Output: "./_output",
	}

	cmdRoot := &cobra.Command{
		Use:   "Collector",
		Short: "Collector is something strange",
		Long:  "Collector is something strange",
		Run: func(cmd *cobra.Command, args []string) {
			option.RunDataCollector()
		},
	}

	flags := cmdRoot.Flags()

	flags.StringVar(&option.Url, "url", option.Url, "input url")
	flags.StringVar(&option.Output, "output", option.Output, "output folder")

	return cmdRoot
}

func (option *DataCollectorOptions) RunDataCollector() {

	provider := provider.NewCollectorProvider(option.Url, option.Output)

	if err := provider.DoWork(); err != nil {
		panic(err)
	}
}
