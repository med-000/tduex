package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/med-000/tduscheexport/pkg/parser"
)

func (s *Service) FetchAllAndExportJSON(req GetCourseRequest, savePath string) (*ExportCourse, error) {
	return s.FetchClassesAndExportJSON(req, savePath)
}

func (s *Service) FetchClassesAndExportJSON(req GetCourseRequest, savePath string) (*ExportCourse, error) {
	course, err := s.FetchClasses(req)
	if err != nil {
		return nil, err
	}

	exportCourse := convertCourseForExport(course)
	if err := s.ExportCourseJSON(exportCourse, savePath); err != nil {
		return nil, err
	}

	return exportCourse, nil
}

func (s *Service) FetchFullAndExportJSON(req GetCourseRequest, savePath string) (*FullExportCourse, error) {
	course, err := s.FetchFull(req)
	if err != nil {
		return nil, err
	}

	exportCourse := convertFullCourseForExport(course)
	if err := s.ExportFullCourseJSON(exportCourse, savePath); err != nil {
		return nil, err
	}

	return exportCourse, nil
}

func (s *Service) ExportCourseJSON(course *ExportCourse, savePath string) error {
	if course == nil {
		return fmt.Errorf("course is nil")
	}
	return writeJSONFile(savePath, course)
}

func (s *Service) ExportFullCourseJSON(course *FullExportCourse, savePath string) error {
	if course == nil {
		return fmt.Errorf("course is nil")
	}
	return writeJSONFile(savePath, course)
}

func convertCourseForExport(course *parser.Course) *ExportCourse {
	if course == nil {
		return nil
	}

	classes := make([]*ExportClass, 0, len(course.Classes))
	for _, class := range course.Classes {
		if class == nil {
			continue
		}

		classes = append(classes, &ExportClass{
			ExternalID: class.ExternalId,
			Day:        class.Day,
			Period:     class.Period,
			Title:      class.Title,
		})
	}

	return &ExportCourse{
		ExternalID: course.ExternalId,
		Year:       course.Year,
		Term:       course.Term,
		Classes:    classes,
	}
}

func convertFullCourseForExport(course *parser.Course) *FullExportCourse {
	if course == nil {
		return nil
	}

	classes := make([]*FullExportClass, 0, len(course.Classes))
	for _, class := range course.Classes {
		if class == nil {
			continue
		}

		events := make([]*FullExportEvent, 0, len(class.Events))
		for _, event := range class.Events {
			if event == nil {
				continue
			}

			events = append(events, &FullExportEvent{
				ExternalID: event.ExternalId,
				Name:       event.Name,
				Category:   event.Category,
				Date:       event.Date,
				GroupName:  event.GroupName,
			})
		}

		classes = append(classes, &FullExportClass{
			ExternalID: class.ExternalId,
			Day:        class.Day,
			Period:     class.Period,
			Title:      class.Title,
			Events:     events,
		})
	}

	return &FullExportCourse{
		ExternalID: course.ExternalId,
		Year:       course.Year,
		Term:       course.Term,
		Classes:    classes,
	}
}

func writeJSONFile(savePath string, value any) error {
	if savePath == "" {
		return fmt.Errorf("save path is empty")
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(savePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(savePath, data, 0644)
}
