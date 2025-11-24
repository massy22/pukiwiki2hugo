package converter

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/massy22/pukiwki2hugo/internal/types"
)

// 事前コンパイル済みの正規表現（性能・可読性の向上）
var (
	reAlignStrip = regexp.MustCompile(`(?m)(^|\|)\s*(LEFT|CENTER|RIGHT):`)
	reCellAlign  = regexp.MustCompile(`(LEFT|CENTER|RIGHT|NEXT):`)
	reLeadingGt  = regexp.MustCompile(`^>+`)
	// リンク・強調・見出しなどの事前コンパイル済み正規表現
	reLinkAll       = regexp.MustCompile(`\[\[([^]]+)]]`)
	reLabelURL      = regexp.MustCompile(`^(.*?):\s*(https?://\S+|mailto:\S+)$`)
	reItalic        = regexp.MustCompile(`'''(.*?)'''`)
	reBold          = regexp.MustCompile(`''(.*?)''`)
	reHeaderLine    = regexp.MustCompile(`(?m)^(\*+)\s*(.+)$`)
	reHeadingAnchor = regexp.MustCompile(` ?\[#[^]]+]`)
	reRecent        = regexp.MustCompile(`#recent\(([0-9]+)\)\s*\n?`)
	reSize          = regexp.MustCompile(`&size\((\d+)\)\{([^}]*)}\s*;?`)
	reColor         = regexp.MustCompile(`&color\(([^)]+)\)\{([^}]*)}\s*;?`)
	reBR            = regexp.MustCompile(`&br;`)
	reCounter       = regexp.MustCompile(`&counter\(([^)]+)\)`)
	reOnline        = regexp.MustCompile(`&online`)
)

// cleanTableTail はテーブル行末の tail 文字列をクリーンアップして返します。
// 仕様:
// - "~" は "<br />" に置換
// - 単独の "h" はヘッダー指定として無視（空文字を返す）
// - 先頭が 'h' で、直後が英数字でない場合は先頭の 'h' を除去
// - 前後空白はトリム
func cleanTableTail(tail string) string {
	t := strings.TrimSpace(tail)
	if t == "" {
		return ""
	}
	// ~ → <br />
	t = strings.ReplaceAll(t, "~", "<br />")
	// テーブル行末 tail 内では、インライン強調/斜体の自動変換を抑止するため、
	// 既に変換済みの <strong>/<em> を PukiWiki の ''/''' に戻す
	// （他の箇所の強調は保持される）。
	reStrong := regexp.MustCompile(`(?i)<strong>(.*?)</strong>`) // 非貪欲
	reEm := regexp.MustCompile(`(?i)<em>(.*?)</em>`)
	t = reStrong.ReplaceAllString(t, `''$1''`)
	t = reEm.ReplaceAllString(t, `'''$1'''`)
	// 単独の h は無視
	if t == "h" {
		return ""
	}
	if strings.HasPrefix(t, "h") {
		r := []rune(t)
		if len(r) > 1 {
			if !unicode.IsLetter(r[1]) && !unicode.IsNumber(r[1]) {
				t = strings.TrimSpace(string(r[1:]))
			}
		}
	}
	return strings.TrimSpace(t)
}

// ConvertPukiToMd は PukiWiki 構文を Markdown に変換します
func ConvertPukiToMd(content string) string {

	// #author(...) はブロック要素。行ごと（改行も含めて）削除
	content = regexp.MustCompile(`(?m)^\s*#author\([^\n]*\)\s*(\r?\n)?`).ReplaceAllString(content, "")

	// #freeze もブロック要素。#author 同様に行ごと削除（オプション引数があっても削除）
	content = regexp.MustCompile(`(?m)^\s*#freeze(?:\([^\n)]*\))?\s*(\r?\n)?`).ReplaceAllString(content, "")

	content = reHeaderLine.ReplaceAllStringFunc(content, func(match string) string {
		parts := regexp.MustCompile(`^(\*+)\s*(.+)$`).FindStringSubmatch(match)
		stars := len(parts[1])
		// Remove anchor like [#l01bc1e0]
		heading := reHeadingAnchor.ReplaceAllString(parts[2], "")
		// Hugo はマークダウンと同じく # をヘッダーに使用
		return strings.Repeat("#", stars) + " " + strings.TrimSpace(heading)
	})

	// PukiWiki リンクを Markdown へ変換
	content = convertLinks(content)

	// 非テーブル行に残るアライメント指定子は後段で削除する（テーブル内はcleanTableLineで処理）

	// 最近のブロックプラグインを置換（後続の空行は1つの改行に正規化）
	content = reRecent.ReplaceAllString(content, "\n")

	// &new{...} はインライン型プラグインだが、今回の要件では中身をそのまま出力する
	// 例: &new{2008-02-10 (日) 22:00:39}; → 2008-02-10 (日) 22:00:39
	content = regexp.MustCompile(`&new\{([^}]*)}\s*;?`).ReplaceAllString(content, `$1`)

	// カウンターインラインを置換
	content = reCounter.ReplaceAllString(content, `<!-- counter $1 -->`)

	// オンラインプラグインを置換
	content = reOnline.ReplaceAllString(content, `<!-- online users -->`)
	// &br; を <br /> に置換
	content = reBR.ReplaceAllString(content, "<br />")
	// &size(n){text} を <span style="font-size:npx;">text</span> に置換
	// 末尾のセミコロンは出力しない（オプション扱い）
	content = reSize.ReplaceAllStringFunc(content, func(m string) string {
		sm := reSize.FindStringSubmatch(m)
		if len(sm) >= 3 {
			return `<span style="font-size:` + sm[1] + `px;">` + sm[2] + `</span>`
		}
		return m
	})

	// &color(name){text} を <span style="color:name;">text</span> に置換
	// 末尾のセミコロンは出力しない（オプション扱い）
	content = reColor.ReplaceAllStringFunc(content, func(m string) string {
		sm := reColor.FindStringSubmatch(m)
		if len(sm) >= 3 {
			return `<span style="color:` + sm[1] + `;">` + sm[2] + `</span>`
		}
		return m
	})

	// インライン強調/斜体を変換（'''...'''→<em>、''...''→<strong>）
	content = convertInlineEmphasis(content)

	// 番号付きリスト（+）をMarkdownへ変換
	content = convertOrderedLists(content)

	// 箇条書き（- のリスト）をMarkdownへ変換
	content = convertLists(content)

	// 引用（>、>>、>>>）の正規化（空白調整とブロックの明確化）
	content = convertBlockquotes(content)

	// テーブルを変換
	content = convertTables(content)

	// テーブル行の末尾に余分なテキストがぶら下がっている場合、
	// 行末の '|' 以降を分離し、テーブルブロックの直後に空行を挟んで独立行として配置する
	content = enforceTableRowTailSeparation(content)

	// 変換後に残ったアライメント指定子を削除
	// 注意: 以前の `\s*(...)` だと直前の改行も巻き込んで消えてしまい、行が結合される不具合があった。
	// 行頭(^)またはテーブル区切りの直後(\|)に限って削除し、前置文字は保持する。
	content = reAlignStrip.ReplaceAllString(content, "$1")

	// アライメント除去の副作用などでテーブル行末にテキストが結合された場合に備え、最終的にもう一度分離を保証
	content = enforceTableRowTailSeparation(content)

	return content
}

// convertLists は PukiWiki の '-' 箇条書きを Markdown の箇条書きに変換します。
// 仕様:
// - 先頭の '-' の個数がネストレベル。Markdown では (レベル-1)*2 スペースでインデントし、'- ' を付与
// - 行頭が '-' のみ、またはテキストが空の行はスキップ
// - リストブロックの前後には空行を1行だけ挿入（既に空行なら追加しない）
// - 表行('|')や見出し('*')等は対象外
func convertLists(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	inList := false
	// 直前のリスト項目が継続行を許可する（末尾が <br />）場合に true
	continueBlock := false
	currentLevel := 1
	// 先頭の空白も取得して、先頭空白/2 をネスト加算として扱う（Markdown由来の入力も許容）
	reList := regexp.MustCompile(`^(\s*)(-+)\s*(.+)$`)

	flushBreakBefore := func() {
		if len(out) > 0 && out[len(out)-1] != "" {
			out = append(out, "")
		}
	}

	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		trimmed := strings.TrimSpace(raw)

		// テーブルや見出し、空行の判定
		if trimmed == "" {
			// 空行はそのまま出力し、リストブロックを終了
			if inList {
				inList = false
			}
			out = append(out, "")
			continue
		}
		if strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, "*") {
			if inList {
				// リスト終了時、直前が空行でなければ空行を挿入
				if len(out) > 0 && out[len(out)-1] != "" {
					out = append(out, "")
				}
				inList = false
				continueBlock = false
			}
			out = append(out, raw)
			continue
		}

		// 引用プレフィックスと本文を分離（リストは引用の子要素になれるため）
		quotePrefix, rest := parseQuotePrefix(raw)

		if m := reList.FindStringSubmatch(rest); m != nil {
			// リスト項目
			leading := m[1]
			hyphens := m[2]
			level := len(hyphens) + (len(leading) / 2)
			text := strings.TrimSpace(m[3])
			// PukiWiki の行末 '~' は強制改行を意味するため、ここで <br /> に変換し、
			// さらに次行を継続行として扱うトリガーにする。
			forcedContinue := false
			if strings.HasSuffix(text, "~") {
				text = strings.TrimRight(strings.TrimSuffix(text, "~"), " \t") + "<br />"
				forcedContinue = true
			}
			if text == "" {
				// 空要素は無視（PukiWiki では意図せぬ '-' 単独を落とす）
				continue
			}
			if !inList {
				// リスト開始前に空行で分離
				flushBreakBefore()
				inList = true
			}
			indent := strings.Repeat("  ", level-1)
			out = append(out, quotePrefix+indent+"- "+text)
			currentLevel = level
			// 末尾が <br /> の場合は継続行を許可
			continueBlock = forcedContinue || strings.HasSuffix(text, "<br />") || strings.HasSuffix(text, "<br/>")

			// 次の行が非リストであれば、ここでブロックを閉じて空行を挿入。
			if i+1 < len(lines) {
				nextRaw := lines[i+1]
				nextTrim := strings.TrimSpace(nextRaw)
				// 次行が継続行（先頭に空白があり、非リスト・非テーブル）の場合は継続扱い
				if !continueBlock {
					if (strings.HasPrefix(nextRaw, " ") || strings.HasPrefix(nextRaw, "\t")) && nextTrim != "" && !strings.HasPrefix(nextTrim, "-") && !strings.HasPrefix(nextTrim, "|") && !strings.HasPrefix(nextTrim, "*") {
						continueBlock = true
					}
				}
				if !continueBlock {
					if nextTrim == "" {
						// 次が空行なのでリストを閉じる
						out = append(out, "")
						inList = false
					} else if !strings.HasPrefix(nextTrim, "-") && !strings.HasPrefix(nextTrim, "|") {
						// 非リスト（テーブルや見出し等は convertLists 内では扱わないが、分離のため空行）
						out = append(out, "")
						inList = false
					}
				}
			}
			continue
		}

		// 通常行
		if inList {
			// リスト項目の継続行として扱う。
			// 先頭の項目行末に <br /> を確実に付ける（既にあれば付けない）。
			if len(out) > 0 {
				last := out[len(out)-1]
				if !strings.HasSuffix(last, "<br />") && !strings.HasSuffix(last, "<br/>") {
					out[len(out)-1] = last + "<br />"
				}
			}
			indent := strings.Repeat("  ", currentLevel)
			// 引用の中での継続行に対応
			qp, rest2 := parseQuotePrefix(raw)
			// 直前項目と同じ引用レベルが理想だが、ここでは入力の '>' を尊重
			out = append(out, qp+indent+strings.TrimSpace(rest2))
			continueBlock = true
			continue
		}
		// リスト外の通常行はそのまま
		out = append(out, raw)
	}
	return strings.Join(out, "\n")
}

// convertLinks は PukiWiki のリンク表記を Markdown のリンクに変換します。
// 例:
//   - [[ページ名]]                  => [ページ名](docs/ページ名)
//   - [[ラベル>ページ名#anchor]]   => [ラベル](docs/ページ名#anchor)
//   - [[公式>http://ex.com]]       => [公式](http://ex.com)
//   - [[ページ名#a1]]              => [ページ名](docs/ページ名#a1)
func convertLinks(content string) string {
	return reLinkAll.ReplaceAllStringFunc(content, func(m string) string {
		inner := reLinkAll.FindStringSubmatch(m)[1]
		label, target, hadAlias := splitAlias(inner)

		// 別名未使用なら、ラベル:URL 形式の外部リンクを試す
		if !hadAlias {
			if m2 := reLabelURL.FindStringSubmatch(inner); len(m2) == 3 {
				label = strings.TrimSpace(m2[1])
				target = strings.TrimSpace(m2[2])
			}
		}

		base, anchor := splitAnchor(target)
		if isExternalURL(base) {
			return "[" + label + "](" + base + anchor + ")"
		}

		// 内部ページ: 別名なしの場合は末尾セグメントをラベルに使う（テキスト自体の正規化はしない）
		if !hadAlias {
			label = lastSegment(base)
		}
		url := buildInternalURL(base, anchor)
		return "[" + label + "](" + url + ")"
	})
}

// splitAlias は PukiWiki の [[label>target]] 形式を分解する。
// '>' が含まれない場合は (inner, inner, false) を返す。
func splitAlias(inner string) (label, target string, hadAlias bool) {
	if idx := strings.Index(inner, ">"); idx >= 0 {
		return strings.TrimSpace(inner[:idx]), strings.TrimSpace(inner[idx+1:]), true
	}
	return inner, inner, false
}

// splitAnchor は target の中からアンカー部 (#...) を切り出す。
// 返り値は (ベース, アンカー) とし、アンカーには先頭の '#' を含む。
func splitAnchor(target string) (base, anchor string) {
	if i := strings.Index(target, "#"); i >= 0 {
		return target[:i], target[i:]
	}
	return target, ""
}

// isExternalURL は http(s):// または mailto: で始まるかを判定する。
func isExternalURL(s string) bool {
	ls := strings.ToLower(s)
	return strings.HasPrefix(ls, "http://") || strings.HasPrefix(ls, "https://") || strings.HasPrefix(ls, "mailto:")
}

// lastSegment はパスの末尾セグメント（最後の '/' の後）を返す。
func lastSegment(path string) string {
	if slash := strings.LastIndex(path, "/"); slash >= 0 {
		return path[slash+1:]
	}
	return path
}

// buildInternalURL は内部ページの URL を生成する。
// 仕様: "docs/" + slugify(base) + anchor
func buildInternalURL(base, anchor string) string {
	return "docs/" + slugify(base) + anchor
}

// convertInlineEmphasis は PukiWiki の強調/斜体を HTML タグに変換します。
// 仕様:
// - 斜体: ”'text”' → <em>text</em>
// - 強調: ”text”   → <strong>text</strong>
// 入れ子を考慮し、先に斜体、次に強調を置換します。
func convertInlineEmphasis(content string) string {
	// 斜体（3アポストロフィ）: 内部に '' を含める入れ子を許容するため非貪欲にマッチ
	content = reItalic.ReplaceAllString(content, `<em>$1</em>`)
	// 強調（2アポストロフィ）: 非貪欲にマッチ
	content = reBold.ReplaceAllString(content, `<strong>$1</strong>`)
	return content
}

// parseQuotePrefix は行頭の '>' 連続（最大3想定）を検出し、
// 引用レベルの '>' プレフィックス文字列（末尾スペース付き）と残りの本文を返します。
func parseQuotePrefix(raw string) (prefix string, rest string) {
	s := strings.TrimLeft(raw, " ")
	// 数える
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '>' {
			count++
		} else {
			break
		}
	}
	if count == 0 {
		return "", raw
	}
	// '>' の後の空白を1つスキップ
	tail := strings.TrimLeft(s[count:], " \t")
	// もとの raw とのインデント差異は無視して正規化
	return strings.Repeat(">", count) + " ", tail
}

// convertOrderedLists は PukiWiki の '+' 番号付きリストを Markdown の番号付きリストに変換します。
// ネストは + の個数で判断し、(level-1)*2 スペース + "1. " を付与します。
// 項目テキストが '~' で終わる、または "+~" で開始する場合は末尾に <br /> を付与し、
// 次行のインデント行を継続行として取り込みます。
func convertOrderedLists(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	inList := false
	continueBlock := false
	currentLevel := 1

	// 先頭の + を検出（> プレフィックスを許容）
	rePlus := regexp.MustCompile(`^(\s*)(\+{1,3})(~?)(\s*)(.*)$`)

	flushBreakBefore := func() {
		if len(out) > 0 && out[len(out)-1] != "" {
			out = append(out, "")
		}
	}

	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		trimmed := strings.TrimSpace(raw)

		if trimmed == "" {
			if inList {
				inList = false
				continueBlock = false
			}
			out = append(out, "")
			continue
		}

		// 引用プレフィックスと本文を分離
		quotePrefix, rest := parseQuotePrefix(raw)

		if m := rePlus.FindStringSubmatch(rest); m != nil {
			// レベルと本文抽出
			leadingSpaces := m[1]
			pluses := m[2]
			hasTilde := m[3] == "~"
			afterTildeSpace := m[4]
			text := strings.TrimSpace(m[5])
			level := len(pluses) + (len(leadingSpaces) / 2)

			if text == "" {
				// 空要素は無視
				continue
			}
			if !inList {
				flushBreakBefore()
				inList = true
			}
			// +~ の場合は段落開始を意味するため、明示的に <br /> 付与
			if hasTilde {
				if text == "" && afterTildeSpace != "" {
					text = strings.TrimSpace(afterTildeSpace)
				}
				if !strings.HasSuffix(text, "<br />") && !strings.HasSuffix(text, "<br/>") {
					text = strings.TrimSpace(text) + "<br />"
				}
			}
			indent := strings.Repeat("  ", level-1)
			out = append(out, quotePrefix+indent+"1. "+text)
			currentLevel = level
			continueBlock = hasTilde || strings.HasSuffix(text, "<br />") || strings.HasSuffix(text, "<br/>")

			// 次行が非リストなら分離（convertLists 同様の簡易処理）
			if i+1 < len(lines) && !continueBlock {
				nt := strings.TrimSpace(lines[i+1])
				if nt == "" || (!strings.HasPrefix(strings.TrimLeft(nt, "> "), "+")) {
					out = append(out, "")
					inList = false
				}
			}
			continue
		}

		// 継続行（インデントがあり、非+ 行）
		if inList {
			// 直前の項目末尾に <br /> を付与
			if len(out) > 0 {
				last := out[len(out)-1]
				if !strings.HasSuffix(last, "<br />") && !strings.HasSuffix(last, "<br/>") {
					out[len(out)-1] = last + "<br />"
				}
			}
			indent := strings.Repeat("  ", currentLevel)
			out = append(out, quotePrefix+indent+strings.TrimSpace(rest))
			continue
		}

		// 非リスト行
		out = append(out, raw)
	}

	return strings.Join(out, "\n")
}

// convertBlockquotes は PukiWiki/Markdown いずれにも存在する引用 '>' を正規化します。
// 連続 '>' の後を1スペースに正規化し、空行でブロックを終了します。
func convertBlockquotes(content string) string {
	lines := strings.Split(content, "\n")
	for i, raw := range lines {
		t := strings.TrimLeft(raw, " ")
		// 引用の正規化（'>' 連続の後にスペースを1つ）
		if strings.HasPrefix(t, ">") {
			count := 0
			for count < len(t) && t[count] == '>' {
				count++
			}
			rest := strings.TrimLeft(t[count:], " \t")
			lines[i] = strings.Repeat(">", count) + " " + rest
			continue
		}
		// 脱出記号 '<', '<<', '<<<' は出力上は何も出さず、その行は空行として扱う（親ブロック継続は呼び出し側で判断）
		tt := strings.TrimSpace(raw)
		if tt == "<" || tt == "<<" || tt == "<<<" {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}

// convertTables は PukiWiki のテーブルを Markdown のテーブルに変換します
func convertTables(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "|") && strings.Contains(line, "|") {
			// テーブルブロックを開始
			var tableLines []string
			// テーブル行の後ろにぶら下がっているテキスト（最終行の「|」以降）を集める
			var trailingAfterTable []string
			j := i
			for j < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[j]), "|") {
				// クリーンなテーブル行を追加
				tableLines = append(tableLines, cleanTableLine(lines[j]))
				// 各行について、最後のパイプ以降の文字列を後段に分離しておく
				raw := lines[j]
				if idx := strings.LastIndex(raw, "|"); idx >= 0 && idx < len(raw)-1 {
					tail := strings.TrimSpace(raw[idx+1:])
					if cleaned := cleanTableTail(tail); cleaned != "" {
						trailingAfterTable = append(trailingAfterTable, cleaned)
					}
				}
				j++
			}
			// 1行以上ある場合、最初の行の後にセパレーターを追加
			if len(tableLines) > 1 {
				// 先頭・末尾のパイプを含むため、列数はパイプ数-1
				cells := strings.Count(tableLines[0], "|") - 1
				if cells > 0 {
					separator := "|" + strings.Repeat("---|", cells)
					tableLines = insert(tableLines, 1, separator)
				}
			}
			result = append(result, tableLines...)
			// テーブル直後に空行を1つ入れて、テーブルと後続テキストを明確に分離
			if len(trailingAfterTable) > 0 {
				// テールがある場合は必ず1行空ける
				result = append(result, "")
			} else if j < len(lines) {
				// 次の行が非テーブルかつ空行でない場合も、確実に空行を挿入してブロックを分離
				next := strings.TrimSpace(lines[j])
				if next != "" && !strings.HasPrefix(next, "|") {
					// 直前が既に空行でないなら追加
					if len(result) == 0 || result[len(result)-1] != "" {
						result = append(result, "")
					}
				}
			}
			// テーブル直後に、ぶら下がっていたテキストを別行として出力（改行を保証）
			for _, tail := range trailingAfterTable {
				// 既に変換済みのテキストをそのまま出力
				result = append(result, tail)
			}
			i = j
		} else {
			// 非テーブル行の汎用チルダ(~)→<br /> 置換。
			// ただし、見出し行（直前の変換で Markdown の '#' 見出しになっている行）は対象外。
			trim := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trim, "#") {
				// 見出し中の '~' はそのまま残す
				result = append(result, lines[i])
			} else {
				result = append(result, strings.ReplaceAll(lines[i], "~", "<br />"))
			}
			i++
		}
	}
	return strings.Join(result, "\n")
}

// cleanTableLine はテーブル行から PukiWiki 特有の構文を削除します
func cleanTableLine(line string) string {
	if !strings.HasPrefix(line, "|") {
		return line
	}

	// 行の最初の | から最後の | までをテーブルとして扱い、それ以降のテキストは別処理とする
	start := strings.Index(line, "|")
	end := strings.LastIndex(line, "|")
	if start == -1 || end <= start {
		return line
	}
	inner := line[start+1 : end]
	// | で分割してセルを取得（空セルも保持して列数を維持する）
	parts := strings.Split(inner, "|")
	validParts := make([]string, len(parts))
	for i, part := range parts {
		validParts[i] = strings.TrimSpace(part)
	}

	// 最后が h の場合、ヘッダーとして除去
	if len(validParts) > 0 && strings.TrimSpace(validParts[len(validParts)-1]) == "h" {
		validParts = validParts[:len(validParts)-1]
	}

	// 各セルをクリーンアップ
	for i, part := range validParts {
		// アライメントプレフィックスを削除
		validParts[i] = reCellAlign.ReplaceAllString(part, "")
		// セル内の単独の "~" は PukiWiki の rowspan 指定（直上セルを継続）を意味する。
		// Markdown では rowspan が表現できないため、空セルとして出力する。
		if strings.TrimSpace(validParts[i]) == "~" {
			validParts[i] = ""
		}
		// セル先頭の "~" は PukiWiki ではセル内ヘッダ指定（Hugo では不要）なので削除する。
		// ただし、上の条件で単独 "~" は既に空セル化済みのため、ここでは内容を持つケースのみが対象。
		if strings.HasPrefix(validParts[i], "~") {
			// 先頭の "~" を1個取り除き、直後の空白もトリム
			validParts[i] = strings.TrimLeft(validParts[i][1:], " \t")
		}
		// PukiWiki のセル先頭の '>' は幅指定等で使われるため、先頭の > のみ削除。
		// HTML タグの '>' を壊さないように全削除はしない。
		validParts[i] = reLeadingGt.ReplaceAllString(validParts[i], "")
	}

	// Markdown テーブル形式に再構築
	return "|" + strings.Join(validParts, "|") + "|"
}

// insert は指定されたインデックスに値をスライスに挿入します
func insert(slice []string, index int, value string) []string {
	if index < 0 || index > len(slice) {
		panic("insert: index out of range")
	}
	newSlice := make([]string, len(slice)+1)
	copy(newSlice[:index], slice[:index])
	newSlice[index] = value
	copy(newSlice[index+1:], slice[index:])
	return newSlice
}

// slugify はページ名からURL-safeなスラッグを作成します
func slugify(name string) string {
	// types.Slugify に委譲して重複実装を回避
	return types.Slugify(name)
}

// enforceTableRowTailSeparation は、テーブル行の最後の '|' 以降に余分な文字列がある場合に
// それを次行へ分離し、テーブルブロックとの間に空行を1つ挿入して返します。
func enforceTableRowTailSeparation(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "|") {
			// テーブル行の最後の '|' を探し、そこから後ろに非空白があれば分離
			last := strings.LastIndex(raw, "|")
			if last >= 0 && last < len(raw)-1 {
				tail := strings.TrimSpace(raw[last+1:])
				if tail != "" {
					// テーブル行本体のみ出力
					base := strings.TrimRight(raw[:last+1], " \t")
					out = append(out, base)
					// 次行が既に空行でなければ空行を1つ入れる
					if len(out) == 0 || out[len(out)-1] != "" {
						out = append(out, "")
					}
					// tail を独立行として出力
					out = append(out, tail)
					continue
				}
			}
		}
		out = append(out, raw)
	}
	// テーブル直後の空行過多を防ぐ: 連続する3行以上の空行は2行までに圧縮（保守的）
	// 今回の用途ではそのままでも問題ないが、将来の保全のため軽く正規化
	normalized := make([]string, 0, len(out))
	emptyCount := 0
	for _, l := range out {
		if strings.TrimSpace(l) == "" {
			emptyCount++
			if emptyCount <= 2 { // 2連続まで許容
				normalized = append(normalized, "")
			}
		} else {
			emptyCount = 0
			normalized = append(normalized, l)
		}
	}
	return strings.Join(normalized, "\n")
}
