package service

import (
	"fmt"
	"os"

	"github.com/gocolly/colly"
	"github.com/med-000/tduscheexport/internal/logger"
	"github.com/med-000/tduscheexport/internal/parser"
	"github.com/med-000/tduscheexport/internal/scraping"
)

func (s *Service) FetchAll(req GetCourseRequest) (*parser.Course, error) {
	return s.FetchClasses(req)
}

func (s *Service) FetchClasses(req GetCourseRequest) (*parser.Course, error) {
	sc, p, c := s.newScrapingContext()

	s.log.Info.Printf("start FetchClasses user=%s year=%d term=%d day=%d period=%d", req.UserID, req.Year, req.Term, req.Day, req.Period)

	coursehtml, err := sc.FetchCourseHTML(c, req.UserID, req.Password, req.Year, req.Term)
	if err != nil {
		s.log.Error.Printf("failed FetchCourseHTML: %v", err)
		return nil, err
	}

	courses := p.ParseCourse(coursehtml)
	if courses == nil {
		s.log.Error.Printf("ParseCourse returned nil")
		return nil, fmt.Errorf("parse course failed")
	}

	courseID := makeCourseID(req.Year, req.Term)
	classresult := make([]*parser.Class, 0, len(courses.Classes))

	for i := range courses.Classes {
		if !matchesClassFilter(courses.Classes[i], req.Day, req.Period) {
			s.log.Info.Printf("skip class[%d] by filter title=%s day=%d period=%d", i, courses.Classes[i].Title, courses.Classes[i].Day, courses.Classes[i].Period)
			continue
		}

		s.log.Info.Printf("class selected index=%d title=%s day=%d period=%d", i, courses.Classes[i].Title, courses.Classes[i].Day, courses.Classes[i].Period)
		classresult = append(classresult, &parser.Class{
			ExternalId: courses.Classes[i].ExternalId,
			Day:        courses.Classes[i].Day,
			Period:     courses.Classes[i].Period,
			Title:      courses.Classes[i].Title,
			URL:        courses.Classes[i].URL,
		})
	}

	s.log.Info.Printf("FetchClasses completed classes=%d", len(classresult))

	return &parser.Course{
		ExternalId: courseID,
		Year:       req.Year,
		Term:       req.Term,
		Classes:    classresult,
	}, nil
}

func (s *Service) FetchFull(req GetCourseRequest) (*parser.Course, error) {
	sc, p, c := s.newScrapingContext()

	s.log.Info.Printf("start FetchFull user=%s year=%d term=%d day=%d period=%d", req.UserID, req.Year, req.Term, req.Day, req.Period)

	coursehtml, err := sc.FetchCourseHTML(c, req.UserID, req.Password, req.Year, req.Term)
	if err != nil {
		s.log.Error.Printf("failed FetchCourseHTML: %v", err)
		return nil, err
	}

	courses := p.ParseCourse(coursehtml)
	if courses == nil {
		s.log.Error.Printf("ParseCourse returned nil")
		return nil, fmt.Errorf("parse course failed")
	}

	courseID := makeCourseID(req.Year, req.Term)
	classresult := make([]*parser.Class, 0, len(courses.Classes))

	for i := range courses.Classes {
		if !matchesClassFilter(courses.Classes[i], req.Day, req.Period) {
			s.log.Info.Printf("skip class[%d] by filter title=%s day=%d period=%d", i, courses.Classes[i].Title, courses.Classes[i].Day, courses.Classes[i].Period)
			continue
		}

		s.log.Info.Printf("fetch class[%d] title=%s url=%s", i, courses.Classes[i].Title, courses.Classes[i].URL)

		classhtml, err := sc.FetchClassHTML(c, courses.Classes[i].URL)
		if err != nil {
			s.log.Error.Printf("failed FetchClassHTML index=%d url=%s err=%v", i, courses.Classes[i].URL, err)
			continue
		}

		class := p.ParserClass(classhtml)
		if class == nil {
			s.log.Error.Printf("ParserClass returned nil index=%d url=%s", i, courses.Classes[i].URL)
			continue
		}

		for ei, event := range class.Events {
			if event.URL == "" {
				s.log.Warn.Printf("skip empty event url class=%d event=%d", i, ei)
				continue
			}

			contenthtml, err := sc.FetchContentHTML(c, event.URL)
			if err != nil {
				s.log.Error.Printf("failed FetchContentHTML class=%d event=%d url=%s err=%v", i, ei, event.URL, err)
				continue
			}

			contents, err := p.ParserContent(contenthtml)
			if err != nil {
				s.log.Error.Printf("failed ParseContent class=%d event=%d url=%s err=%v", i, ei, event.URL, err)
				continue
			}

			event.Content = contents
		}

		classresult = append(classresult, &parser.Class{
			ExternalId: courses.Classes[i].ExternalId,
			Day:        courses.Classes[i].Day,
			Period:     courses.Classes[i].Period,
			Title:      courses.Classes[i].Title,
			URL:        courses.Classes[i].URL,
			Events:     class.Events,
		})
	}

	s.log.Info.Printf("FetchFull completed classes=%d", len(classresult))

	return &parser.Course{
		ExternalId: courseID,
		Year:       req.Year,
		Term:       req.Term,
		Classes:    classresult,
	}, nil
}

func (s *Service) newScrapingContext() (*scraping.Scraper, *parser.Parser, *colly.Collector) {
	allowdomain := os.Getenv("ALLOW_DOMAIN")

	c := colly.NewCollector(
		colly.AllowedDomains(allowdomain),
	)

	scraperlogger, _ := logger.NewScraperLogger()
	sc := scraping.NewScraper(scraperlogger)

	parserlogger, _ := logger.NewParserLogger()
	p := parser.NewParser(parserlogger)

	return sc, p, c
}

func makeCourseID(year int, term int) string {
	return fmt.Sprintf("%d_%d", year, term)
}

func matchesClassFilter(class *parser.Class, day int, period int) bool {
	if class == nil {
		return false
	}
	if day > 0 && class.Day != day {
		return false
	}
	if period > 0 && class.Period != period {
		return false
	}
	return true
}
