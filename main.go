package main

import (
	"log"
	"os"

	"github.com/chyroc/mdcat/mdcat"
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
			&cli.BoolFlag{
				Name:  "link",
				Value: false,
				Usage: "convert linked markdown file",
			},
			&cli.StringFlag{
				Name:  "config",
				Value: "",
				Usage: "config file",
			},
		},
		Action: func(c *cli.Context) error {
			setupLog()

			file := c.Args().First()
			config, err := mdcat.ParseConfig(
				c.String("config"),
				c.String("title"),
				c.String("output"),
				c.Bool("link"))
			if err != nil {
				return err
			}
			if file == "" {
				return cli.ShowAppHelp(c)
			}

			log.Printf("input: file=%q, config=%#v", file, config)

			return mdcat.Run(file, config)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func setupLog() {
	log.SetPrefix("[mdcat] ")
	log.SetFlags(log.LstdFlags)
}
