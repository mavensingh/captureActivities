package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/joho/godotenv"
)

// func executeCronJob() {
// 	gocron.Every(1).Seconds().Do(CaptureScreen)
// 	<-gocron.Start()
// }

func createRepo() string {
	name := "storage"
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(name, 0755)
		if errDir != nil {
			log.Println(err)
		}

	}
	return name
}

func generateFullFilePath(filepath, filename string) string {
	return fmt.Sprintf("%s/%s", filepath, filename)
}

func remove(filepath string) {
	err := os.Remove(filepath)
	if err != nil {
		log.Println("error to remove file/directory")
		return
	}
}

func handleForcecontrol(folderName string) {
	ctx := context.Background()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()

	}()
	go func() {
		select {
		case <-c:
			cancel()
			remove(folderName)
			os.Exit(1)
		case <-ctx.Done():
		}
	}()
}

func runContinueslyScreenShots(folderName string) {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				//Call the periodic function here.
				fileName := CaptureScreen()
				mailsPrepared(folderName, fileName)
				remove(generateFullFilePath(folderName, fileName))
			}
		}
	}()
	// quit := make(chan bool, 1)
	// // main will continue to wait until there is an entry in quit
	// <-quit

}

func runContinueslyclicksCapture(db *sql.DB) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				//Call the periodic function here.
				duration := captureClicks()
				getUsedProduct()
				uploadClick(db, duration)
				// go getUsedProduct()
			}
		}
	}()

}

// func capturerunningProcess() {
// 	ticker := time.NewTicker(1 * time.Second)
// 	go func() {
// 		for {
// 			select {
// 			case <-ticker.C:
// 				// Call the periodic function here.
// 				getUsedProduct()

// 			}
// 		}
// 	}()
// }

// generate random data for line chart
func generateBarItems(length *[]DailyGraph) []opts.LineData {
	items := make([]opts.LineData, 0)
	for _, item := range *length {
		items = append(items, opts.LineData{Value: item.Count, Name: item.Product})
	}
	return items
}

func GetDays(c *gin.Context) {
	dbType := "mysql"
	db := loadEnv().connect(dbType)
	length := getUsedProductPerDays(db)
	db.Close()
	titles := []string{}
	for _, title := range *length {
		titles = append(titles, title.Days)
	}
	bar := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "My first bar chart generated by go-echarts",
		Subtitle: "It's extremely easy to use, right?",
	}))
	bar.Tooltip.Show = true
	// Put data into instance
	bar.SetXAxis(titles).
		AddSeries("Days", generateBarItems(length))
	// Where the magic happens

	bar.Render(c.Writer)
}

func Daily(c *gin.Context) {
	dbType := "mysql"
	db := loadEnv().connect(dbType)
	length := getUsedProductPerDay(db)
	db.Close()
	titles := []string{}
	for _, title := range *length {
		titles = append(titles, title.Product)
	}
	// log.Println(titles)
	bar := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "My first bar chart generated by go-echarts",
		Subtitle: "It's extremely easy to use, right?",
	}))
	bar.Tooltip.Show = true
	// bar.Tooltip.Formatter = fmt.Sprintf("daily based code\n product: count:")
	// Put data into instance
	bar.SetXAxis(titles).
		AddSeries("Today's clicks", generateBarItems(length))

	bar.Render(c.Writer)
}

func main() {
	// go executeCronJob()
	// CaptureScreen()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file -> ", err)
		return
	}
	folderName := createRepo()
	// log.Println(folderName)
	handleForcecontrol(folderName)
	now := time.Now()
	hour := now.Hour()
	dbType := "mysql"
	db := loadEnv().connect(dbType)
	if hour >= 10 && hour < 18 {
		// fmt.Println("Running..")
		log.Println("ENABLED")
		// capturerunningProcess()
		// getUsedProduct()
		runContinueslyclicksCapture(db)
		runContinueslyScreenShots(folderName)

	} else {
		// fmt.Println("sleeping..")
		// log.Println("sleeping for ", time.Hour)
		// time.Sleep(16 + time.Hour)
		// log.Println("Capture screen Method are disabled until next office time.")
		log.Println("DISABLED")
		runContinueslyclicksCapture(db)
	}

	//setup routes
	r := SetupRouter()

	// running
	r.Run(":8000")

}
