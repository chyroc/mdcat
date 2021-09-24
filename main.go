package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

//go:embed static/template.html
var template string

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

			return run(file, title, output)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(file, title, output string) error {
	// args
	file, filename, target, err := getFilepath(file)
	if err != nil {
		return err
	}
	if title == "" {
		title = filename
	}
	if output == "" {
		output = target
	}

	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read input fail: %w", err)
	}

	html, err := convertWithGitHubApi(title, template, string(bs))
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(output, []byte(html), 0o666); err != nil {
		return err
	}

	fmt.Println("success")

	return nil
}

func getFilepath(file string) (string, string, string, error) {
	filename := filepath.Base(file)

	filedir := filepath.Dir(file)
	target := ""
	if strings.Contains(filename, ".") {
		target = filename[:len(filename)-len(filepath.Ext(filename))] + ".html"
	} else {
		target = filename
	}

	return file, filename, filedir + "/" + target, nil
}

func convertWithGitHubApi(title string, htmlTemplate, text string) (string, error) {
	body := strings.NewReader(fmt.Sprintf(`{"text": %q}`, text))
	req, err := http.NewRequest(http.MethodPost, "https://api.github.com/markdown", body)
	if err != nil {
		return "", fmt.Errorf("request github fail: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request github fail: %w", err)
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("request github fail: %w", err)
	}

	if resp.StatusCode == 200 {
		htmlTemplate = strings.ReplaceAll(htmlTemplate, "$MD_TITLE", title)
		htmlTemplate = strings.ReplaceAll(htmlTemplate, "$MD_HTML", string(bs))

		return htmlTemplate, nil
	}

	res := struct {
		Message string `json:"message"`
	}{}
	if err = json.Unmarshal(bs, &res); err == nil {
		return "", fmt.Errorf("request github fail: %s", res.Message)
	}
	return "", fmt.Errorf("request github fail: %s", bs)
}
