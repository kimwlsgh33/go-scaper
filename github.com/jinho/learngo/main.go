package main

import (
	"os"
	"strings"

	"github.com/jinho/learngo/scrapper"

	"github.com/labstack/echo/v4"
)

const FILE_NAME string = "jobs.csv"

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	// 사용자에게 csv 파일 전달되면, 서버 데이터 삭제
	defer os.Remove(FILE_NAME)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrape(term)
	// 현재 경로의 jobs.csv 파일을 찾아서 사용자에게 jobs.csv라는 파일로 전해준다.
	return c.Attachment(FILE_NAME, FILE_NAME)
}

func checkErr(err error) {
  if err != nil {
    panic(err)
  }
}

// Language: go
// Path: src/github.com/jinho/learngo/scrapper/scrapper.go
package scrapper

import (
  "fmt"
  "log"
  "os"
  "strings"
    
  "github.com/gocolly/colly"
  "github.com/gocolly/colly/queue"
)

// CleanString cleans a string
func CleanString(str string) string {
  return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

// Scrape Indeed by a term
func Scrape(term string) {
  c := colly.NewCollector(
    // Visit only domains: indeed.com, www.indeed.com
    colly.AllowedDomains("indeed.com", "www.indeed.com"),
  )

  // Before making a request print "Visiting ..."
  c.OnRequest(func(r *colly.Request) {
    fmt.Println("Visiting", r.URL.String())
  })

  // On every a element which has href attribute call callback
  c.OnHTML("a[href]", func(e *colly.HTMLElement) {
    // Link found
    // Print link
    link := e.Attr("href")
    fmt.Printf("Link found: %q -> %s

", e.Text, link)
    // Visit link found on page on a new thread
    go e.Request.Visit(link)
  })

  // Start scraping on https://hackerspaces.org
  c.Visit("https://www.indeed.com/jobs?q=python&limit=50")
}

// Scrape indeed by a term
func Scrape2(term string) {
  // Instantiate default collector
  c := colly.NewCollector(
    // Visit only domains: indeed.com, www.indeed.com
    colly.AllowedDomains("indeed.com", "www.indeed.com"),
  )

  // Instantiate the queue with 2 consumer threads
  q, _ := queue.New(
    2, // Number of consumer threads
    &queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
  )

  // Before making a request print "Visiting ..."
  c.OnRequest(func(r *colly.Request) {
    fmt.Println("Visiting", r.URL.String())
  })

  // On every a element which has href attribute call callback
  c.OnHTML("a[href]", func(e *colly.HTMLElement) {
    // Link found
    // Print link
    link := e.Attr("href")
    fmt.Printf("Link found: %q -> %s

", e.Text, link)
    // Check if the href attribute contains "topic" or "user"
    // If true add the link to the queue
    if strings.Contains(link, "topic") || strings.Contains(link, "user") {
      q.AddURL(link)
    }
  })

  // Consume URLs
  q.Run(c)
}

// Scrape indeed by a term

