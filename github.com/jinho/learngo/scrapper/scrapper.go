package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type titles map[int]string

type extractedJob struct {
	id       string
	title    string
	company  string
	location string
	salary   string
	summary  string
}

func Scrape(term string) {
	start := time.Now()
	// url 설정, 첫번째 페이지
	baseURL := "https://kr.indeed.com/jobs?q=" + term + "&l&vjk=1015284880e2ff62"
	var totalJobs []extractedJob

	totalPage := getPages(baseURL)

	// 채널 생성
	pageChan := make(chan []extractedJob)

	for i := 0; i < totalPage; i++ {
		go getJobs(baseURL, i, pageChan)
	}

	for i := 0; i < totalPage; i++ {
		extractedJobs := <-pageChan
		totalJobs = append(totalJobs, extractedJobs...)
	}

	writeJobs(totalJobs)

	end := time.Now()
	fmt.Println("Done! 경과시간 : ", end.Sub(start))
}

func writeJobs(totalJobs []extractedJob) {
	// os 라이브러리 이용해서 csv 파일 생성
	file, err := os.Create("jobs.csv")
	checkErr(err)

	// 파일작성기 생성
	w := csv.NewWriter(file)

	// 끝날때 파일에 데이터 입력 - 모든 정보를 가져온 이후에 데이터 입력
	defer w.Flush()

	headers := []string{"ID", "TITLE", "COMPANY", "LOCATION", "SALARY", "SUMMARY"}

	// 파일에 헤더 쓰기 - 첫번째 행
	wErr := w.Write(headers)
	checkErr(wErr)

	c := make(chan []string)

	// 파일에 데이터 쓰기 - 다음 행
	for _, job := range totalJobs {
		go makeSlice(job, c)
	}

	for i := 0; i < len(totalJobs); i++ {
		jwErr := w.Write(<-c)
		checkErr(jwErr)
	}
}

func makeSlice(job extractedJob, c chan<- []string) {
	c <- []string{"https://kr.indeed.com/jobs?q=python&l&vjk=" + job.id, job.title, job.company, job.location, job.salary, job.summary}
}

func extractJob(item *goquery.Selection, c chan<- extractedJob) {
	id, _ := item.Attr("data-jk")
	title := item.Find(".jobTitle").Text()
	company := item.Find(".companyName").Text()
	location := item.Find(".companyLocation").Text()
	salary := item.Find(".salary-snippet").Text()
	summary := item.Find(".job-snippet").Text()

	c <- extractedJob{id: id, title: title, company: company, location: location, salary: salary, summary: summary}
}

func getJobs(baseURL string, page int, c chan<- []extractedJob) {
	pageURL := baseURL + "&start=" + strconv.Itoa(page*10)

	jobs := []extractedJob{}

	//  하나의 인덱스마다, 하나의 일자리

	// fmt.Println("**********************************************************")
	fmt.Println("Requesting : ", pageURL)
	// fmt.Println("\n**********************************************************")

	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	jobChan := make(chan extractedJob)

	searchItem := doc.Find(".sponTapItem")

	searchItem.Each(func(i int, item *goquery.Selection) {
		go extractJob(item, jobChan)
	})

	for i := 0; i < searchItem.Length(); i++ {
		jobs = append(jobs, <-jobChan)
	}

	c <- jobs
}

func getPages(url string) int {
	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	page := 0

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		page = s.Find(".pn").Length()
	})

	return page
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
}

func CleanString(str string) string {
	// TrimSpace - 문자열에서 공백 제거
	// Fields - 문자열에서 단어 별로 나누어 배열만들기
	// Join - 문자열 배열사이에 특정 문자를 넣어서 하나의 문자열 만들기
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
