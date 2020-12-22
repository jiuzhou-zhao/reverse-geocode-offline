## reverse geo code 从经纬度查询省份，城市等

1. 数据来源 [OSMBoundariesMap](https://wambachers-osm.website/boundaries)
2. cmd/reverse-geocode-offline 下为一个提供经纬度到城市或省转换的`http`服务器代码
3. 使用示例比较简单，就不写了，可以从`benchmark`上测试上来看
    ```go
    func BenchmarkQuery(b *testing.B) {
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
    ```
> 文章参考 [Doc](https://patdz.github.io/2020/12/22/di-tu-jing-wei-du-suo-zai-sheng-fen-cheng-shi/)