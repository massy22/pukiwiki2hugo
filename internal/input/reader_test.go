package input

import (
    "encoding/hex"
    "os"
    "path/filepath"
    "testing"

    "github.com/massy22/pukiwki2hugo/internal/types"
)

func TestDecodePageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasErr   bool
	}{
		{
			name:     "シンプル",
			input:    "48656c6c6f", // "Hello" hex
			expected: "Hello",
			hasErr:   false,
		},
		{
			name:     "空",
			input:    "",
			expected: "",
			hasErr:   false,
		},
		{
			name:     "日本語",
			input:    "e38386e382b9e38388", // テスト hex
			expected: "テスト",
			hasErr:   false,
		},
		{
			name:     "無効",
			input:    "invalid",
			expected: "",
			hasErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodePageName(tt.input)
			if (err != nil) != tt.hasErr {
				t.Errorf("decodePageName(%q) error = %v; hasErr %v", tt.input, err, tt.hasErr)
				return
			}
			if result != tt.expected {
				t.Errorf("decodePageName(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetDefaultPage(t *testing.T) {
    dir := t.TempDir()
    // 最低限の pukiwiki.ini.php を生成
    ini := []byte("$defaultpage = 'ホーム';\n")
    if err := os.WriteFile(filepath.Join(dir, "pukiwiki.ini.php"), ini, 0644); err != nil {
        t.Fatalf("failed to write ini: %v", err)
    }
    result, err := GetDefaultPage(dir)
    if err != nil {
        t.Fatalf("GetDefaultPage error: %v", err)
    }
    expected := "ホーム"
    if result != expected {
        t.Errorf("GetDefaultPage() = %q; want %q", result, expected)
    }
}

func TestReadPages(t *testing.T) {
    dir := t.TempDir()
    wikiDir := filepath.Join(dir, "wiki")
    if err := os.MkdirAll(wikiDir, 0755); err != nil {
        t.Fatalf("mkdir wiki: %v", err)
    }
    // ページ名をUTF-8でhex化したファイルを作成
    name := "ガイド/第1章～導入"
    hexName := hex.EncodeToString([]byte(name)) + ".txt"
    content := []byte("内容です")
    if err := os.WriteFile(filepath.Join(wikiDir, hexName), content, 0644); err != nil {
        t.Fatalf("write page: %v", err)
    }

    pages, err := ReadPages(dir)
    if err != nil {
        t.Fatalf("ReadPages error: %v", err)
    }
    if len(pages) != 1 {
        t.Fatalf("expected 1 page, got %d", len(pages))
    }
    page := pages[0]
    if page.Name != name {
        t.Errorf("Name = %q; want %q", page.Name, name)
    }
    if page.Slug != types.Slugify(page.Name) {
        t.Errorf("Slug = %q; want %q", page.Slug, types.Slugify(page.Name))
    }
    if page.Content != string(content) {
        t.Errorf("Content = %q; want %q", page.Content, string(content))
    }
    if page.Date.IsZero() {
        t.Error("Date is zero")
    }
}

// typesパッケージのslugifyを使うため、importするが、このパッケージなので直接呼び
// 注意: Slugify の検証は converter 側で行うため、ここでは types.Slugify を参照して一致性のみを確認します。
