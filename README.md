# PukiWiki to Hugo Converter

PukiWiki 1.5.4 から Hugo 静的サイトジェネレーターへの移行ツールです。
Go言語で実装された CLI ツールで、Wiki ページの構文変換と Hugo サイト構造生成を行います。

## Features

- PukiWiki 構文変換（Markdown/Hugo 互換）
  - 見出し（`*`/`**`/`***`）、アンカー除去（`[#id]`）
  - 内部/外部リンク、別名リンク、アンカー付きリンク
  - テーブル（セル整形、ヘッダ指定 `~` の除去、行末 tail 分離）
  - 箇条書き（`-`）/番号付きリスト（`+`）、引用（`>`）
  - インライン強調／斜体（`''`/`'''`）
  - インラインプラグイン: `&size(...)`, `&color(...)`, `&br;`, `&new{...}`, `&counter(...)`, `&online`
  - ブロックプラグイン: `#recent(n)` の除去（改行に正規化）、`#author(...)`/`#freeze(...)` 行の削除
- Hugo 構造生成: `content/` 配下に Front Matter 付きファイルを出力
- デフォルトページ処理: `pukiwiki.ini.php` の `$defaultpage` を解析してトップの `_index.md` を作成
- Gone マッピング生成（オプション）: 旧 URL に対する 410 Gone の一覧を出力

## Installation

```bash
go mod tidy
go build -o pukiwki2hugo .
```

## Usage

```bash
./pukiwki2hugo convert -i <PukiWikiディレクトリ> -o <出力ディレクトリ> --gone
```

### Options

- `-i, --input`: PukiWiki root directory (default: ".")
- `-o, --output`: Hugo site output directory (default: "hugo-site")
- `-g, --gone`: Generate gone-redirects.yaml for SEO

### Examples

```bash
# Convert sample wiki
./pukiwki2hugo convert -i sample_pukiwiki -o hugo-site --gone

# View help
./pukiwki2hugo --help
```

## Output Structure

```
hugo-site/
├── content/
│   ├── _index.md          # Default page from pukiwiki.ini.php
│   └── docs/
│       ├── ガイド/_index.md
│       ├── ガイド/第1章/_index.md
│       └── ...
└── gone-redirects.yaml    # SEO mappings
```

## Front Matter Format

```yaml
---
title: "ページ名"
date: 2025-11-24T10:00:00Z
lastmod: 2025-11-24T10:00:00Z
slug: "ページ名"
draft: false
---

コンテンツ...
```

## Requirements

- Go 1.21+
- PukiWiki 1.5.4 互換ディレクトリ

## License

MIT License

本ツールは PukiWiki のソースコードを利用せず、PukiWiki 互換の“記法/データ形式”を入力として扱う独立実装です。そのため、PukiWiki 本体のライセンス（GPL v2 以降）の拘束は受けません。プロジェクトのソースコードは MIT ライセンスの下で配布します。

注意事項:
- 生成対象となる各 Wiki ページの“内容”の著作権・ライセンスは、元ページの著作者/サイトの規約に従います（本ツールのライセンスとは独立です）。
- 依存ライブラリはそれぞれのライセンスに従います（例: `spf13/cobra` は Apache-2.0）。再配布時は各ライブラリのライセンス表記の保持にご留意ください。

## Author


