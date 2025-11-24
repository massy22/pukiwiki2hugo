
package converter

import (
    "testing"
)

func TestConvertPukiToMd(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
		{
			name:     "空入力",
			input:    "",
			expected: "",
		},
		{
			name:     "シンプルデータ",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "ヘッダー変換",
			input:    "* Header1\n** Header2",
			expected: "# Header1\n## Header2",
		},
		{
			name:     "アンカー削除",
			input:    "*** Header [#anchor]",
			expected: "### Header",
		},
		{
			name:     "リンク変換",
			input:    "[[page name]]",
			expected: "[page name](docs/page-name)",
		},
		{
			name:     "非テーブルチルダ改行",
			input:    "Text with~\nnext line",
			expected: "Text with<br />\nnext line",
		},
  {
            name:     "サイズプラグイン",
            input:    "&size(24){big text}",
            expected: `<span style="font-size:24px;">big text</span>`,
        },
        {
            name:     "テーブルセル先頭の~（ヘッダ指定）は削除（強調と併用）",
            input:    "|~''番号''|値|",
            expected: "|<strong>番号</strong>|値|",
        },
        {
            name:     "テーブルセル先頭の~（ヘッダ指定）は削除（&size 併用）",
            input:    "|~&size(18){Q.};|B|",
            expected: "|<span style=\"font-size:18px;\">Q.</span>|B|",
        },
        {
            name:     "#freeze 行は削除（引数なし）",
            input:    "#freeze\n本文",
            expected: "本文",
        },
        {
            name:     "#freeze 行は削除（引数あり）",
            input:    "#freeze(2025-11-29)\n次の行",
            expected: "次の行",
        },
        {
            name:     "見出し内の~は改行にしない（*）",
            input:    "* タイトル~続き",
            expected: "# タイトル~続き",
        },
        {
            name:     "見出し内の~は改行にしない（***）",
            input:    "*** サブタイトル ~ 続き",
            expected: "### サブタイトル ~ 続き",
        },
		{
			name:     "brタグ",
			input:    "&br;",
			expected: "<br />",
		},
		{
			name: "カウンターインライン",
			input:    "&counter(test)",
			expected: "<!-- counter test -->",
		},
		{
			name: "オンラインユーザ",
			input:    "&online",
			expected: "<!-- online users -->",
		},
  {
            name:     "非テーブルCENTER削除とsize/br置換（強調へ変換）",
            input:    "CENTER:&size(18){''計：90''};&br;",
            expected: "<span style=\"font-size:18px;\"><strong>計：90</strong></span><br />",
        },
		{
			name:     "recentプラグイン削除",
			input:    "#recent(10)\n\nContent",
			expected: "\nContent",
		},
		{
			name:     "テーブル変換",
			input:    "|a|b|c|\n|d|e|f|",
			expected: "|a|b|c|\n|---|---|---|\n|d|e|f|",
		},
		{
			name:     "テーブル内CENTER削除とヘッダh除去",
			input:    "|場所|CENTER:容量|備考|h\n|A|10|note|",
			expected: "|場所|容量|備考|\n|---|---|---|\n|A|10|note|",
		},
  {
            name:     "テーブル内の~はrowspanとして空セル扱い、外側の~は<br />",
            input:    "|a|~|b|\nOutside ~ text",
            expected: "|a||b|\n\nOutside <br /> text",
        },
        {
            name:     "空セルを保持して列数を維持（単一行4列）",
            input:    "|~|40|~||",
            expected: "||40|||",
        },
        {
            name:     "空セルを保持して列数を維持（複数行4列＋セパレータ）",
            input:    "|~|40|~||\n|A|B|C|D|",
            expected: "||40|||\n|---|---|---|---|\n|A|B|C|D|",
        },
        {
            name:     "&new インライン（セミコロン無し）",
            input:    "&new{2008-02-10 (日) 22:00:39}",
            expected: "2008-02-10 (日) 22:00:39",
        },
        {
            name:     "&new インライン（セミコロン有り）",
            input:    "&new{2008-02-10 (日) 22:00:39};",
            expected: "2008-02-10 (日) 22:00:39",
        },
        {
            name:     "&new インライン（文中で使用）",
            input:    "- コメント -- [[管理者]] &new{2008-02-10 (日) 10:31:07};",
            expected: "- コメント -- [管理者](docs/管理者) 2008-02-10 (日) 10:31:07",
        },
		{
			name: "テーブル行末h除去とテール分離",
   input: "|会議室A|10|備考|h<span style=\"font-size:18px\">''計：90''</span>;~",
   expected: "|会議室A|10|備考|\n\n<span style=\"font-size:18px\">''計：90''</span>;<br />",
		},
		{
			name: "テーブル複数行のテールを順に出力",
			input: "|a|b|h<span>A</span>~\n|c|d|<span>B</span>",
			expected: "|a|b|\n|---|---|\n|c|d|\n\n<span>A</span><br />\n<span>B</span>",
		},
		{
			name:     "テーブルの直後が通常行でも空行を挿入",
			input:    "|a|b|\n次の行",
			expected: "|a|b|\n\n次の行",
		},
		{
			name:     "末尾のパイプ後に空白のみならテールを作らない",
			input:    "|a|b|   \n|c|d|",
			expected: "|a|b|\n|---|---|\n|c|d|",
		},
        {
            name:     "アライメント除去後も分離が保たれる",
            input:    "|a|b|\nCENTER:小計",
            expected: "|a|b|\n\n小計",
        },
        {
            name:     "単純なハイフンの箇条書き",
            input:    "- item1\n- item2",
            expected: "- item1\n- item2",
        },
        {
            name:     "ネストしたハイフンの箇条書き",
            input:    "- a\n-- b\n--- c",
            expected: "- a\n  - b\n    - c",
        },
        {
            name:     "リスト前後の空行を保証",
            input:    "冒頭テキスト\n- a\n- b\n末尾テキスト",
            expected: "冒頭テキスト\n\n- a\n- b\n\n末尾テキスト",
        },
        {
            name:     "リスト項目内のインライン置換が有効(&br;/&size)",
            input:    "- &size(18){大}&br;小",
            expected: "- <span style=\"font-size:18px;\">大</span><br />小",
        },
        {
            name:     "リストとテーブルの隣接でも相互干渉しない",
            input:    "- a\n|x|y|\n|1|2|",
            expected: "- a\n\n|x|y|\n|---|---|\n|1|2|",
        },
        {
            name:     "PukiWiki内部リンク（別名+アンカー）",
            input:    "[[参考資料>ガイド#sec1]]",
            expected: "[参考資料](docs/ガイド#sec1)",
        },
        {
            name:     "入れ子ページのリンクは表示テキストを末尾セグメント（親名は含めない）",
            input:    "[[ガイド/第1章～導入]]",
            expected: "[第1章～導入](docs/ガイド/第1章-導入)",
        },
        {
            name:     "PukiWiki外部リンク（別名）",
            input:    "[[公式>http://example.com]]",
            expected: "[公式](http://example.com)",
        },
        {
            name:     "PukiWiki外部リンク（ラベル:URL 形式）",
            input:    "[[ニュース記事:http://example.com/news]]",
            expected: "[ニュース記事](http://example.com/news)",
        },
        {
            name:     "PukiWiki外部リンク（行頭ハイフン付き・後続テキスト付き）",
            input:    "-[[ニュース記事:http://example.com/news]]（配信元）",
            expected: "- [ニュース記事](http://example.com/news)（配信元）",
        },
        {
            name:     "PukiWiki内部リンク（同名）",
            input:    "[[使い方]]",
            expected: "[使い方](docs/使い方)",
        },
        {
            name:     "リスト継続行のレンダリング",
            input:    "- aaaa\n  bbbb\n  - cccc",
            expected: "- aaaa<br />\n  bbbb\n  - cccc",
        },
      		{
			name:     "ハイフン直後に空白なし＋末尾~で継続（強調へ変換）",
			input:    "-対戦する当人同士しか対戦の様子を確認できませんので、原則的に''自己申告制''です。~\nお互いにルールを守って楽しい対戦会にしましょう！",
			expected: "- 対戦する当人同士しか対戦の様子を確認できませんので、原則的に<strong>自己申告制</strong>です。<br />\n  お互いにルールを守って楽しい対戦会にしましょう！",
		},
		{
			name:     "インライン強調と斜体の基本",
			input:    "''bold'' と '''italic'''",
			expected: "<strong>bold</strong> と <em>italic</em>",
		},
		{
			name:     "インライン入れ子（斜体の中に強調）",
			input:    "'''a ''b'' c'''",
			expected: "<em>a <strong>b</strong> c</em>",
		},
		{
			name:     "番号付きリスト（+）の基本",
			input:    "+ a\n++ b",
			expected: "1. a\n  1. b",
		},
		{
			name:     "番号付きリスト +~ で段落継続",
			input:    "+~ 見出し\n 継続行",
			expected: "1. 見出し<br />\n  継続行",
		},
		{
			name:     "引用の正規化（スペース付与）",
			input:    ">引用\n>>多段",
			expected: "> 引用\n>> 多段",
		},
	}


    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ConvertPukiToMd(tt.input)
            if result != tt.expected {
                t.Errorf("ConvertPukiToMd(%q) = %q; want %q", tt.input, result, tt.expected)
            }
        })
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
		{"スラッシュ", "a/b/c", "a/b/c"},
		{"数字", "page123", "page123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slugify(tt.input)
			if result != tt.expected {
				t.Errorf("slugify(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		index    int
		value    string
		expected []string
	}{
		{
			name:     "挿入中間",
			slice:    []string{"a", "b", "c"},
			index:    1,
			value:    "x",
			expected: []string{"a", "x", "b", "c"},
		},
		{
			name:     "挿入先頭",
			slice:    []string{"a", "b"},
			index:    0,
			value:    "x",
			expected: []string{"x", "a", "b"},
		},
		{
			name:     "挿入末尾",
			slice:    []string{"a", "b"},
			index:    2,
			value:    "x",
			expected: []string{"a", "b", "x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := insert(tt.slice, tt.index, tt.value)
			if len(result) != len(tt.expected) {
				t.Fatalf("insert() length = %d; want %d", len(result), len(tt.expected))
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("insert()[%d] = %q; want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestCleanTableLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "簡単",
			input:    "|a|b|",
			expected: "|a|b|",
		},
		{
			name:     "アライメント",
			input:    "|LEFT:a|CENTER:b|RIGHT:c|",
			expected: "|a|b|c|",
		},
		{
			name:     ">削除",
			input:    "|>|a|b|",
			expected: "||a|b|",
		},
		{
			name:     "テーブルセル内の~はrowspanとして空セル",
			input:    "|~|a|b|",
			expected: "||a|b|",
		},
		{
			name:     "行末のh削除",
			input:    "|a|b|h",
			expected: "|a|b|",
		},
		{
			name:     "閉じるパイプなし",
			input:    "|a|b",
			expected: "|a|",
		},
		{
			name:     "HTMLタグの '>' は保持",
			input:    "|<span style=\"font-size:18px\">X</span>|",
			expected: "|<span style=\"font-size:18px\">X</span>|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanTableLine(tt.input)
			if result != tt.expected {
				t.Errorf("cleanTableLine(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertTables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "簡単テーブル",
			input:    "|a|b|\n|c|d|",
			expected: "|a|b|\n|---|---|\n|c|d|",
		},
		{
			name:     "単一行",
			input:    "|a|b|",
			expected: "|a|b|",
		},
		{
			name:     "複数テーブル",
			input:    "|a|b|\n\n|x|y|",
			expected: "|a|b|\n\n|x|y|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertTables(tt.input)
			if result != tt.expected {
				t.Errorf("convertTables(%q) =\n%q\nwant:\n%q", tt.input, result, tt.expected)
			}
		})
	}
}
