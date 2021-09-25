package internal

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

//go:embed template.html
var template string

var (
	done             = map[string]bool{}
	mainInputFile    string
	mainHtmlBaseFile string
)

func Run(inputFile, title string, link bool, output string) error {
	outputFile := genTargetFilePath(inputFile, output)
	mainInputFile = inputFile
	mainHtmlBaseFile = filepath.Base(outputFile)

	_, err := run(inputFile, outputFile, title, link)
	if err != nil {
		return err
	}

	log.Printf("mdcat success")

	return nil
}

func run(inputFile, outputFile, title string, link bool) (bool, error) {
	title = getTitle(inputFile, title)

	if done[inputFile] {
		return true, nil
	}
	done[inputFile] = true

	log.Printf("cat %q -> %q", inputFile, outputFile)

	bs, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return false, fmt.Errorf("read input fail: %w", err)
	}

	html, err := convertWithGitHubApi(title, string(bs))
	if err != nil {
		return false, err
	}

	if link {
		replacedHTML, hrefFiles, replaced := replaceChildMarkdownLink(inputFile, html)
		if replaced {
			html = replacedHTML
		}
		for _, href := range hrefFiles {
			hrefInputFile := getAbsoluteFilePathOfTwoFile(inputFile, href)
			hrefOutputFile := genHtmlName(hrefInputFile, getAbsoluteFilePathOfTwoFile(outputFile, href))

			if !done[hrefInputFile] {
				_, _ = run(hrefInputFile, hrefOutputFile, "", link)
			}
		}
	}

	if err = writeFile(outputFile, html); err != nil {
		return false, fmt.Errorf("write file %q failed: %w", outputFile, err)
	}
	return false, nil
}

func replaceChildMarkdownLink(inputFile, parentHtml string) (string, []string, bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(parentHtml))
	if err != nil {
		return "", nil, false
	}
	links := []string{}
	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		href := selection.AttrOr("href", "")
		if !strings.HasSuffix(href, ".md") || href == ".md" {
			return
		}
		hrefInputFile := getAbsoluteFilePathOfTwoFile(inputFile, href)
		if _, err := readMd(hrefInputFile); err != nil {
			return
		}

		links = append(links, href)
		selection.SetAttr("href", genHtmlName(hrefInputFile, href))
	})

	h, _ := doc.Html()
	return h, links, true
}

func genTargetFilePath(inputFile string, output string) string {
	if output != "" {
		return output
	}

	return genHtmlName(inputFile, inputFile)
}

func genHtmlName(inputFile, file string) (res string) {
	if strings.Contains(file, ".") {
		file = file[:len(file)-len(filepath.Ext(file))] + ".html"
	} else {
		file = file + ".html"
	}
	if inputFile == mainInputFile {
		file = filepath.Join(filepath.Dir(file), mainHtmlBaseFile)
	}
	return file
}

func getTitle(file string, title string) string {
	if title != "" {
		return title
	}
	return filepath.Base(file)
}

func convertWithGitHubApi(title string, text string) (string, error) {
	htmlString, err := githubConvertMarkdown(text)
	if err != nil {
		return "", err
	}

	htmlTemplate := template // 防止 template 被修改
	htmlTemplate = strings.ReplaceAll(htmlTemplate, "$MD_TITLE", title)
	htmlTemplate = strings.ReplaceAll(htmlTemplate, "$MD_HTML", htmlString)

	return htmlTemplate, nil
}

var readMd = func(path string) (string, error) {
	bs, err := ioutil.ReadFile(path)
	return string(bs), err
}

func githubConvertMarkdown(text string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	body := strings.NewReader(fmt.Sprintf(`{"text": %q}`, text))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/markdown", body)
	if err != nil {
		return "", fmt.Errorf("request github fail: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request github fail: %w", err)
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("request github fail: %w", err)
	}

	if resp.StatusCode == 200 {
		return string(bs), nil
	}

	res := struct {
		Message string `json:"message"`
	}{}
	if err = json.Unmarshal(bs, &res); err == nil {
		return "", fmt.Errorf("request github fail: %s", res.Message)
	}
	return "", fmt.Errorf("request github fail: %s", bs)
}

func writeFile(targetFile, content string) error {
	_ = os.MkdirAll(filepath.Dir(targetFile), 0o777)
	if err := ioutil.WriteFile(targetFile, []byte(content), 0o666); err != nil {
		return fmt.Errorf("write file %q failed: %w", targetFile, err)
	}
	return nil
}

func getAbsoluteFilePathOfTwoFile(a, b string) string {
	return filepath.Join(filepath.Dir(a), b)
}
