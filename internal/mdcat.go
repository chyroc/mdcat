package internal

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed template.html
var template string

func Run(file, title, output string) error {
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

	html, err := convertWithGitHubApi(title, string(bs))
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

func convertWithGitHubApi(title string, text string) (string, error) {
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

	htmlTemplate := template // 防止 template 被修改
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
