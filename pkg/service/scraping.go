package service

import (
	"fmt"
	"os"

	"github.com/gocolly/colly"
	"github.com/med-000/tduscheexport/pkg/logger"
	"github.com/med-000/tduscheexport/pkg/parser"
	"github.com/med-000/tduscheexport/pkg/scraping"
)

func (s *Service) FetchAll(req GetCourseRequest) (*parser.Course, error) {
	allowdomain := os.Getenv("ALLOW_DOMAIN")

	c := colly.NewCollector(
		colly.AllowedDomains(allowdomain),
	)

	scraperlogger, _ := logger.NewScraperLogger()
	sc := scraping.NewScraper(scraperlogger)

	parserlogger, _ := logger.NewParserLogger()
	p := parser.NewParser(parserlogger)

	s.log.Info.Printf("start FetchAll user=%s year=%d term=%d", req.UserID, req.Year, req.Term)

	coursehtml, err := sc.FetchCourseHTML(c, req.UserID, req.Password, req.Year, req.Term)
	if err != nil {
		s.log.Error.Printf("failed FetchCourseHTML: %v", err)
		return nil, err
	}
	s.log.Info.Printf("fetched course html")

	courses := p.ParseCourse(coursehtml)
	if courses == nil {
		s.log.Error.Printf("ParseCourse returned nil")
		return nil, fmt.Errorf("parse course failed")
	}

	classes := courses.Classes
	s.log.Info.Printf("parsed course classes=%d", len(classes))

	courseID := makeCourseID(req.Year, req.Term)

	var classresult []*parser.Class

	for i := range classes {
		s.log.Info.Printf("fetch class[%d] title=%s url=%s", i, classes[i].Title, classes[i].URL)

		classhtml, err := sc.FetchClassHTML(c, classes[i].URL)
		if err != nil {
			s.log.Error.Printf("failed FetchClassHTML index=%d url=%s err=%v", i, classes[i].URL, err)
			continue
		}

		class := p.ParserClass(classhtml)
		if class == nil {
			s.log.Error.Printf("ParserClass returned nil index=%d url=%s", i, classes[i].URL)
			continue
		}

		var eventCount int
		var contentSuccess int

		for ei, e := range class.Events {
			eventCount++

			if e.URL == "" {
				s.log.Warn.Printf("skip empty event url class=%d event=%d", i, ei)
				continue
			}

			contenthtml, err := sc.FetchContentHTML(c, e.URL)
			if err != nil {
				s.log.Error.Printf("failed FetchContentHTML class=%d event=%d url=%s err=%v", i, ei, e.URL, err)
				continue
			}

			contents, err := p.ParserContent(contenthtml)
			if err != nil {
				s.log.Error.Printf("failed ParseContent class=%d event=%d url=%s err=%v", i, ei, e.URL, err)
				continue
			}

			e.Content = contents
			contentSuccess++
		}

		s.log.Info.Printf("class parsed index=%d title=%s events=%d contents_attached=%d", i, classes[i].Title, eventCount, contentSuccess)

		classresult = append(classresult, &parser.Class{
			ExternalId: classes[i].ExternalId,
			Day:        classes[i].Day,
			Period:     classes[i].Period,
			Title:      classes[i].Title,
			URL:        classes[i].URL,
			Events:     class.Events,
		})
	}

	s.log.Info.Printf("FetchAll completed classes=%d", len(classresult))

	return &parser.Course{
		ExternalId: courseID,
		Year:       req.Year,
		Term:       req.Term,
		Classes:    classresult,
	}, nil
}

// CourseIDの変換関数
func makeCourseID(year int, term int) string {
	return fmt.Sprintf("%d_%d", year, term)
}
