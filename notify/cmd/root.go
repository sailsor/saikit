package cmd

import (
	"notify/internal/transports/http"
	"os"

	"github.com/spf13/cobra"

	"notify/internal"
	"notify/internal/infra"
)

var cfgPath []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		option := notify.AppOptions{}
		app := notify.NewApp(option.WithConfPath(cfgPath))

		app.Infra = infra.NewInfra()

		app.RegisterTran(http.NewGinServer(app))

		app.Start()
		app.AwaitSignal()

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&cfgPath, "config", []string{"conf"}, "config path")
}
