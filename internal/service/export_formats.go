package service

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *Service) ExportCourseCSV(course *ExportCourse, savePath string) error {
	if course == nil {
		return fmt.Errorf("course is nil")
	}

	rows := [][]string{{"externalId", "year", "term", "day", "period", "title"}}
	for _, class := range course.Classes {
		if class == nil {
			continue
		}
		rows = append(rows, []string{
			class.ExternalID,
			fmt.Sprintf("%d", course.Year),
			fmt.Sprintf("%d", course.Term),
			fmt.Sprintf("%d", class.Day),
			fmt.Sprintf("%d", class.Period),
			class.Title,
		})
	}

	return writeCSVFile(savePath, rows)
}

func (s *Service) ExportFullCourseCSV(course *FullExportCourse, savePath string) error {
	if course == nil {
		return fmt.Errorf("course is nil")
	}

	rows := [][]string{{"classExternalId", "year", "term", "day", "period", "classTitle", "eventExternalId", "eventName", "category", "date", "groupName"}}
	for _, class := range course.Classes {
		if class == nil {
			continue
		}
		if len(class.Events) == 0 {
			rows = append(rows, []string{
				class.ExternalID,
				fmt.Sprintf("%d", course.Year),
				fmt.Sprintf("%d", course.Term),
				fmt.Sprintf("%d", class.Day),
				fmt.Sprintf("%d", class.Period),
				class.Title,
				"",
				"",
				"",
				"",
				"",
			})
			continue
		}

		for _, event := range class.Events {
			if event == nil {
				continue
			}
			rows = append(rows, []string{
				class.ExternalID,
				fmt.Sprintf("%d", course.Year),
				fmt.Sprintf("%d", course.Term),
				fmt.Sprintf("%d", class.Day),
				fmt.Sprintf("%d", class.Period),
				class.Title,
				event.ExternalID,
				event.Name,
				event.Category,
				event.Date,
				event.GroupName,
			})
		}
	}

	return writeCSVFile(savePath, rows)
}

func (s *Service) ExportCourseXLSX(course *ExportCourse, savePath string) error {
	if course == nil {
		return fmt.Errorf("course is nil")
	}

	rows := [][]string{{"externalId", "year", "term", "day", "period", "title"}}
	for _, class := range course.Classes {
		if class == nil {
			continue
		}
		rows = append(rows, []string{
			class.ExternalID,
			fmt.Sprintf("%d", course.Year),
			fmt.Sprintf("%d", course.Term),
			fmt.Sprintf("%d", class.Day),
			fmt.Sprintf("%d", class.Period),
			class.Title,
		})
	}

	return writeXLSXFile(savePath, "classes", rows)
}

func (s *Service) ExportFullCourseXLSX(course *FullExportCourse, savePath string) error {
	if course == nil {
		return fmt.Errorf("course is nil")
	}

	rows := [][]string{{"classExternalId", "year", "term", "day", "period", "classTitle", "eventExternalId", "eventName", "category", "date", "groupName"}}
	for _, class := range course.Classes {
		if class == nil {
			continue
		}
		if len(class.Events) == 0 {
			rows = append(rows, []string{
				class.ExternalID,
				fmt.Sprintf("%d", course.Year),
				fmt.Sprintf("%d", course.Term),
				fmt.Sprintf("%d", class.Day),
				fmt.Sprintf("%d", class.Period),
				class.Title,
				"",
				"",
				"",
				"",
				"",
			})
			continue
		}

		for _, event := range class.Events {
			if event == nil {
				continue
			}
			rows = append(rows, []string{
				class.ExternalID,
				fmt.Sprintf("%d", course.Year),
				fmt.Sprintf("%d", course.Term),
				fmt.Sprintf("%d", class.Day),
				fmt.Sprintf("%d", class.Period),
				class.Title,
				event.ExternalID,
				event.Name,
				event.Category,
				event.Date,
				event.GroupName,
			})
		}
	}

	return writeXLSXFile(savePath, "events", rows)
}

func (s *Service) ExportFullCourseICS(course *FullExportCourse, savePath string) error {
	if course == nil {
		return fmt.Errorf("course is nil")
	}

	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//tduex//EN\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")
	b.WriteString("X-WR-TIMEZONE:Asia/Tokyo\r\n")

	stamp := time.Now().UTC().Format("20060102T150405Z")
	for _, class := range course.Classes {
		if class == nil {
			continue
		}
		for _, event := range class.Events {
			if event == nil {
				continue
			}

			start, end, ok := parseEventDateRange(event.Date)
			if !ok {
				continue
			}

			b.WriteString("BEGIN:VEVENT\r\n")
			b.WriteString(foldICSLine("UID:" + safeICS(event.ExternalID) + "@tduex"))
			b.WriteString(foldICSLine("DTSTAMP:" + stamp))
			b.WriteString(foldICSLine("DTSTART;TZID=Asia/Tokyo:" + start.Format("20060102T150405")))
			b.WriteString(foldICSLine("DTEND;TZID=Asia/Tokyo:" + end.Format("20060102T150405")))
			b.WriteString(foldICSLine("SUMMARY:" + safeICS(class.Title+" - "+event.Name)))
			description := strings.Join([]string{
				"class: " + class.Title,
				"category: " + event.Category,
				"group: " + event.GroupName,
			}, "\\n")
			b.WriteString(foldICSLine("DESCRIPTION:" + safeICS(description)))
			b.WriteString(foldICSLine("CATEGORIES:" + safeICS(event.Category)))
			b.WriteString("END:VEVENT\r\n")
		}
	}

	b.WriteString("END:VCALENDAR\r\n")
	return writeTextFile(savePath, []byte(b.String()))
}

func writeCSVFile(savePath string, rows [][]string) error {
	if err := ensureParentDir(savePath); err != nil {
		return err
	}

	file, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.WriteAll(rows); err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

func writeXLSXFile(savePath string, sheetName string, rows [][]string) error {
	if err := ensureParentDir(savePath); err != nil {
		return err
	}

	file, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	if err := writeZipFile(zipWriter, "[Content_Types].xml", contentTypesXML); err != nil {
		return err
	}
	if err := writeZipFile(zipWriter, "_rels/.rels", relsXML); err != nil {
		return err
	}
	if err := writeZipFile(zipWriter, "xl/workbook.xml", workbookXML(sheetName)); err != nil {
		return err
	}
	if err := writeZipFile(zipWriter, "xl/_rels/workbook.xml.rels", workbookRelsXML); err != nil {
		return err
	}
	if err := writeZipFile(zipWriter, "xl/styles.xml", stylesXML); err != nil {
		return err
	}
	if err := writeZipFile(zipWriter, "xl/worksheets/sheet1.xml", worksheetXML(rows)); err != nil {
		return err
	}

	return zipWriter.Close()
}

func writeZipFile(zipWriter *zip.Writer, name string, content string) error {
	w, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

func worksheetXML(rows [][]string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for r, row := range rows {
		b.WriteString(fmt.Sprintf(`<row r="%d">`, r+1))
		for c, value := range row {
			cellRef := excelColumnName(c+1) + fmt.Sprintf("%d", r+1)
			b.WriteString(fmt.Sprintf(`<c r="%s" t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, cellRef, xmlEscape(value)))
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</sheetData></worksheet>`)
	return b.String()
}

func workbookXML(sheetName string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<sheets><sheet name="` + xmlEscape(sheetName) + `" sheetId="1" r:id="rId1"/></sheets></workbook>`
}

func excelColumnName(n int) string {
	result := ""
	for n > 0 {
		n--
		result = string(rune('A'+(n%26))) + result
		n /= 26
	}
	return result
}

func xmlEscape(value string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(value))
	return buf.String()
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func writeTextFile(savePath string, data []byte) error {
	if err := ensureParentDir(savePath); err != nil {
		return err
	}
	return os.WriteFile(savePath, data, 0644)
}

func parseEventDateRange(value string) (time.Time, time.Time, bool) {
	parts := strings.Split(value, " - ")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, false
	}

	const layout = "2006/01/02 15:04"
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		loc = time.FixedZone("JST", 9*60*60)
	}

	start, err := time.ParseInLocation(layout, strings.TrimSpace(parts[0]), loc)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	end, err := time.ParseInLocation(layout, strings.TrimSpace(parts[1]), loc)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

func safeICS(value string) string {
	replacer := strings.NewReplacer(
		`\\`, `\\\\`,
		";", `\;`,
		",", `\,`,
		"\r\n", `\n`,
		"\n", `\n`,
	)
	return replacer.Replace(value)
}

func foldICSLine(line string) string {
	const max = 73
	if len(line) <= max {
		return line + "\r\n"
	}

	var b strings.Builder
	for len(line) > max {
		b.WriteString(line[:max])
		b.WriteString("\r\n ")
		line = line[max:]
	}
	b.WriteString(line)
	b.WriteString("\r\n")
	return b.String()
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`

const relsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

const workbookRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

const stylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts>
  <fills count="1"><fill><patternFill patternType="none"/></fill></fills>
  <borders count="1"><border/></borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/></cellXfs>
  <cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`
