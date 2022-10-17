package cmd

import (
	"code.jshyjdtech.com/godev/hykit/tool/protoc"
	"github.com/spf13/cobra"
)

var protocCmd = &cobra.Command{
	Use:   "protoc",
	Short: "grpc protoc",
	Long: `1：在执行前需要注意，先把proto 文件 复制到项目下，
2：需要在项目根目录下执行
生成的protobuf文件会被放到项目的 internal/infra/third_party/package/*.pb.go,
`,
	Run: func(cmd *cobra.Command, args []string) {
		protocer := protoc.NewProtocer(
			protoc.WithProtocLogger(logger),
		)
		protocer.Run(v)
	},
}

func init() {
	rootCmd.AddCommand(protocCmd)

	protocCmd.Flags().StringP("from_proto", "f", "", "proto 文件")

	protocCmd.Flags().StringP("target", "t", "internal/infra/third_party/protobuf", "生成的源码保存路径")

	protocCmd.Flags().BoolP("mock", "m", false, "生成 mock 文件")

	protocCmd.Flags().StringP("package", "p", "", "package名称")

	err := v.BindPFlags(protocCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}
