package types

import (
	"strings"
	"time"
	"unicode"
)

type Page struct {
	Name    string
	Slug    string
	Content string
	Date    time.Time
}

func NewPage(name, content string, date time.Time) *Page {
	return &Page{
		Name:    name,
		Slug:    Slugify(name),
		Content: content,
		Date:    date,
	}
}

// Slugify はページ名からURL-safeなスラッグを作成します
func Slugify(name string) string {
	name = strings.Map(func(r rune) rune {
		if r == '/' || unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		} else if unicode.IsSpace(r) {
			return '-'
		}
		return '-'
	}, name)
	return name
}
