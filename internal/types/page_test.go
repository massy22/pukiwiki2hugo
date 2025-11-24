package types

import (
	"testing"
	"time"
)

func TestNewPage(t *testing.T) {
	name := "Test Page"
	content := "Some content"
	date := time.Now()

	page := NewPage(name, content, date)

	if page.Name != name {
		t.Errorf("page.Name = %q; want %q", page.Name, name)
	}
	if page.Content != content {
		t.Errorf("page.Content = %q; want %q", page.Content, content)
	}
	if !page.Date.Equal(date) {
		t.Errorf("page.Date = %v; want %v", page.Date, date)
	}
	expectedSlug := Slugify(name)
	if page.Slug != expectedSlug {
		t.Errorf("page.Slug = %q; want %q", page.Slug, expectedSlug)
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"空", "", ""},
		{"簡潔", "hello", "hello"},
		{"スペース", "hello world", "hello-world"},
		{"特殊文字", "hello/world", "hello/world"},
		{"日本語", "テストページ", "テストページ"},
		{"混合", "test page テスト", "test-page-テスト"},
		{"数字", "page123", "page123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}


}
