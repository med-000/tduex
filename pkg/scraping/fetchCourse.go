package scraping

import (
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

// FetchCourseHTML コース(講義全体)を取得
func (s *Scraper) FetchCourseHTML(c *colly.Collector, userId string, pass string, year int, term int) (string, error) {

	var html string
	var loggedIn bool

	loginURL := os.Getenv("LOGIN_URL")
	baseURL := os.Getenv("BASE_URL")

	//<scrip>タグを抽出(window.location.hrefから持ってきてる)
	redirectRe := regexp.MustCompile(`window\.location\.href="([^"]+)"`)

	//response来た時に呼び出される関数
	c.OnResponse(func(r *colly.Response) {
		body := string(r.Body)

		//requestのURLの中にlogin.phpがあったら（つまり一個前の通信がlogin.phpだったら）
		// login.phpのresponseのbodyからredirect処理を抜き出して
		if strings.Contains(r.Request.URL.String(), "login.php") {
			match := redirectRe.FindStringSubmatch(body)
			if len(match) > 1 {
				s.log.Info.Println("Get Redirect URL by login.php")
				next := baseURL + match[1]
				r.Request.Visit(next)
				s.log.Info.Printf("Visit loginURL")
				return
			}
		}

		//acs=がURLにあったら(つまり、ログインできて)responseが帰ってきたら
		if strings.Contains(r.Request.URL.String(), "acs_=") {
			loggedIn = true
			s.log.Info.Printf("Success Login")

			// 学期変更
			r.Request.Post(
				baseURL+"/webclass/index.php",
				map[string]string{
					"year":     strconv.Itoa(int(year)),
					"semester": strconv.Itoa(int(term)),
				},
			)
			s.log.Info.Printf("Success POST Year: %d Term: %d", year, term)
			return
		}

		//index.phpに入れたら
		if strings.Contains(r.Request.URL.String(), "index.php") {
			s.log.Info.Printf("Success courseHTML by index.php")
			html = body
		}
	})

	// STEP1
	if err := c.Visit(loginURL); err != nil {
		s.log.Error.Printf("Cannt visit %s", loginURL)
		return "", err
	}

	// STEP2
	if err := c.Post(loginURL, map[string]string{
		"username": userId,
		"val":      pass,
	}); err != nil {
		s.log.Error.Printf("Cannt post & login by %s", loginURL)
		return "", err
	}

	//メモリからuseIdなどをを消す
	defer func() {
		userId = ""
		pass = ""
	}()

	if !loggedIn || html == "" {
		s.log.Error.Printf("Don`t loggedIn or scraping is empty")
		return "", errors.New("failed to fetch scraping")
	}

	return html, nil
}
