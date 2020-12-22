package reverse_geocode_offline

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/pkg/errors"
)

type GeoDataInfo struct {
	ID        int64
	ParentIDs []int64
	Name      string
	LocalName string
	NameEn    string
}

type GeoDataProvider interface {
	Contains(longitude, latitude float64) int64
	GetDataInfos(id int64) *GeoDataInfo
	GetDataIDMap() map[int64]*geojson.Feature

	GetGeoDataProvider(key string) GeoDataProvider
	InitAppendGeoDataProvider(key string, provider GeoDataProvider)
}

type GeoDataProviderImpl struct {
	key               string
	featureCollection []*geojson.FeatureCollection
	featureIDMap      map[int64]*geojson.Feature
	subProviderMap    map[string]GeoDataProvider
}

func readGzFile(fileName string) (bs []byte, err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return
	}
	defer func() {
		_ = gzipReader.Close()
	}()
	return ioutil.ReadAll(gzipReader)
}

func NewGeoDataProvider(key string, geoJsonFiles []string) (GeoDataProvider, error) {
	providerImpl := &GeoDataProviderImpl{
		key:            key,
		subProviderMap: make(map[string]GeoDataProvider),
	}
	var featureCollections []*geojson.FeatureCollection
	featureIDMap := make(map[int64]*geojson.Feature)
	for _, f := range geoJsonFiles {
		bs, err := readGzFile(f)
		if err != nil {
			err = errors.Wrapf(err, "read file: %v", f)
			return nil, err
		}
		featureCollection, err := geojson.UnmarshalFeatureCollection(bs)
		if err != nil {
			err = errors.Wrapf(err, "parse file: %v", f)
			return nil, err
		}
		for _, feature := range featureCollection.Features {
			fid := providerImpl.getFeatureID(feature)
			if fid == 0 {
				err = fmt.Errorf("no feature id: %+v", feature)
				return nil, err
			}
			featureIDMap[fid] = feature
		}
		featureCollections = append(featureCollections, featureCollection)
	}
	providerImpl.featureCollection = featureCollections
	providerImpl.featureIDMap = featureIDMap

	return providerImpl, nil
}

func (provider *GeoDataProviderImpl) getFeatureID(feature *geojson.Feature) int64 {
	if idi, ok := feature.Properties["id"]; ok {
		if idf, ok := idi.(float64); ok {
			return int64(idf)
		}
	}
	return 0
}

func (provider *GeoDataProviderImpl) mapFeature(feature *geojson.Feature) *GeoDataInfo {
	info := &GeoDataInfo{
		ID: provider.getFeatureID(feature),
	}
	if v, ok := feature.Properties["parents"]; ok {
		if v != nil {
			vs := strings.Split(v.(string), ",")
			for _, v := range vs {
				id, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					continue
				}
				info.ParentIDs = append(info.ParentIDs, id)
			}
		}
	}
	if v, ok := feature.Properties["name"]; ok {
		info.Name = v.(string)
	}
	if v, ok := feature.Properties["local_name"]; ok {
		info.LocalName = v.(string)
	}
	if v, ok := feature.Properties["name_en"]; ok {
		info.NameEn = v.(string)
	}
	return info
}

func (provider *GeoDataProviderImpl) Contains(longitude, latitude float64) int64 {
	point := orb.Point{longitude, latitude}
	for _, featureCollection := range provider.featureCollection {
		for _, feature := range featureCollection.Features {
			// Try on a MultiPolygon to begin
			multiPoly, isMulti := feature.Geometry.(orb.MultiPolygon)
			if isMulti {
				if planar.MultiPolygonContains(multiPoly, point) {
					return provider.getFeatureID(feature)
				}
			} else {
				// Fallback to Polygon
				polygon, isPoly := feature.Geometry.(orb.Polygon)
				if isPoly {
					if planar.PolygonContains(polygon, point) {
						return provider.getFeatureID(feature)
					}
				}
			}
		}
	}
	return 0
}
func (provider *GeoDataProviderImpl) GetDataInfos(id int64) *GeoDataInfo {
	if feature, ok := provider.featureIDMap[id]; ok {
		return provider.mapFeature(feature)
	}
	return nil
}

func (provider *GeoDataProviderImpl) GetDataIDMap() map[int64]*geojson.Feature {
	return provider.featureIDMap
}

func (provider *GeoDataProviderImpl) GetGeoDataProvider(key string) GeoDataProvider {
	if prov, ok := provider.subProviderMap[key]; ok {
		return prov
	}
	return nil
}

func (provider *GeoDataProviderImpl) InitAppendGeoDataProvider(key string, prov GeoDataProvider) {
	provider.subProviderMap[key] = prov
	for key, feature := range prov.GetDataIDMap() {
		if _, ok := provider.featureIDMap[key]; ok {
			continue
		}
		provider.featureIDMap[provider.getFeatureID(feature)] = feature
	}
}
