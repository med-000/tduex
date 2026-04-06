package scraping

import "github.com/med-000/notifyclass/pkg/logger"

type Scraper struct {
	log *logger.ScraperLogger
}

func NewScraper(log *logger.ScraperLogger) *Scraper {
	return &Scraper{
		log: log,
	}
}
