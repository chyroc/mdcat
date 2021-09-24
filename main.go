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
)

//go:embed static/template.html
var template string

func main() {
	// args
	file, filename, target := getFilepath()

	bs, err := ioutil.ReadFile(file)
	if err != nil {
		assert(fmt.Errorf("read input fail: %w", err))
		return
	}

	html, err := convertWithGitHubApi(filename, template, string(bs))
	if err != nil {
		assert(err)
		return
	}

	if err = ioutil.WriteFile(target, []byte(html), 0666); err != nil {
		assert(err)
	}

	fmt.Println("success")
}

func getFilepath() (string, string, string) {
	if len(os.Args) < 2 {
		assert(fmt.Errorf("mdcat usage: mdcat <markdown_file.md>"))
		return "", "", ""
	}
	file := os.Args[1]
	filename := filepath.Base(file)

	filedir := filepath.Dir(file)
	target := ""
	if strings.Contains(filename, ".") {
		target = filename[:len(filename)-len(filepath.Ext(filename))] + ".html"
	} else {
		target = filename
	}

	return file, filename, filedir + "/" + target
}

func assert(err error) {
	if err != nil {
		log.Fatalln(err)
	}
	panic(err)
}

func convertWithGitHubApi(filename string, htmlTemplate, text string) (string, error) {
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
		htmlTemplate = strings.ReplaceAll(htmlTemplate, "$MD_TITLE", filename)
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
