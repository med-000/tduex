package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseCourse index.phpのhtml専用
func (p *Parser) ParseCourse(html string) *Course {
	//htmlをdocに変換
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		p.log.Error.Printf("Cannt Read Html \n Error Detail:%s", err)
		return nil
	}

	//年度取得
	yearStr := doc.Find(`select[name="year"] option[selected]`).AttrOr("value", "")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		p.log.Error.Printf("failed to parse year: value=%s err=%v", yearStr, err)
	}
	//学期取得
	termStr := doc.Find(`select[name="semester"] option[selected]`).AttrOr("value", "")
	term, err := strconv.Atoi(termStr)
	if err != nil {
		p.log.Error.Printf("failed to parse term: value=%s err=%v", termStr, err)
	}

	var classes []*Class

	//#schedule-table tbody trというcssセレクタからテーブルを抜き出して一つづつループ処理
	doc.Find("#schedule-table tbody tr").Each(func(i int, tr *goquery.Selection) {
		periodText := strings.TrimSpace(tr.Find(".schedule-table-class_order").Text())
		periodText = strings.ReplaceAll(periodText, "限", "")

		period, err := strconv.Atoi(periodText)
		if err != nil || period <= 0 {
			return
		}

		tr.Find("td").Each(func(j int, td *goquery.Selection) {
			//リンクタグだけ抽出して
			a := td.Find("a")
			if a.Length() == 0 {
				return
			}

			// title抜き出し
			rawTitle := strings.TrimSpace(a.Text())
			title := titleTrimmer(rawTitle)
			if title == "" {
				p.log.Warn.Printf("Undefind Class Title")
				return
			}

			//hrefなし弾き
			link, exists := a.Attr("href")
			if !exists || link == "" {
				day := dayTranslator(j)
				p.log.Error.Printf("href missing or empty: title=%s day=%s曜日 period=%d限", title, day, period)
				return
			}

			// リンク作成
			fullURL := link
			if !strings.HasPrefix(link, "http") {
				fullURL = "https://els.sa.dendai.ac.jp" + link
			}

			// ID抽出（安全版)
			id, err := extractCourseID(fullURL)
			if err != nil {
				p.log.Warn.Printf("failed to extract course id: url=%s err=%v", fullURL, err)
				return
			}

			classes = append(classes, &Class{
				ExternalId: id,
				Day:        j,
				Period:     period,
				Title:      title,
				URL:        fullURL,
			})
		})
	})

	p.log.Info.Printf("Success Return Couses")
	return &Course{
		ExternalId: fmt.Sprintf("%d-%d", year, term),
		Year:       year,
		Term:       term,
		Classes:    classes,
	}
}

// titleの不要な部分を削る
var trailingCodeRe = regexp.MustCompile(`$begin:math:text$\[\^\)\]\*$end:math:text$$`)

func titleTrimmer(s string) string {
	s = strings.TrimPrefix(s, "» ")
	s = strings.TrimSpace(s)

	if idx := strings.Index(s, "("); idx != -1 {
		s = s[:idx]
	}

	return strings.TrimSpace(s)
}

// courseIDを切り取る
func extractCourseID(fullURL string) (string, error) {
	const prefix = "/course.php/"

	idx := strings.Index(fullURL, prefix)
	if idx == -1 {
		return "", fmt.Errorf("course id not found in url: %s", fullURL)
	}

	idPart := fullURL[idx+len(prefix):]

	parts := strings.Split(idPart, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", fmt.Errorf("invalid course id format: %s", fullURL)
	}

	return parts[0], nil
}

var days = []string{"", "月", "火", "水", "木", "金", "土", "日"}

// dayの数から曜日変換(0からスタート)
func dayTranslator(day int) string {
	if day < 0 || day >= len(days) {
		return ""
	}
	return days[day]
}
