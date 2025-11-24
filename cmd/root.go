package cmd

import (
    "fmt"
    "github.com/massy22/pukiwki2hugo/internal/converter"
    "github.com/massy22/pukiwki2hugo/internal/input"
    "github.com/massy22/pukiwki2hugo/internal/types"
    "github.com/spf13/cobra"
    "log"
    "os"
    "path/filepath"
    "strings"
    "time"
)

var rootCmd = &cobra.Command{
	Use:   "pukiwki2hugo",
	Short: "Convert PukiWiki to Hugo",
	Long: `A tool to migrate PukiWiki sites to Hugo static websites.

Pass the path to your PukiWiki directory (containing wiki/, plugin/, etc.)
and get a complete Hugo site structure.`,
	Run: func(cmd *cobra.Command, args []string) {
			// デフォルトアクション

	},
}

var inputDir string
var outputDir string
var generateGone bool

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	convertCmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert PukiWiki site to Hugo",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("変換を開始します...")
			pages, err := input.ReadPages(inputDir)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%d ページが見つかりました", len(pages))

			defaultPage, err := input.GetDefaultPage(inputDir)
			if err != nil {
				log.Fatal(err)
			}
			for _, page := range pages {
				converted := converter.ConvertPukiToMd(page.Content)
				var outputFile string
				if page.Name == defaultPage {
					outputFile = filepath.Join(outputDir, "content", "_index.md")
				} else {
					outputFile = filepath.Join(outputDir, "content", "docs", page.Slug, "_index.md")
				}

                // 入れ子のページは、front matter の title/slug に親を含めない（葉のみ）
                displayTitle := page.Name
                displaySlug := page.Slug
                if page.Name != defaultPage {
                    parts := strings.Split(page.Name, "/")
                    if len(parts) > 1 {
                        leaf := parts[len(parts)-1]
                        displayTitle = leaf
                        displaySlug = types.Slugify(leaf)
                    }
                }

                os.MkdirAll(filepath.Dir(outputFile), 0755)
                // YAML フロントマターのインデントが混入しないよう、先頭に余白のないテンプレートを使用
                frontMatter := fmt.Sprintf(`---
title: "%s"
date: %s
lastmod: %s
slug: "%s"
draft: false
---

%s`, yamlEscape(displayTitle), page.Date.Format(time.RFC3339), page.Date.Format(time.RFC3339), displaySlug, converted)
                _ = os.WriteFile(outputFile, []byte(frontMatter), 0644)
            }

			if generateGone {
				createGoneMapping(pages, outputDir)
			}


		},
	}

	convertCmd.Flags().StringVarP(&inputDir, "input", "i", ".", "Path to PukiWiki root directory")
	convertCmd.Flags().StringVarP(&outputDir, "output", "o", "hugo-site", "Output directory for Hugo site")
	convertCmd.Flags().BoolVarP(&generateGone, "gone", "g", false, "Generate Gone redirects mapping")

	rootCmd.AddCommand(convertCmd)
}

// yamlEscape は YAML のダブルクォート文字列内で必要なエスケープを行います。
// 現状ではタイトルに含まれる `"` を `\"` に置換して安全に埋め込めるようにします。
func yamlEscape(s string) string {
    // バックスラッシュ→エスケープ、次にダブルクォートをエスケープ
    // 既にバックスラッシュが含まれている場合を考慮して順序に注意
    s = strings.ReplaceAll(s, "\\", "\\\\")
    s = strings.ReplaceAll(s, "\"", "\\\"")
    return s
}

func createGoneMapping(pages []*types.Page, outputDir string) {
	file, err := os.Create(filepath.Join(outputDir, "gone-redirects.yaml"))
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	for _, page := range pages {
		file.WriteString("- url: \"/wiki/" + page.Name + "\"\n")
		file.WriteString("  code: 410\n")
	}
}
