package mdcat

import (
	"bytes"
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
	template2 "text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
)

//go:embed template.html
var template string

var (
	done             = map[string]bool{}
	metas            = map[string]*mdMeta{}
	mainInputFile    string
	mainHtmlBaseFile string
	conf             *Config
)

type mdMeta struct {
	Source string
	Title  string
	Slug   string
}

func Run(inputFile string, config *Config) error {
	outputFile := genTargetFilePath(inputFile, config.Output, "", "")
	mainInputFile = inputFile
	mainHtmlBaseFile = filepath.Base(outputFile)

	_, err := run(inputFile, outputFile, config.Title, config.Link, config.Gitalk, config.FastClick)
	if err != nil {
		return err
	}

	log.Printf("mdcat success")

	return nil
}

func run(inputFile, outputFile, title string, link bool, configGitalk *ConfigGitalk, fastClick bool) (bool, error) {
	title = getTitle(inputFile, title)

	if done[inputFile] {
		return true, nil
	}
	done[inputFile] = true

	log.Printf("cat %q -> %q", inputFile, outputFile)

	sourceText := ""
	meta := getMeta(inputFile)
	slug := ""
	if meta == nil {
		bs, err := ioutil.ReadFile(inputFile)
		if err != nil {
			return false, fmt.Errorf("read input fail: %w", err)
		}
		sourceText = string(bs)
	} else {
		sourceText = meta.Source
		slug = meta.Slug
	}

	html, err := convertWithGitHubApi(title, sourceText, configGitalk.clone(slug), fastClick)
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
			hrefOutputFile := genTargetFilePath(hrefInputFile, "", outputFile, href)

			if !done[hrefInputFile] {
				_, _ = run(hrefInputFile, hrefOutputFile, "", link, configGitalk, fastClick)
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
		targetHref := genHtmlName(hrefInputFile, href) // a/b/c.index
		if conf != nil && conf.IsOmitHtmlSuffix {
			targetHref = targetHref[:len(targetHref)-len(filepath.Ext(targetHref))] + "/" // a/b/c/
		}
		selection.SetAttr("href", targetHref)
	})

	h, _ := doc.Html()
	return h, links, true
}

func genTargetFilePath(inputFile string, output string, parentOutputFile, href string) string {
	// 如果指定了 output，直接返回，只有入口文件，可能走这里
	if output != "" {
		return output
	}

	hrefAbsoluteFile := ""
	if parentOutputFile != "" {
		// 如果 parentOutputFile 不为空，说明不是入口文件，需要根据父文件和href，计算出当前文件的路径
		hrefAbsoluteFile = getAbsoluteFilePathOfTwoFile(parentOutputFile, href)
	} else {
		// parentOutputFile 为空，说明 inputFile 就是入口文件，不需要计算，直接用即可
		hrefAbsoluteFile = inputFile
	}

	target := genHtmlName(inputFile, hrefAbsoluteFile) // a/b/c.html

	if conf != nil && conf.IsOmitHtmlSuffix {
		prefix := target[:len(target)-len(filepath.Ext(target))]
		target = prefix + "/index.html" // a/b/c/index.html
	}

	return target
}

// relativePathCurrentDir 是相对于当前目录的相对路径
// file 是一个指向这个文件的不知道「当前目录」的相对路径，可能是 ./a ../a .../a ../file/a 等
// 这个函数的目的，是生成和 file 同层级的，以 html 为后缀的文件相对路径，如 ./a.html ../a.html ../file/a.html 等
func genHtmlName(curRelativePath, anyRelativePath string) (res string) {
	// 将 file 的后缀替换为 .html
	if strings.Contains(anyRelativePath, ".") {
		anyRelativePath = anyRelativePath[:len(anyRelativePath)-len(filepath.Ext(anyRelativePath))] + ".html"
	} else {
		anyRelativePath = anyRelativePath + ".html"
	}

	// 如果这个文件是入口文件，则路径是定死的，直接返回
	if curRelativePath == mainInputFile {
		return filepath.Join(filepath.Dir(anyRelativePath), mainHtmlBaseFile)
	}

	// markdown 文件中有 slug meta，替换后缀名
	meta := getMeta(curRelativePath)
	if meta != nil && meta.Slug != "" {
		anyRelativePath = replaceSlugHtmlName(anyRelativePath, meta.Slug)
	}

	// if

	return anyRelativePath
}

func replaceSlugHtmlName(path, slug string) string {
	dir, base := filepath.Split(path)
	ext := filepath.Ext(base)
	if ext == "" {
		// "/a/b/c" => ["/a/b", "c"] => ["/a/b", "slug"] => "/a/b/slug"
		return dir + slug
	}

	// "/a/b/c.html" => ["/a/b", "c.html"] => ["/a/b", "slug.html"] => "/a/b/slug.html"
	return dir + slug + ".html"
}

func getTitle(file string, title string) string {
	if title != "" {
		return title
	}
	meta := getMeta(file)
	if meta == nil || meta.Title == "" {
		return filepath.Base(file)
	}
	return meta.Title
}

func convertWithGitHubApi(title, text string, configGitalk *ConfigGitalk, fastClick bool) (string, error) {
	htmlString, err := githubConvertMarkdown(text)
	if err != nil {
		return "", err
	}

	// htmlTemplate := template // 防止 template 被修改
	// htmlTemplate = strings.ReplaceAll(htmlTemplate, "$MD_TITLE", title)
	// htmlTemplate = strings.ReplaceAll(htmlTemplate, "$MD_HTML", htmlString)

	return buildTemplate(template, map[string]interface{}{
		"Title":     title,
		"Html":      htmlString,
		"Gitalk":    configGitalk,
		"FastClick": fastClick,
	})
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

func getMeta(file string) *mdMeta {
	if v, ok := metas[file]; ok {
		return v
	}
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		panic(err)
	}
	source, m, err := ParseMarkdownMeta(string(bs))
	if err != nil {
		panic(err)
	}
	if len(m) == 0 {
		return nil
	}
	meta := new(mdMeta)
	meta.Source = source
	meta.Title = m["title"]
	meta.Slug = m["slug"]
	metas[file] = meta
	return meta
}

func buildTemplate(templat string, data interface{}) (string, error) {
	t, err := template2.New("").Parse(templat)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
