package main

import (
	"fmt"
	"os"

	"github.com/lexycore/gitlab-tools/internal/cli"
)

func main() {
	app := cli.CreateCLI()
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
