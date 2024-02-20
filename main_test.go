package main

import (
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

var maxFactorial = 500

type Report struct {
	Overflow int
	Success  int
	Failed   int
}

func generateBarItems(report Report) []opts.BarData {
	items := make([]opts.BarData, 0)
	items = append(items, opts.BarData{Value: report.Success})
	items = append(items, opts.BarData{Value: report.Failed})
	items = append(items, opts.BarData{Value: report.Overflow})
	return items
}

func init() {
	testMode = true
}

func TestFactorialNoMemo(t *testing.T) {
	app := Application{}
	app.serve()
	server := httptest.NewServer(app.server)
	defer func(app *Application) {
		app.quit <- syscall.SIGINT
		server.Close()
	}(&app)

	rpt := Report{}
	for i := 0; i < maxFactorial; i++ {

		res, err := http.Get(fmt.Sprintf("%s/factorial-no-memo?n=%d", server.URL, i))
		if err != nil {
			rpt.Failed++
			continue
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			rpt.Failed++
			continue
		}

		data := strings.Trim(string(body), "\n")
		expected := app.factorialBig(big.NewInt(int64(i))).String()
		if data != expected {
			if strings.HasPrefix(data, "-") || data == "0" {
				rpt.Overflow++
			}
			continue
		}
		rpt.Success++
	}

	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "factorial with no memoization",
		Subtitle: "Benckmark for success, overflow and failure",
	}))

	// Put data into instance
	bar.SetXAxis([]string{"Success", "Failure", "Overflow"}).
		AddSeries("report", generateBarItems(rpt), charts.WithBar3DChartOpts(opts.Bar3DChart{}))
	// Where the magic happens
	f, _ := os.Create("report-fact-with-memo.html")
	bar.Render(f)

}

func TestFactorial(t *testing.T) {
	app := Application{}
	app.serve()
	server := httptest.NewServer(app.server)
	defer func(app *Application) {
		app.quit <- syscall.SIGINT
		server.Close()
	}(&app)

	rpt := Report{}
	for i := 0; i < maxFactorial; i++ {

		res, err := http.Get(fmt.Sprintf("%s/factorial?n=%d", server.URL, i))
		if err != nil {
			log.Println(err)
			rpt.Failed++
			continue
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			rpt.Failed++
			continue
		}

		data := strings.Trim(string(body), "\n")
		expected := app.factorialBig(big.NewInt(int64(i))).String()
		if data != expected {
			if strings.HasPrefix(data, "-") || data == "0" {
				rpt.Overflow++
			}
			continue
		}
		rpt.Success++
	}

	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "factorial memoized with interger",
		Subtitle: "Benckmark for success, overflow and failure",
	}))

	// Put data into instance
	bar.SetXAxis([]string{"Success", "Failure", "Overflow"}).
		AddSeries("report", generateBarItems(rpt))
	// Where the magic happens
	f, _ := os.Create("report-fact-with-memo.html")
	bar.Render(f)

}

func TestFactorialWithBigInteger(t *testing.T) {
	app := Application{}
	app.serve()
	server := httptest.NewServer(app.server)
	defer func(app *Application) {
		app.quit <- syscall.SIGINT
		server.Close()
	}(&app)

	rpt := Report{}
	for i := 0; i < maxFactorial; i++ {

		res, err := http.Get(fmt.Sprintf("%s/factorial?n=%d", server.URL, i))
		if err != nil {
			log.Println(err)
			rpt.Failed++
			continue
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			rpt.Failed++
			continue
		}

		data := string(body)
		if strings.HasPrefix(data, "-") {
			rpt.Overflow++
			continue
		}
		rpt.Success++
	}

	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "factorial memoized with big interger",
		Subtitle: "Benckmark for success, overflow and failure",
	}))

	// Put data into instance
	bar.SetXAxis([]string{"Success", "Failure", "Overflow"}).
		AddSeries("report", generateBarItems(rpt))
	// Where the magic happens
	f, _ := os.Create("report-fact-with-big.html")
	bar.Render(f)

}
