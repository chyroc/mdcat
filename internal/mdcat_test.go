package internal

import (
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
