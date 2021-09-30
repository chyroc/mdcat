package mdcat

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Parse(t *testing.T) {
	as := assert.New(t)

	leftText, meta, err := ParseMarkdownMeta(`-
Title: goldmark-meta
Summary: Add YAML metadata to the document
-
# Hello goldmark-meta
`,
	)

	as.Nil(err)
	as.Equal(`# Hello goldmark-meta`+"\n", leftText)
	as.Equal(map[string]string{
		"Summary": "Add", "Title": "goldmark-meta",
	}, meta)
}

func Test_Meta(t *testing.T) {
	as := assert.New(t)

	tests := []struct {
		name       string
		args       string
		k          string
		v          string
		errContain string
	}{
		{name: "1", args: "a:b", k: "a", v: "b"},
		{name: "1", args: "a: b", k: "a", v: "b"},
		{name: "1", args: "a : b", k: "a", v: "b"},
		{name: "1", args: "a : b ", k: "a", v: "b"},
		{name: "1", args: " a : b ", k: "a", v: "b"},
		{name: "1", args: ` a :" b "`, k: "a", v: " b "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, v, err := parseMeta(tt.args)
			if tt.errContain != "" {
				as.NotNil(err, fmt.Sprintf("%s, got={%v,%v}", tt.name, k, v))
				as.Contains(err.Error(), tt.errContain, fmt.Sprintf("%s, got={%v,%v}", tt.name, k, v))
				return
			}

			as.Nil(err, tt.name)
			as.Equal(tt.k, k)
			as.Equal(tt.v, v)
		})
	}
}
