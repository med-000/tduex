package parser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (p *Parser) ParserContent(html string) ([]*Content, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		p.log.Error.Printf("Cannt Read Html \n Error Detail:%s", err)
		return nil, err
	}

	var contents []*Content

	doc.Find(`a[href*="file_down.php"]`).Each(func(i int, a *goquery.Selection) {
		href, exists := a.Attr("href")
		if !exists || href == "" {
			p.log.Error.Printf("Notfound href:%s", err)
			return
		}

		filename := strings.TrimSpace(a.Text())
		filename = strings.TrimPrefix(filename, "»")
		filename = strings.TrimSpace(filename)

		contents = append(contents, &Content{
			URL:      "https://els.sa.dendai.ac.jp/" + href,
			FileName: filename,
		})
	})

	return contents, nil
}
