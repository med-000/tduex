package service

import (
	"github.com/med-000/notifyclass/pkg/logger"
	"github.com/med-000/notifyclass/pkg/parser"
	"github.com/med-000/notifyclass/pkg/repository"
	"gorm.io/gorm"
)

func (s *Service) SaveAll(dbConn *gorm.DB, course *parser.Course) error {
	repositoryLogger, _ := logger.NewRepositoryLogger()
	courseRepo := repository.NewCourseRepository(dbConn, repositoryLogger)
	classRepo := repository.NewClassRepository(dbConn, repositoryLogger)
	eventRepo := repository.NewEventRepository(dbConn, repositoryLogger)
	contentRepo := repository.NewContentRepository(dbConn, repositoryLogger)

	dbCourse := repository.ToDBCourse(course)
	err := courseRepo.Save(dbCourse)
	if err != nil {
		s.log.Error.Printf("defined course save err:%s", err)
		return err
	}
	s.log.Info.Printf("save course")

	for _, class := range course.Classes {
		dbClass := repository.ToDBClass(class, dbCourse.ID)
		err := classRepo.Save(dbClass)
		if err != nil {
			s.log.Error.Printf("defined class save err:%s", err)
			return err
		}

		for _, event := range class.Events {
			dbEvent := repository.ToDBEvent(event, dbClass.ID)
			err := eventRepo.Save(dbEvent)
			if err != nil {
				s.log.Error.Printf("defined event save err:%s", err)
				return err
			}
			for _, content := range event.Content {
				dbContent := repository.ToDBContent(content, dbEvent.ID)
				err := contentRepo.Save(dbContent)
				if err != nil {
					s.log.Error.Printf("defined content save err:%s", err)
					return err
				}
			}
			s.log.Info.Printf("save content")
		}
		s.log.Info.Printf("save event")

		s.log.Info.Printf("save group")
	}
	s.log.Info.Printf("save class")
	return nil
}
