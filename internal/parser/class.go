package parser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseClass course.phpのhtml専用
func (p *Parser) ParserClass(html string) *Class {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		p.log.Error.Printf("Cannt Read Html \n Error Detail:%s", err)
		return nil
	}

	class := &Class{
		Title: strings.TrimSpace(
			doc.Find("a.course-name").First().Text(),
		),
		Events: []*Event{},
	}

	// folder（group）ごとに回す
	doc.Find(".cl-contentsList_folder").Each(func(i int, folder *goquery.Selection) {

		groupName := strings.TrimSpace(folder.Find(".panel-title").Text())

		// eventを回す
		folder.Find(".cl-contentsList_listGroupItem").Each(func(j int, item *goquery.Selection) {

			name := strings.TrimSpace(item.Find(".cm-contentsList_contentName").Text())
			category := strings.TrimSpace(item.Find(".cl-contentsList_categoryLabel").Text())
			date := strings.TrimSpace(item.Find(".cm-contentsList_contentDetailListItemData").Text())

			var id string
			var fullURL string

			// id取得
			if a := item.Find(".cl-contentsList_contentDetailListItemData a"); a.Length() > 0 {
				href, _ := a.Attr("href")

				parts := strings.Split(href, "/contents/")
				if len(parts) > 1 {
					id = strings.Split(parts[1], "/")[0]
					if id == "" {
						p.log.Error.Printf("Id is nil")
					}
				}
			}

			// URL取得
			if a := item.Find(".cl-contentsList_contentInfo a"); a.Length() > 0 {
				link, _ := a.Attr("href")
				if link != "" {
					fullURL = "https://els.sa.dendai.ac.jp" + link
				} else {
					p.log.Error.Printf("Link is nil")
				}
			}

			// Event生成
			e := &Event{
				ExternalId: id,
				Name:       name,
				Category:   category,
				URL:        fullURL,
				Date:       date,
				GroupName:  groupName,
			}
			class.Events = append(class.Events, e)
		})
	})

	return class
}
