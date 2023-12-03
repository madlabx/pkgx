package chartx

import (
	"github.com/stretchr/testify/require"
	"github.com/wcharczuk/go-chart/v2"
	"math/rand"
	"testing"
	"time"
)

func generateFloatItems(num int) []float64 {
	var arrayf []float64
	for i := 0; i < num; i++ {
		arrayf = append(arrayf, float64(10)*(rand.Float64()-0.5))
	}
	return arrayf
}

func TestDraw(t *testing.T) {
	floatArr1 := generateFloatItems(20)
	floatArr2 := generateFloatItems(20)
	x := []time.Time{
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
		time.Date(2021, time.January, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2021, time.January, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2021, time.January, 16, 0, 0, 0, 0, time.UTC),
		time.Date(2021, time.January, 17, 0, 0, 0, 0, time.UTC),
		time.Date(2021, time.January, 21, 0, 0, 0, 0, time.UTC),
		time.Date(2021, time.January, 22, 0, 0, 0, 0, time.UTC),
		time.Date(2021, time.January, 23, 0, 0, 0, 0, time.UTC),
	}

	err := DrawTimeSeries("test2", "tradeDate", x, "Y1", floatArr1, "output1.png")
	require.Nil(t, err)
	t.Log(err)

	err = DrawReturnTrend("test", "TradeDate",
		ChartYAxis{
			Name:     "Asset",
			Interval: 20000,
		}, ChartYAxis{
			Name:     "Percentage(%)",
			Interval: 20000,
		}, x, []ChartLine{
			{"Y1", floatArr1, chart.YAxisPrimary},
			{"Y2", floatArr2, chart.YAxisSecondary},
		}, 1440, 810, "output2.png")
	require.Nil(t, err)

}
