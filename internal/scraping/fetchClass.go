package scraping

import (
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

// FetchClassHTML FetchCourseHTMLからもらったurlをもとにclassの詳細を返す 詳細を叩く前提
func (s *Scraper) FetchClassHTML(c *colly.Collector, url string) (string, error) {
	var html string

	baseURL := os.Getenv("BASE_URL")

	redirectRe := regexp.MustCompile(`window\.location\.href\s*=\s*"([^"]+)"`)

	//collyのコピー(ログイン情報などのコピー)
	cc := c.Clone()

	cc.OnResponse(func(r *colly.Response) {
		body := string(r.Body)

		// JSリダイレクト
		match := redirectRe.FindStringSubmatch(body)
		if len(match) > 1 {
			s.log.Info.Printf("Get Redirect URL by %s", body)
			next := baseURL + match[1]
			_ = r.Request.Visit(next)
			s.log.Info.Printf("Visit classURL")
			return
		}

		// 最終ページ
		if strings.Contains(r.Request.URL.Path, "course.php") {
			s.log.Info.Printf("Success get classHtml by course.php")
			html = body
		}
	})

	if err := cc.Visit(url); err != nil {
		s.log.Error.Printf("Cannt visit %s", url)
		return "", err
	}

	return html, nil
}
