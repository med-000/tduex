package scraping

import (
	"fmt"

	"github.com/gocolly/colly"
)

func (s *Scraper) FetchContentHTML(c *colly.Collector, url string) (string, error) {
	var body string

	collector := c.Clone()

	collector.OnResponse(func(r *colly.Response) {
		body = string(r.Body)
	})

	if err := collector.Visit(url); err != nil {
		return "", err
	}
	collector.Wait()

	if body == "" {
		return "", fmt.Errorf("empty response: %s", url)
	}

	return body, nil
}
