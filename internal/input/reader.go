package input

import (
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/massy22/pukiwki2hugo/internal/types"
)

func ReadPages(inputDir string) ([]*types.Page, error) {
	var pages []*types.Page

	err := filepath.WalkDir(filepath.Join(inputDir, "wiki"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".txt") {
			return nil
		}

		filename := strings.TrimSuffix(filepath.Base(path), ".txt")
		pageName, err := decodePageName(filename)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		date := time.Now()
		page := types.NewPage(pageName, string(content), date)
		pages = append(pages, page)

		return nil
	})

	return pages, err
}

func decodePageName(encoded string) (string, error) {
	b, err := hex.DecodeString(encoded)
	return string(b), err
}

func GetDefaultPage(inputDir string) (string, error) {
	filePath := filepath.Join(inputDir, "pukiwiki.ini.php")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "FrontPage", nil
	}
	lines := strings.Split(string(content), "\n")
	re := regexp.MustCompile(`\$defaultpage\s*=\s*['"]([^'"]+)['"];?`)
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			defaultpage := matches[1]
			return defaultpage, nil
		}
	}
	return "FrontPage", nil
}
