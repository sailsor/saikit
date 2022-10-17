package new

var (
	mainfc = &FileContent{
		FileName: "main.go",
		Dir:      ".",
		Content: `package main

import (
	"{{.ProPath}}{{.ServerName}}/cmd"
	_ "gorm.io/driver/mysql"
)

func main() {
  cmd.Execute()
}
`,
	}
)

func initMainFiles() {
	Files = append(Files, mainfc)
}
