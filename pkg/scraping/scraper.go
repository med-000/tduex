package scraping

import "github.com/med-000/tduscheexport/pkg/logger"

type Scraper struct {
	log *logger.ScraperLogger
}

func NewScraper(log *logger.ScraperLogger) *Scraper {
	return &Scraper{
		log: log,
	}
}
