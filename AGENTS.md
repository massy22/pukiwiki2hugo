# PukiWiki to Hugo 変換ツール開発者向けAIエージェントガイド

最終更新: 2025-11-30 00:33 JST

このプロジェクトでは、日本語でやり取りすることを推奨します。コミュニケーションは日本語で実施してください。

このドキュメントは、PukiWiki to Hugo 変換ツールプロジェクトにおける AI アシスタント（例: Junie など）を使用した開発ガイドラインを提供します。

## AIエージェントセットアップ

このプロジェクトでは、コード分析、編集、開発ワークフロー管理に AI エージェントを使用します。

### 推奨AIエージェント機能

- **コード理解**: Goコード構造、パッケージ、依存関係の解析
- **Regex検索**: コードベース内のテキスト検索
- **シンボル編集**: コンテキスト認識による安全なコード修正
- **エラー分析**: コンパイルエラーと提案のデバッグ

### エージェントを使用した開発ワークフロー

1. **計画**: エージェントが要件を分析し、タスク分解を作成
2. **実装**: エージェントがGoモジュール/パッケージのコード作成
3. **テスト**: エージェントがビルドコマンドと実行テストを実行
4. **デバッグ**: エージェントが構文/実行時エラーを特定・修正
5. **ドキュメント**: エージェントがREADME、コードコメントを更新

## コード構造認識

- **main.go**: エントリーポイント、Cobra CLI統合
- **cmd/root.go**: CLIコマンドとフラグ、変換ロジック
- **internal/converter/**: PukiWikiからMarkdownへの構文変換
- **internal/input/**: ページ読み込みと設定解析
- **internal/types/**: ページ構造とslugify関数

### 主なコンポーネント

- `Converter.ConvertPukiToMd()`: Regexベースの構文変換
- `Input.ReadPages()`: 16進デコードファイル名処理
- `Types.Page.slugify()`: URL-safe slug生成
- FrontMatter: Hugo互換メタデータ生成

## エージェント実行共通コマンド

```bash
# ビルドとテスト
go mod tidy
go build -o pukiwki2hugo .
./pukiwki2hugo convert -i sample_pukiwiki -o test-output --gone

# コード分析
# エージェント機能例: シンボル検索、テキスト検索、リネーム/置換（IDE のリファクタリング機能等）

# エラーデバッグ
# 構文チェック: go build -v
# 実行時: コード内にログ追加, fmt.Println

# 構造確認
find docs/ -name "*.md" | head -10 # ファイル作成確認
grep "^###\|^##\|^\*\{1,3\}" test.md | head -5 # ヘッダー変換確認
```

## エージェント使用ベストプラクティス

### コード編集の場合

- シンボル操作でフルコンテキスト意識を使用
- Goモジュール構造とimportを保存
- 主要変更後ビルドテスト
- 説明的なエラーメッセージとログ利用

### テストの場合

- サンプルデータで一貫したテスト
- 出力構造がHugo要件に合致するか確認
- リンクパスの解決正しさ確認


#### テスト実行方法

基本は Go 標準の `go test` を使用します。プロジェクトルートで実行してください。

```bash
# 全パッケージのテスト実行（推奨）
go test ./...

# 個別パッケージのテスト実行
go test ./internal/converter
go test ./internal/input
go test ./internal/types

# 詳細とカバレッジ（概要％）
go test -v -cover ./...

# カバレッジレポート（HTML）
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html

# データ競合の検出（重いので必要時）
go test -race ./...
```

注意事項:

- zsh 等で `./...` が補完候補として誤修正される場合は、プロジェクトルートで実行しているか確認してください。
  - それでも補正される場合は `GOFLAGS="-count=1" go test ./...` のように先頭に環境変数を付ける、
    もしくは `bash -lc 'go test ./...'` とシェルを切り替えて実行してください。
- Windows PowerShell では `go test ./...` をそのまま実行できます。

テスト追加のベストプラクティス:

- テーブル駆動テスト(Table-driven tests)を使用し、複数のケースをまとめてテスト
- テストファイル命名: `{package}_test.go`
- 期待する振る舞い（仕様）をテストに固定化し、変換ロジックのリグレッションを検出できるようにする

テスト追加例（テスト関数内）:

```go
package example_test

import "testing"

// ドキュメント用の最小サンプル。実際の対象関数に置き換えて利用してください。
func TestExample(t *testing.T) {
    input := "テスト入力"
    expected := "期待出力"
    // 実際には対象関数を呼び出してください: result := targetFunction(input)
    result := input // ダミー実装
    if result != expected {
        t.Errorf("expect %q; got %q", expected, result)
    }
}
```


### ドキュメントの場合

- 新機能にREADME.md更新
- コメント内に使用例を含める
- メンテナンス用にRegexパターンを文書化

## エージェントを使用したトラブルシューティング

### 一般的な問題

1. **Regexパターンエラー**
   - エージェントが失敗パターンを分析
   - 正しい正規表現を提案

2. **パス解決**
   - エージェントがslugifyとパス構築ロジックを確認
   - Hugoディレクトリ構造準拠を確認

3. **Import問題**
   - エージェントがパッケージimport循環を解決
   - modファイル更新を確認

### デバッグコマンド

```bash
# Goモジュール確認
go list -m all

# 冗長ビルド
go build -v 2>&1 | grep -E "error|warning"

# 変換出力テスト
ls -la generated/content/
cat generated/content/_index.md | head -20
```

## 進化と拡張

プロジェクトの成長に伴い、エージェント能力も進化させる:

- **プラグインサポート**: 追加PukiWikiプラグイン変換
- **設定**: 変換ルールを設定可能に
- **テーマ**: Hugoテーマ統合
- **SEO最適化**: メタデータ抽出強化
