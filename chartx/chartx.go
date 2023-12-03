package chartx

import (
	"errors"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"math"
	"math/rand"
	"os"
	"risk_manager/pkg/log"
	"time"

	"github.com/fogleman/gg"
	"github.com/wcharczuk/go-chart/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type ChartLine struct {
	Name  string
	YDate []float64
	YAxis chart.YAxisType
}

type ChartYAxis struct {
	Name     string
	Interval int
}

func createGraphs(title, xAxisName string, yAxisPri, yAxisSec ChartYAxis, X []time.Time, lines []ChartLine, width, height int) (*chart.Chart, error) {

	if len(X) == 0 || len(lines) == 0 {
		return nil, errors.New("EmptyX or EmptyY or not same number of YName and YData")
	}

	yPriMin, yPriMax, ySecMin, ySecMax := func() (float64, float64, float64, float64) {
		yPriMax := float64(-math.MaxInt64)
		ySecMax := float64(-math.MaxInt64)
		yPriMin := float64(math.MaxInt64)
		ySecMin := float64(math.MaxInt64)
		for _, line := range lines {
			if line.YAxis == chart.YAxisPrimary {
				for _, y := range line.YDate {
					yPriMin = math.Min(yPriMin, y)
					yPriMax = math.Max(yPriMax, y)
				}
			} else if line.YAxis == chart.YAxisSecondary {
				for _, y := range line.YDate {
					ySecMin = math.Min(ySecMin, y)
					ySecMax = math.Max(ySecMax, y)
				}
			}
		}

		return yPriMin, yPriMax, ySecMin, ySecMax
	}()

	_ = func() (ticks []chart.Tick) {
		var i int
		if ySecMax > 0 {
			for i = 0; float64(yAxisSec.Interval*i) < ySecMax; i++ {

				ticks = append(ticks, chart.Tick{
					Value: float64(yAxisSec.Interval * i),
					Label: fmt.Sprintf("%d", yAxisSec.Interval*i),
				})
			}
			ticks = append(ticks, chart.Tick{
				Value: float64(yAxisSec.Interval * i),
				Label: fmt.Sprintf("%d", yAxisSec.Interval*i),
			})
		}

		if ySecMin < 0 {
			for i = 0; float64(-yAxisSec.Interval*i+yAxisSec.Interval) > ySecMin; i++ {

				ticks = append(ticks, chart.Tick{
					Value: float64(-yAxisSec.Interval * i),
					Label: fmt.Sprintf("%d", -yAxisSec.Interval*i),
				})
			}
			ticks = append(ticks, chart.Tick{
				Value: float64(-yAxisSec.Interval * i),
				Label: fmt.Sprintf("%d", -yAxisSec.Interval*i),
			})
		}
		return
	}()

	graph := chart.Chart{
		Width:  width,
		Height: height,
		Title:  title,
		XAxis: chart.XAxis{
			Name:           xAxisName,
			ValueFormatter: chart.TimeValueFormatterWithFormat("2006-01-02"),
			//TickPosition:   chart.TickPositionUnset, // 将 X 轴放在 Y 轴为 0 的位置
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    50,
				Left:   25,
				Right:  25,
				Bottom: 20,
			},
			FillColor: drawing.ColorFromHex("efefef"),
		},
		YAxis: chart.YAxis{
			//Zero: chart.GridLine{Value: 0},
			Name: yAxisPri.Name,
			Range: &chart.ContinuousRange{
				Min: yPriMin,
				Max: yPriMax,
			},

			Ticks: func() (ticks []chart.Tick) {
				var i int
				if yPriMax > 0 {
					for i = 0; float64(yAxisPri.Interval*i) < yPriMax; i++ {

						ticks = append(ticks, chart.Tick{
							Value: float64(yAxisPri.Interval * i),
							Label: fmt.Sprintf("%d", yAxisPri.Interval*i),
						})
					}
					ticks = append(ticks, chart.Tick{
						Value: float64(yAxisPri.Interval * i),
						Label: fmt.Sprintf("%d", yAxisPri.Interval*i),
					})
				}

				if yPriMin < 0 {
					for i = 0; float64(-yAxisPri.Interval*i+yAxisPri.Interval) > yPriMin; i++ {

						ticks = append(ticks, chart.Tick{
							Value: float64(-yAxisPri.Interval * i),
							Label: fmt.Sprintf("%d", -yAxisPri.Interval*i),
						})
					}
					ticks = append(ticks, chart.Tick{
						Value: float64(-yAxisPri.Interval * i),
						Label: fmt.Sprintf("%d", -yAxisPri.Interval*i),
					})
				}
				return
			}(),
		},
		YAxisSecondary: chart.YAxis{
			Name: yAxisSec.Name,
			Range: &chart.ContinuousRange{
				Min: math.Ceil((ySecMin-float64(yAxisSec.Interval))/float64(yAxisSec.Interval)) * float64(yAxisSec.Interval),
				Max: math.Ceil((3*ySecMax-2*ySecMin)/float64(yAxisSec.Interval)) * float64(yAxisSec.Interval),
			},
			//
			//Ticks: ticks,
		},
	}

	var mainSeries chart.TimeSeries
	for i, line := range lines {
		series := chart.TimeSeries{
			Name:    line.Name,
			YAxis:   line.YAxis,
			XValues: X,
			YValues: line.YDate,
			Style: chart.Style{
				StrokeColor: getRandomColor(i),
				DotColor:    getRandomColor(i),
				DotWidth:    2,
				//StrokeWidth: 1,
				//DotColor:    getRandomColor(1),
				//DotWidth:    2s
			},
		}

		if line.YAxis == chart.YAxisPrimary {
			series.Style.StrokeWidth = 1
		} else {
			series.Style.StrokeDashArray = []float64{2.5, 2.5}
		}
		graph.Series = append(graph.Series, series)
		if i == 0 {
			mainSeries = series
		}
	}

	maxSeries := &chart.MaxSeries{
		Style: chart.Style{
			StrokeColor:     chart.ColorAlternateGray,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		InnerSeries: mainSeries,
	}
	minSeries := &chart.MinSeries{
		Style: chart.Style{
			StrokeColor:     chart.ColorAlternateGray,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		InnerSeries: mainSeries,
	}
	graph.Series = append(graph.Series, maxSeries)
	graph.Series = append(graph.Series, minSeries)
	graph.Series = append(graph.Series, chart.LastValueAnnotationSeries(maxSeries))
	graph.Series = append(graph.Series, chart.LastValueAnnotationSeries(minSeries))

	graph.Elements = []chart.Renderable{chart.Legend(&graph)}

	return &graph, nil
}

// 获取随机颜色
func getRandomColor(index int) drawing.Color {
	colors := []drawing.Color{
		chart.ColorRed,
		chart.ColorBlue,
		chart.ColorGreen,
		chart.ColorOrange,
		chart.ColorLightGray,
		// 还可以添加更多颜色
	}

	return colors[index%len(colors)]
}

func createGraph(title string, XName string, X []time.Time, YName string, Y []float64, width, height int) (*chart.Chart, error) {
	if len(X) == 0 || len(Y) == 0 {
		return nil, errors.New("EmptyX or EmptyY")
	}
	graph := chart.Chart{
		Width:  width,
		Height: height,
		Title:  title,
		XAxis: chart.XAxis{
			Name:           XName,
			ValueFormatter: chart.TimeValueFormatterWithFormat("2006-01-02"),
		},
		YAxis: chart.YAxis{
			Name: YName,
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "Series 1",
				Style: chart.Style{
					StrokeColor: chart.ColorRed, // 设置线条颜色为红色
					StrokeWidth: 1,              // 设置线条宽度为2.5像素
					DotColor:    chart.ColorRed,
					DotWidth:    2,
				},
				XValues: X,
				YValues: Y,
			},
		},
	}

	return &graph, nil
}

func DrawReturnTrend(title, xAxisName string, yAxisPri, yAxisSec ChartYAxis, X []time.Time, lines []ChartLine, width, height int, chartPngName string) error {
	graph, err := createGraphs(title, xAxisName, yAxisPri, yAxisSec, X, lines, width, height)
	if err != nil {
		return err
	}

	// 创建输出文件
	file, err := os.Create(chartPngName)
	if err != nil {
		return err
	}
	defer file.Close()

	// 绘制图表并保存为图片
	err = graph.Render(chart.PNG, file)
	if err != nil {
		return err
	}

	return nil
}

func DrawTimeSeries(title string, XName string, X []time.Time, YName string, Y []float64, chartPngName string) error {
	if len(X) == 0 || len(Y) == 0 {
		return errors.New("EmptyX or EmptyY")
	}
	graph, err := createGraph(title, XName, X, YName, Y, 1200, 800)

	// 创建输出文件
	file, err := os.Create(chartPngName)
	if err != nil {
		return err
	}
	defer file.Close()

	// 绘制图表并保存为图片
	err = graph.Render(chart.PNG, file)
	if err != nil {
		return err
	}

	return nil
}

func chart_draw() {
	// 创建图表对象
	graph := chart.Chart{
		Width:  1200,
		Height: 600,
		Title:  "Trend Chart",
		XAxis: chart.XAxis{
			Name:           "X",
			ValueFormatter: chart.TimeValueFormatterWithFormat("2006-01-02"),
		},
		YAxis: chart.YAxis{
			Name: "Y",
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "Series 1",
				Style: chart.Style{
					StrokeColor: chart.ColorRed, // 设置线条颜色为红色
					StrokeWidth: 1,              // 设置线条宽度为2.5像素
					DotColor:    chart.ColorRed,
					DotWidth:    2,
				},
				XValues: []time.Time{
					time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2021, time.January, 23, 0, 0, 0, 0, time.UTC),
				},
				YValues: []float64{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			},
		},
	}

	// 创建输出文件
	file, err := os.Create("chart.png")
	if err != nil {
		panic(err)
	}

	// 绘制图表并保存为图片
	err = graph.Render(chart.PNG, file)
	if err != nil {
		panic(err)
	}

	// 关闭文件
	file.Close()
}

//
//func drawChart() {
//	xv, yv := xvalues(), yvalues()
//
//	priceSeries := chart.TimeSeries{
//		Name: "SPY",
//		Style: chart.Style{
//			Show:        true,
//			StrokeColor: chart.GetDefaultColor(0),
//		},
//		XValues: xv,
//		YValues: yv,
//	}
//
//	smaSeries := chart.SMASeries{
//		Name: "SPY - SMA",
//		Style: chart.Style{
//			Show:            true,
//			StrokeColor:     drawing.ColorRed,
//			StrokeDashArray: []float64{5.0, 5.0},
//		},
//		InnerSeries: priceSeries,
//	}
//
//	bbSeries := &chart.BollingerBandsSeries{
//		Name: "SPY - Bol. Bands",
//		Style: chart.Style{
//			Show:        true,
//			StrokeColor: drawing.ColorFromHex("efefef"),
//			FillColor:   drawing.ColorFromHex("efefef").WithAlpha(64),
//		},
//		InnerSeries: priceSeries,
//	}
//
//	graph := chart.Chart{
//		XAxis: chart.XAxis{
//			Style:        chart.Style{Show: true},
//			TickPosition: chart.TickPositionBetweenTicks,
//		},
//		YAxis: chart.YAxis{
//			Style: chart.Style{Show: true},
//			Range: &chart.ContinuousRange{
//				Max: 220.0,
//				Min: 180.0,
//			},
//		},
//		Series: []chart.Series{
//			bbSeries,
//			priceSeries,
//			smaSeries,
//		},
//	}
//	file, err := os.Create("chart.png")
//	if err != nil {
//		fmt.Println("Error creating PNG:", err)
//		os.Exit(1)
//	}
//	graph.Render(chart.PNG, file)
//}

func xvalues() []time.Time {
	rawx := []string{"2015-07-17", "2015-07-20", "2015-07-21", "2015-07-22", "2015-07-23", "2015-07-24", "2015-07-27", "2015-07-28", "2015-07-29", "2015-07-30", "2015-07-31", "2015-08-03", "2015-08-04", "2015-08-05", "2015-08-06", "2015-08-07", "2015-08-10", "2015-08-11", "2015-08-12", "2015-08-13", "2015-08-14", "2015-08-17", "2015-08-18", "2015-08-19", "2015-08-20", "2015-08-21", "2015-08-24", "2015-08-25", "2015-08-26", "2015-08-27", "2015-08-28", "2015-08-31", "2015-09-01", "2015-09-02", "2015-09-03", "2015-09-04", "2015-09-08", "2015-09-09", "2015-09-10", "2015-09-11", "2015-09-14", "2015-09-15", "2015-09-16", "2015-09-17", "2015-09-18", "2015-09-21", "2015-09-22", "2015-09-23", "2015-09-24", "2015-09-25", "2015-09-28", "2015-09-29", "2015-09-30", "2015-10-01", "2015-10-02", "2015-10-05", "2015-10-06", "2015-10-07", "2015-10-08", "2015-10-09", "2015-10-12", "2015-10-13", "2015-10-14", "2015-10-15", "2015-10-16", "2015-10-19", "2015-10-20", "2015-10-21", "2015-10-22", "2015-10-23", "2015-10-26", "2015-10-27", "2015-10-28", "2015-10-29", "2015-10-30", "2015-11-02", "2015-11-03", "2015-11-04", "2015-11-05", "2015-11-06", "2015-11-09", "2015-11-10", "2015-11-11", "2015-11-12", "2015-11-13", "2015-11-16", "2015-11-17", "2015-11-18", "2015-11-19", "2015-11-20", "2015-11-23", "2015-11-24", "2015-11-25", "2015-11-27", "2015-11-30", "2015-12-01", "2015-12-02", "2015-12-03", "2015-12-04", "2015-12-07", "2015-12-08", "2015-12-09", "2015-12-10", "2015-12-11", "2015-12-14", "2015-12-15", "2015-12-16", "2015-12-17", "2015-12-18", "2015-12-21", "2015-12-22", "2015-12-23", "2015-12-24", "2015-12-28", "2015-12-29", "2015-12-30", "2015-12-31", "2016-01-04", "2016-01-05", "2016-01-06", "2016-01-07", "2016-01-08", "2016-01-11", "2016-01-12", "2016-01-13", "2016-01-14", "2016-01-15", "2016-01-19", "2016-01-20", "2016-01-21", "2016-01-22", "2016-01-25", "2016-01-26", "2016-01-27", "2016-01-28", "2016-01-29", "2016-02-01", "2016-02-02", "2016-02-03", "2016-02-04", "2016-02-05", "2016-02-08", "2016-02-09", "2016-02-10", "2016-02-11", "2016-02-12", "2016-02-16", "2016-02-17", "2016-02-18", "2016-02-19", "2016-02-22", "2016-02-23", "2016-02-24", "2016-02-25", "2016-02-26", "2016-02-29", "2016-03-01", "2016-03-02", "2016-03-03", "2016-03-04", "2016-03-07", "2016-03-08", "2016-03-09", "2016-03-10", "2016-03-11", "2016-03-14", "2016-03-15", "2016-03-16", "2016-03-17", "2016-03-18", "2016-03-21", "2016-03-22", "2016-03-23", "2016-03-24", "2016-03-28", "2016-03-29", "2016-03-30", "2016-03-31", "2016-04-01", "2016-04-04", "2016-04-05", "2016-04-06", "2016-04-07", "2016-04-08", "2016-04-11", "2016-04-12", "2016-04-13", "2016-04-14", "2016-04-15", "2016-04-18", "2016-04-19", "2016-04-20", "2016-04-21", "2016-04-22", "2016-04-25", "2016-04-26", "2016-04-27", "2016-04-28", "2016-04-29", "2016-05-02", "2016-05-03", "2016-05-04", "2016-05-05", "2016-05-06", "2016-05-09", "2016-05-10", "2016-05-11", "2016-05-12", "2016-05-13", "2016-05-16", "2016-05-17", "2016-05-18", "2016-05-19", "2016-05-20", "2016-05-23", "2016-05-24", "2016-05-25", "2016-05-26", "2016-05-27", "2016-05-31", "2016-06-01", "2016-06-02", "2016-06-03", "2016-06-06", "2016-06-07", "2016-06-08", "2016-06-09", "2016-06-10", "2016-06-13", "2016-06-14", "2016-06-15", "2016-06-16", "2016-06-17", "2016-06-20", "2016-06-21", "2016-06-22", "2016-06-23", "2016-06-24", "2016-06-27", "2016-06-28", "2016-06-29", "2016-06-30", "2016-07-01", "2016-07-05", "2016-07-06", "2016-07-07", "2016-07-08", "2016-07-11", "2016-07-12", "2016-07-13", "2016-07-14", "2016-07-15"}

	var dates []time.Time
	for _, ts := range rawx {
		parsed, _ := time.Parse(chart.DefaultDateFormat, ts)
		dates = append(dates, parsed)
	}
	return dates
}

func yvalues() []float64 {
	return []float64{212.47, 212.59, 211.76, 211.37, 210.18, 208.00, 206.79, 209.33, 210.77, 210.82, 210.50, 209.79, 209.38, 210.07, 208.35, 207.95, 210.57, 208.66, 208.92, 208.66, 209.42, 210.59, 209.98, 208.32, 203.97, 197.83, 189.50, 187.27, 194.46, 199.27, 199.28, 197.67, 191.77, 195.41, 195.55, 192.59, 197.43, 194.79, 195.85, 196.74, 196.01, 198.45, 200.18, 199.73, 195.45, 196.46, 193.90, 193.60, 192.90, 192.87, 188.01, 188.12, 191.63, 192.13, 195.00, 198.47, 197.79, 199.41, 201.21, 201.33, 201.52, 200.25, 199.29, 202.35, 203.27, 203.37, 203.11, 201.85, 205.26, 207.51, 207.00, 206.60, 208.95, 208.83, 207.93, 210.39, 211.00, 210.36, 210.15, 210.04, 208.08, 208.56, 207.74, 204.84, 202.54, 205.62, 205.47, 208.73, 208.55, 209.31, 209.07, 209.35, 209.32, 209.56, 208.69, 210.68, 208.53, 205.61, 209.62, 208.35, 206.95, 205.34, 205.87, 201.88, 202.90, 205.03, 208.03, 204.86, 200.02, 201.67, 203.50, 206.02, 205.68, 205.21, 207.40, 205.93, 203.87, 201.02, 201.36, 198.82, 194.05, 191.92, 192.11, 193.66, 188.83, 191.93, 187.81, 188.06, 185.65, 186.69, 190.52, 187.64, 190.20, 188.13, 189.11, 193.72, 193.65, 190.16, 191.30, 191.60, 187.95, 185.42, 185.43, 185.27, 182.86, 186.63, 189.78, 192.88, 192.09, 192.00, 194.78, 192.32, 193.20, 195.54, 195.09, 193.56, 198.11, 199.00, 199.78, 200.43, 200.59, 198.40, 199.38, 199.54, 202.76, 202.50, 202.17, 203.34, 204.63, 204.38, 204.67, 204.56, 203.21, 203.12, 203.24, 205.12, 206.02, 205.52, 206.92, 206.25, 204.19, 206.42, 203.95, 204.50, 204.02, 205.92, 208.00, 208.01, 207.78, 209.24, 209.90, 210.10, 208.97, 208.97, 208.61, 208.92, 209.35, 207.45, 206.33, 207.97, 206.16, 205.01, 204.97, 205.72, 205.89, 208.45, 206.50, 206.56, 204.76, 206.78, 204.85, 204.91, 204.20, 205.49, 205.21, 207.87, 209.28, 209.34, 210.24, 209.84, 210.27, 210.91, 210.28, 211.35, 211.68, 212.37, 212.08, 210.07, 208.45, 208.04, 207.75, 208.37, 206.52, 207.85, 208.44, 208.10, 210.81, 203.24, 199.60, 203.20, 206.66, 209.48, 209.92, 208.41, 209.66, 209.53, 212.65, 213.40, 214.95, 214.92, 216.12, 215.83}
}

func draw2() {
	// 创建一个新的图表
	p := plot.New()
	p.Title.Text = "折线图"
	p.X.Label.Text = "X轴"
	p.Y.Label.Text = "Y轴"

	// 示例数据
	xValues := []float64{0, 1, 2, 3, 4, 5, 6}
	yValues := []float64{120, 200, 150, 80, 70, 110, 130}

	// 创建数据点
	points := make(plotter.XYs, len(xValues))
	for i := range points {
		points[i].X = xValues[i]
		points[i].Y = yValues[i]
	}

	// 绘制折线图
	err := plotutil.AddLinePoints(p, "数据", points)
	if err != nil {
		log.Fatal(err)
	}

	err = p.Save(4*vg.Inch, 4*vg.Inch, "chart2.png")
	if err != nil {
		log.Fatal(err)
	}
}

func gg_draw() {
	width := 800
	height := 400

	dc := gg.NewContext(width, height)

	// 设置绘图属性
	dc.SetRGB(0, 0, 0) // 设置线条颜色为黑色
	dc.SetLineWidth(2) // 设置线条宽度为2

	// 定义折线的坐标点
	points := []struct {
		X, Y float64
	}{
		{100, 100},
		{200, 300},
		{300, 200},
		{400, 250},
		{500, 150},
	}

	// 绘制折线路径
	dc.MoveTo(points[0].X, points[0].Y)
	for _, point := range points[1:] {
		dc.LineTo(point.X, point.Y)
	}

	// 绘制折线
	dc.Stroke()

	// 保存绘制结果为图片
	dc.SavePNG("output.png")
}

func generateBarItems() []opts.BarData {
	items := make([]opts.BarData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.BarData{Value: rand.Intn(300)})
	}
	return items
}

func echats_draw() {
	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "My first bar chart generated by go-echarts",
		Subtitle: "It's extremely easy to use, right?",
	}))

	// Put data into instance
	bar.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		AddSeries("Category A", generateBarItems()).
		AddSeries("Category B", generateBarItems())
	// Where the magic happens
	f, _ := os.Create("bar.html")
	bar.Render(f)
}
