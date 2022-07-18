package mdcat

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/chyroc/go-lambda"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Link             bool          `yaml:"link"`
	Title            string        `yaml:"title"`
	Output           string        `yaml:"output"`
	Gitalk           *ConfigGitalk `yaml:"gitalk"`
	FastClick        bool          `yaml:"fast_click"`
	IsOmitHtmlSuffix bool          `yaml:"omit_html_suffix"`
}

type ConfigGitalk struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	Repo         string   `yaml:"repo"`
	Owner        string   `yaml:"owner"`
	Admin        []string `yaml:"admin"`
	ID           string   `yaml:"id"`
	Labels       []string `yaml:"labels"`

	Slug        string
	AdminLitel  string
	LabelsLitel string
}

func ParseConfig(file string, title, output string, link bool) (*Config, error) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Link:             link,
				Title:            title,
				Output:           output,
				Gitalk:           nil,
				IsOmitHtmlSuffix: false,
			}, nil
		}
		return nil, err
	}
	t := new(Config)
	err = yaml.Unmarshal(bs, &t)
	if err != nil {
		return nil, err
	}

	t.Link = t.Link || link
	if output != "" {
		t.Output = output
	}
	if title != "" {
		t.Title = title
	}

	f := func(l []string) (string, error) {
		s, err := lambda.New(l).MapList(func(idx int, obj interface{}) interface{} {
			return fmt.Sprintf("'%s'", obj.(string))
		}).ToJoin(",")
		if err != nil {
			return "", err
		}
		return "[" + s + "]", nil
	}

	if t.Gitalk != nil {
		t.Gitalk.AdminLitel, err = f(t.Gitalk.Admin)
		if err != nil {
			return nil, err
		}
		t.Gitalk.LabelsLitel, err = f(t.Gitalk.Labels)
		if err != nil {
			return nil, err
		}
	}

	conf = t

	return t, nil
}

func (r *ConfigGitalk) clone(slug string) *ConfigGitalk {
	if r == nil {
		return r
	}
	return &ConfigGitalk{
		ClientID:     r.ClientID,
		ClientSecret: r.ClientSecret,
		Repo:         r.Repo,
		Owner:        r.Owner,
		Admin:        r.Admin,
		ID:           r.ID,
		Labels:       r.Labels,
		AdminLitel:   r.AdminLitel,
		LabelsLitel:  r.LabelsLitel,

		Slug: slug,
	}
}
