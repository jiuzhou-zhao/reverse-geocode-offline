package reverse_geocode_offline

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func randFloats(min, max float64, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = min + rand.Float64()*(max-min)
	}
	return res
}

func BenchmarkQueryCountry(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	dataProvider, err := NewGeoDataProvider("", []string{
		"./data/level2/earth_l2.geojson.gz",
	})
	assert.Nil(b, err)
	longitudes := randFloats(-180, 180, b.N)
	latitudes := randFloats(-90, 90, b.N)
	b.ResetTimer()

	trueResultCnt, falseResultCnt := 0, 0
	for i := 0; i < b.N; i++ {
		id := dataProvider.Contains(longitudes[i], latitudes[i])
		if id == 0 {
			falseResultCnt++
		} else {
			trueResultCnt++
		}
	}
	b.Logf("%v - %v\n", trueResultCnt, falseResultCnt)
}

func BenchmarkQueryProvince(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	dataProvider, err := NewGeoDataProvider("", []string{
		"./data/level3-4/1de31a0b3f4cd12f023cf3116a95995f.geojson.gz",
	})
	assert.Nil(b, err)
	longitudes := randFloats(73.33, 135.05, b.N)
	latitudes := randFloats(3.51, 53.33, b.N)
	b.ResetTimer()

	trueResultCnt, falseResultCnt := 0, 0
	for i := 0; i < b.N; i++ {
		id := dataProvider.Contains(longitudes[i], latitudes[i])
		if id == 0 {
			falseResultCnt++
		} else {
			trueResultCnt++
		}
	}
	b.Logf("%v - %v\n", trueResultCnt, falseResultCnt)
}

func BenchmarkQueryCity(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	dataProvider, err := NewGeoDataProvider("", []string{
		"./data/level5/e16360941e35b68a71987d8123182594.geojson.gz",
	})
	assert.Nil(b, err)
	longitudes := randFloats(73.33, 135.05, b.N)
	latitudes := randFloats(3.51, 53.33, b.N)
	b.ResetTimer()

	trueResultCnt, falseResultCnt := 0, 0
	for i := 0; i < b.N; i++ {
		id := dataProvider.Contains(longitudes[i], latitudes[i])
		if id == 0 {
			falseResultCnt++
		} else {
			trueResultCnt++
		}
	}
	b.Logf("%v - %v\n", trueResultCnt, falseResultCnt)
}
