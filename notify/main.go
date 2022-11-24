package main

import (
	"notify/cmd"

	_ "gorm.io/driver/mysql"
)

func main() {
	cmd.Execute()
}
