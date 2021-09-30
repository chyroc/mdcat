package mdcat

import (
	"fmt"
	"strings"
)

func ParseMarkdownMeta(source string) (string, map[string]string, error) {
	lines := strings.Split(source, "\n")
	startMeta := false // meta 区
	// startPreText := false // 正文区，注意，这里默认是 false，即么；进入 meta 区会刷新这个值
	texts := []string{}
	meta := map[string]string{}
	for i := 0; i < len(lines); i++ {
		v := lines[i]

		if i == 0 {
			// -- 必须在第一行
			if onlyContain(v, '-') {
				startMeta = true
				// startText = false
				// 必然不是 正文
			} else {
				// startMeta = false
				// startText = true
				texts = append(texts, v)
			}
		} else {
			if startMeta {
				// 收集 meta 信息，知道再次遇到 --
				if onlyContain(v, '-') {
					startMeta = false
					// startText = true
					// 必然不是 正文
				} else {
					// meta 信息
					// 必然不是 正文
					fmt.Println("meta:", v)
					metak, metav, err := parseMeta(v)
					if err != nil {
						return "", nil, err
					}
					meta[metak] = metav
				}
			} else {
				// 必然不是 meta
				texts = append(texts, v)
			}
		}
	}
	return strings.Join(texts, "\n"), meta, nil
}

func parseMeta(s string) (string, string, error) {
	r := newFindkv(s)
	return r.parse()
}

func newFindkv(s string) *findkv {
	return &findkv{idx: 0, runes: []rune(s)}
}

type findkv struct {
	idx   int
	runes []rune
}

func (r *findkv) parse() (string, string, error) {
	key, err := r.findLiter()
	if err != nil {
		return "", "", err
	}
	err = r.nextChat(':')
	if err != nil {
		return "", "", err
	}
	val, err := r.findLiter()
	if err != nil {
		return "", "", err
	}
	return key, val, nil
}

func (r *findkv) findLiter() (string, error) {
	state := 0
	values := []int32{}
	for r.idx < len(r.runes) {
		v := r.runes[r.idx]
		switch state {
		case 0:
			if v == ' ' {
				r.idx++
				continue
			}
			if v == '"' {
				state = 2
				r.idx++
				continue
			}
			state = 1
			values = append(values, v)
			r.idx++
		case 1:
			if v == ' ' {
				r.idx++
				return string(values), nil
			}
			if v == '"' {
				return "", fmt.Errorf("不合法的meta: %q", r.runes)
			}
			if v == ':' {
				return string(values), nil
			}
			values = append(values, v)
			r.idx++
		case 2:
			if v == '"' {
				r.idx++
				return string(values), nil
			}
			values = append(values, v)
			r.idx++
		}
	}
	return string(values), nil
}

func (r *findkv) nextChat(x rune) error {
	for r.idx < len(r.runes) {
		v := r.runes[r.idx]
		if v == x {
			r.idx++
			return nil
		}
		if v == ' ' {
			r.idx++
			continue
		}
		return fmt.Errorf("不合法的meta: %q", r.runes)
	}
	return nil
}

func onlyContain(s string, b rune) bool {
	for _, v := range s {
		if v != b {
			return false
		}
	}
	return true
}
