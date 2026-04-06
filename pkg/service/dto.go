package service

type GetCourseRequest struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
	Year     int    `json:"year"`
	Term     int    `json:"term"`
}
