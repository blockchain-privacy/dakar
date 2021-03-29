package main

import (
	cli "backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"backend/db/analytics"

	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/wcharczuk/go-chart/v2"
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

		privacyTypes := []string{constants.PrivacyMixing, constants.PrivacyOrigin,
			constants.PrivacyCollateralCreation, constants.PrivacyCollateralPayment,
			constants.PrivacyDestination, ""}

		for _, privacyType := range privacyTypes {
			ts, dbErr := analytics.GetPrivacyTypeData(dgraph, privacyType)
			if dbErr != nil {
				info(err)
				return
			}

			if len(privacyType) == 0 {
				privacyType = "all"
			}

			timeMap := make(map[time.Time]int)
			for _, t := range ts {
				timeMap[t] = timeMap[t] + 1
			}
			maxTs, maxVal := findMaximum(timeMap)
			info(privacyType, "maximum count:", maxTs, maxVal)

			chartErr := drawChart(cliArgs.ChartDir, timeMap, privacyType)
			if chartErr != nil {
				info(chartErr)
				return
			}
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

func drawChart(dir string, data map[time.Time]int, chartName string) error {
	if len(dir) == 0 {
		return errors.New("directory is empty")
	}

	if len(chartName) == 0 {
		return errors.New("chart name is empty")
	}

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

	graph := chart.Chart{
		Series: []chart.Series{},
	}

	var series chart.TimeSeries

	for _, t := range timeData {
		series.XValues = append(series.XValues, t.ts)
		series.YValues = append(series.YValues, float64(t.count))
	}

	graph.Series = append(graph.Series, series)

	file, err := os.Create(dir + "/" + chartName)
	if err != nil {
		return err
	}

	err = graph.Render(chart.PNG, file)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}
