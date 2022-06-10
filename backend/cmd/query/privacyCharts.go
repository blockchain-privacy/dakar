package main

import (
	"backend/constants"
	dban "backend/db/analytics"
	"backend/external"

	"encoding/csv"
	"os"
	"time"
)

type privacyTypePair struct {
	label string
	start string
	stop  string
}

// exportTransactionData exports all transaction timestamps in a CSV-file per privacy type (mixing, destination, ...)
func exportTransactionData(database external.Database, directory string) {
	if len(directory) == 0 {
		info("invalid directory:", directory)
		return
	}

	privacyTypes := []privacyTypePair{
		{label: "mixing", start: "0", stop: constants.StrPrivacyMixingLast},
		// {label: "mixing 0", start: constants.StrPrivacyMixing0, stop: constants.StrPrivacyMixing0},
		// {label: "mixing 1", start: constants.StrPrivacyMixing1, stop: constants.StrPrivacyMixing1},
		// {label: "mixing 2", start: constants.StrPrivacyMixing2, stop: constants.StrPrivacyMixing2},
		// {label: "mixing 3", start: constants.StrPrivacyMixing3, stop: constants.StrPrivacyMixing3},
		// {label: "mixing 4", start: constants.StrPrivacyMixing4, stop: constants.StrPrivacyMixing4},
		{label: "origin", start: constants.StrPrivacyOriginFirst, stop: constants.StrPrivacyOriginLast},
		{label: "destination", start: constants.StrPrivacyDestinationFirst,
			stop: constants.StrPrivacyDestinationLast},
		{label: "collateral creation", start: constants.StrPrivacyCollateralCreationFirst,
			stop: constants.StrPrivacyCollateralCreationLast},
		{label: "collateral payment", start: constants.StrPrivacyCollateralPaymentFirst,
			stop: constants.StrPrivacyCollateralPaymentLast},
		{label: "all", start: "0", stop: "500"},
	}

	for _, privacyType := range privacyTypes {
		ts, dbErr := dban.GetPrivacyTypeData(database, privacyType.start, privacyType.stop)
		if dbErr != nil {
			info(dbErr)
			return
		}
		writeTimestampsToCSV(directory+"/"+privacyType.label, ts)
	}
}

func writeTimestampsToCSV(fileName string, txs []time.Time) {
	f, err := os.Create(fileName + ".csv")
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			info(err)
		}
	}(f)

	if err != nil {
		info(err)
		return
	}

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, t := range txs {
		line := []string{t.Format(time.RFC3339)}
		if err := w.Write(line); err != nil {
			info("error writing record to file", err)
			return
		}
	}
}

//
//func createPrivacyCharts(database external.Database, chartDir string) {
//	if len(chartDir) == 0 {
//		info("invalid chart directory:", chartDir)
//		return
//	}
//
//	privacyTypes := []privacyTypePair{
//		{label: "mixing", start: "0", stop: constants.StrPrivacyMixingLast},
//		// {label: "mixing 0", start: constants.StrPrivacyMixing0, stop: constants.StrPrivacyMixing0},
//		// {label: "mixing 1", start: constants.StrPrivacyMixing1, stop: constants.StrPrivacyMixing1},
//		// {label: "mixing 2", start: constants.StrPrivacyMixing2, stop: constants.StrPrivacyMixing2},
//		// {label: "mixing 3", start: constants.StrPrivacyMixing3, stop: constants.StrPrivacyMixing3},
//		// {label: "mixing 4", start: constants.StrPrivacyMixing4, stop: constants.StrPrivacyMixing4},
//		{label: "origin", start: constants.StrPrivacyOriginFirst, stop: constants.StrPrivacyOriginLast},
//		{label: "destination", start: constants.StrPrivacyDestinationFirst,
//			stop: constants.StrPrivacyDestinationLast},
//		{label: "collateral creation", start: constants.StrPrivacyCollateralCreationFirst,
//			stop: constants.StrPrivacyCollateralCreationLast},
//		{label: "collateral payment", start: constants.StrPrivacyCollateralPaymentFirst,
//			stop: constants.StrPrivacyCollateralPaymentLast},
//		{label: "all", start: "0", stop: "500"},
//	}
//
//	durations := []dur{
//		// {label: "block", d: 1},
//		{label: "day", d: time.Hour * 24, sma: true},
//		{label: "7 days", d: time.Hour * 24 * 7},
//	}
//
//	for _, privacyType := range privacyTypes {
//		ts, dbErr := dban.GetPrivacyTypeData(database, privacyType.start, privacyType.stop)
//		if dbErr != nil {
//			info(dbErr)
//			return
//		}
//
//		createCharts(durations, ts, chartDir, privacyType.label)
//	}
//}
//
//func findMaximum(data map[time.Time]int) (ts time.Time, maxVal int) {
//	for k, v := range data {
//		if v > maxVal {
//			ts = k
//			maxVal = v
//		}
//	}
//
//	return
//}
//
//func createCharts(durations []dur, ts []time.Time, dir string, privacyType string) {
//	if len(ts) == 0 {
//		return
//	}
//
//	for _, d := range durations {
//		timeMap := make(map[time.Time]int)
//		for _, t := range ts {
//			timeMap[t.Truncate(d.d)] = timeMap[t.Truncate(d.d)] + 1
//		}
//		maxTS, maxVal := findMaximum(timeMap)
//		info(privacyType, "maximum count:", maxTS, maxVal)
//
//		chartErr := drawChart(dir, timeMap, privacyType, d.label, d.sma)
//		if chartErr != nil {
//			info(chartErr)
//			return
//		}
//	}
//}
//
//func drawChart(dir string, data map[time.Time]int, chartName string, durationLabel string, sma bool) error {
//	if len(dir) == 0 {
//		return errors.New("chart directory is empty")
//	}
//
//	if len(chartName) == 0 {
//		return errors.New("chart name is empty")
//	}
//
//	file, err := os.Create(dir + "/" + chartName + "_" + durationLabel + ".png")
//	if err != nil {
//		return err
//	}
//
//	defer func() {
//		err = file.Close()
//		if err != nil {
//			panic(err)
//		}
//	}()
//
//	type dataPoint struct {
//		ts    time.Time
//		count int
//	}
//
//	timeData := make([]dataPoint, 0, len(data))
//
//	for k, v := range data {
//		timeData = append(timeData, dataPoint{
//			ts:    k,
//			count: v,
//		})
//	}
//
//	sort.Slice(timeData, func(i, j int) bool {
//		return timeData[i].ts.Before(timeData[j].ts)
//	})
//
//	var series chart.TimeSeries
//
//	for _, t := range timeData {
//		series.XValues = append(series.XValues, t.ts)
//		series.YValues = append(series.YValues, float64(t.count))
//	}
//
//	allSeries := []chart.Series{series}
//
//	if sma {
//		allSeries = append(allSeries, chart.SMASeries{
//			Style: chart.Style{
//				StrokeColor: drawing.ColorRed,
//				StrokeWidth: 2.0,
//			},
//			InnerSeries: series,
//		})
//	}
//
//	graph := chart.Chart{
//		Title:  "Number of '" + chartName + "' privacy transactions per " + durationLabel,
//		Height: 1080,
//		Width:  1920,
//		Series: allSeries,
//	}
//
//	return graph.Render(chart.PNG, file)
//}
