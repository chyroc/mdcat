package internal

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	as := assert.New(t)

	for i := 0; i < 2; i++ {
		t.Run("", func(t *testing.T) {
			res, err := convertWithGitHubApi("title", `# hi

## you are great

- list1
- list2
`)
			as.Nil(err)
			as.Contains(res, "<title>title</title>")
			as.Contains(res, "you are great</h2>")
			as.Contains(res, "hi</h1>")
			as.Contains(res, "<li>list1</li>")
			as.Contains(res, "<li>list2</li>")
		})
	}
}

func Test_Rel(t *testing.T) {
	as := assert.New(t)

	runWithFunc := func() {
		as.Nil(Run("testdata/1.md", "Hi", true, "dist/index.html"))
	}

	runWithGo := func() {
		cmd := exec.Command(`go`, `run`, `main.go`, `--link`, `--output`, `dist/index.html`, `--title`, `Hi`, `testdata/1.md`)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		as.Nil(cmd.Run())
	}

	for _, f := range []func(){runWithFunc, runWithGo} {
		t.Run("", func(t *testing.T) {
			pwd, err := os.Getwd()
			as.Nil(err)
			if strings.HasSuffix(pwd, "/internal") {
				as.Nil(os.Chdir(".."))
			}

			_ = os.RemoveAll("dist")
			defer os.RemoveAll("dist")

			f()

			assertFileContain(t, "dist/index.html", []string{
				`<title>Hi</title>`,
				`<li><a href="./2.html">url2</a></li>`,
				`<li><a href="./3.html">url3</a></li>`,
				`<li><a href="./4/4.html">url4</a></li>`,
			})

			assertFileContain(t, "dist/2.html", []string{
				`<title>2.md</title>`,
				`<p><a href="./3.html">url3</a></p>`,
			})

			assertFileContain(t, "dist/3.html", []string{
				`<title>3.md</title>`,
			})

			assertFileContain(t, "dist/4/4.html", []string{
				`<title>4.md</title>`,
				`<p><a href="../index.html">rev_url</a></p>`,
			})
		})
	}
}

func assertFileContain(t *testing.T, file string, contains []string) {
	as := assert.New(t)

	bs, err := ioutil.ReadFile(file)
	as.Nil(err)

	for _, v := range contains {
		as.Contains(string(bs), v)
	}
}
