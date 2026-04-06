package service

type GetCourseRequest struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
	Year     int    `json:"year"`
	Term     int    `json:"term"`
	Day      int    `json:"day"`
	Period   int    `json:"period"`
}

type ExportCourse struct {
	ExternalID string         `json:"externalId"`
	Year       int            `json:"year"`
	Term       int            `json:"term"`
	Classes    []*ExportClass `json:"classes"`
}

type ExportClass struct {
	ExternalID string `json:"externalId"`
	Day        int    `json:"day"`
	Period     int    `json:"period"`
	Title      string `json:"title"`
}

type FullExportCourse struct {
	ExternalID string             `json:"externalId"`
	Year       int                `json:"year"`
	Term       int                `json:"term"`
	Classes    []*FullExportClass `json:"classes"`
}

type FullExportClass struct {
	ExternalID string             `json:"externalId"`
	Day        int                `json:"day"`
	Period     int                `json:"period"`
	Title      string             `json:"title"`
	Events     []*FullExportEvent `json:"events"`
}

type FullExportEvent struct {
	ExternalID string `json:"externalId"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	Date       string `json:"date"`
	GroupName  string `json:"groupName"`
}
