package main

import (
	cli "backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	dban "backend/db/analytics"

	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

var thisLogger *log.Logger

func initLogger() {
	thisLogger = log.New(log.Writer(), "\033[0;31mquery\033[0m\t", log.Flags())
}
func info(v ...interface{}) {
	thisLogger.Println(v)
}

// setup cli
func getExplorerCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.Logfile, cli.DBPort, cli.DBHost, cli.ChartDir)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	if len(cliArgs.ChartDir) == 0 {
		flag.PrintDefaults()
		return cliArgs, errors.New("specify output directory for charts")
	}

	return cliArgs, err
}

type dur struct {
	label string
	d     time.Duration
	sma   bool
}

type privacyTypePair struct {
	label string
	start string
	stop  string
}

func main() {

	cliArgs, err := getExplorerCLIArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	// setup Logging
	if len(cliArgs.Logfile) > 0 {
		if f, err := cli.GetLogfile(cliArgs.Logfile); err == nil {
			defer func() {
				if err = f.Close(); err != nil {
					fmt.Println(err)
				}
			}()
		}
	}

	initLogger()

	// create dgraph client
	dgraph, c, err := db.CreateClient(cliArgs.DBEndpoint)
	if err != nil {
		info(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			info(err)
		}
	}()
	if len(cliArgs.ChartDir) > 0 {

		privacyTypes := []privacyTypePair{
			{label: "mixing", start: "0", stop: constants.StrPrivacyMixingLast},
			//{label: "mixing 0", start: constants.StrPrivacyMixing0, stop: constants.StrPrivacyMixing0},
			//{label: "mixing 1", start: constants.StrPrivacyMixing1, stop: constants.StrPrivacyMixing1},
			//{label: "mixing 2", start: constants.StrPrivacyMixing2, stop: constants.StrPrivacyMixing2},
			//{label: "mixing 3", start: constants.StrPrivacyMixing3, stop: constants.StrPrivacyMixing3},
			//{label: "mixing 4", start: constants.StrPrivacyMixing4, stop: constants.StrPrivacyMixing4},
			{label: "origin", start: constants.StrPrivacyOriginFirst, stop: constants.StrPrivacyOriginLast},
			{label: "destination", start: constants.StrPrivacyDestinationFirst,
				stop: constants.StrPrivacyDestinationLast},
			{label: "collateral creation", start: constants.StrPrivacyCollateralCreationFirst,
				stop: constants.StrPrivacyCollateralCreationLast},
			{label: "collateral payment", start: constants.StrPrivacyCollateralPaymentFirst,
				stop: constants.StrPrivacyCollateralPaymentLast},
			{label: "all", start: "0", stop: "500"},
		}

		durations := []dur{
			//{label: "block", d: 1},
			{label: "day", d: time.Hour * 24, sma: true},
			{label: "7 days", d: time.Hour * 24 * 7},
		}

		for _, privacyType := range privacyTypes {
			ts, dbErr := dban.GetPrivacyTypeData(dgraph, privacyType.start, privacyType.stop)
			if dbErr != nil {
				info(err)
				return
			}

			createCharts(durations, ts, cliArgs.ChartDir, privacyType.label)
		}
	}
}

func findMaximum(data map[time.Time]int) (ts time.Time, maxVal int) {
	for k, v := range data {
		if v > maxVal {
			ts = k
			maxVal = v
		}
	}

	return
}

func createCharts(durations []dur, ts []time.Time, dir string, privacyType string) {
	if len(ts) == 0 {
		return
	}

	for _, d := range durations {
		timeMap := make(map[time.Time]int)
		for _, t := range ts {
			timeMap[t.Truncate(d.d)] = timeMap[t.Truncate(d.d)] + 1
		}
		maxTs, maxVal := findMaximum(timeMap)
		info(privacyType, "maximum count:", maxTs, maxVal)

		chartErr := drawChart(dir, timeMap, privacyType, d.label, d.sma)
		if chartErr != nil {
			info(chartErr)
			return
		}
	}
}

func drawChart(dir string, data map[time.Time]int, chartName string, durationLabel string, sma bool) error {
	if len(dir) == 0 {
		return errors.New("chart directory is empty")
	}

	if len(chartName) == 0 {
		return errors.New("chart name is empty")
	}

	file, err := os.Create(dir + "/" + chartName + "_" + durationLabel + ".png")
	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}()

	type dataPoint struct {
		ts    time.Time
		count int
	}

	var timeData []dataPoint

	for k, v := range data {
		timeData = append(timeData, dataPoint{
			ts:    k,
			count: v,
		})
	}

	sort.Slice(timeData, func(i, j int) bool {
		return timeData[i].ts.Before(timeData[j].ts)
	})

	var series chart.TimeSeries

	for _, t := range timeData {
		series.XValues = append(series.XValues, t.ts)
		series.YValues = append(series.YValues, float64(t.count))
	}

	allSeries := []chart.Series{series}

	if sma {
		allSeries = append(allSeries, chart.SMASeries{
			Style: chart.Style{
				StrokeColor: drawing.ColorRed,
				StrokeWidth: 2.0,
			},
			InnerSeries: series,
		})
	}

	graph := chart.Chart{
		Title:  "Number of '" + chartName + "' privacy transactions per " + durationLabel,
		Height: 1080,
		Width:  1920,
		Series: allSeries,
	}

	return graph.Render(chart.PNG, file)
}
