package main

import (
	"log"
	"os"

	"github.com/chyroc/mdcat/internal"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "mdcat",
		Usage: "convert markdown file to github style html page",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "title",
				Value: "",
				Usage: "html page title",
			},
			&cli.StringFlag{
				Name:  "output",
				Value: "",
				Usage: "output filename, default is <input>.html",
			},
		},
		Action: func(c *cli.Context) error {
			file := c.Args().First()
			title := c.String("title")
			output := c.String("output")

			if file == "" {
				return cli.ShowAppHelp(c)
			}

			return internal.Run(file, title, output)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
