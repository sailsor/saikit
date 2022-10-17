package new

var (
	cmdfc = &FileContent{
		FileName: "root.go",
		Dir:      "cmd",
		Content: `package cmd

import (
	"os"
	"github.com/spf13/cobra"
{{ range $Import := .ImportServer}}"{{$Import}}"
{{end}}
	{{.PackageName}} "{{.ProPath}}{{.ServerName}}/internal"
	"{{.ProPath}}{{.ServerName}}/internal/infra"
)

var cfgPath []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		appOpt := {{.PackageName}}.AppOptions{}
		app := {{.PackageName}}.NewApp(
			appOpt.WithConfPath(cfgPath))

		app.Infra = infra.NewInfra()

{{ range $Server := .RunTrans}}{{$Server}}
{{end}}
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
`,
	}
)

func initCmdFiles() {
	Files = append(Files, cmdfc)
}
