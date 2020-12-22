package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jiuzhou-zhao/reverse-geocode-offline"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Web      string
	GeoJsons map[string][]string
}

type rgcRequest struct {
	Key       string  `form:"key" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	Latitude  float64 `form:"latitude" binding:"required"`
}

type rgcResponse struct {
	Code    int
	Message string `json:"message,omitempty"`
	Hit     bool
	Result  string
}

func main() {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))
	v.SetDefault("chart::values", map[string]interface{}{
		"ingress": map[string]interface{}{
			"annotations": map[string]interface{}{
				"traefik.frontend.rule.type":                 "PathPrefix",
				"traefik.ingress.kubernetes.io/ssl-redirect": "true",
			},
		},
	})
	v.SetConfigName("config")
	v.AddConfigPath(".")

	err := v.ReadInConfig()
	if err != nil {
		logrus.Fatalf("error reading config: %s", err)
	}

	var cfg Config
	err = v.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("error unmarshal config: %s", err)
	}

	r := gin.New()

	rootProvider, _ := reverse_geocode_offline.NewGeoDataProvider("", nil)
	for key, geoJsonFiles := range cfg.GeoJsons {
		provider, err := reverse_geocode_offline.NewGeoDataProvider(key, geoJsonFiles)
		if err != nil {
			logrus.Fatalf("invalid json files: %v-%v", key, geoJsonFiles)
		}
		rootProvider.InitAppendGeoDataProvider(key, provider)
	}

	//
	r.Use(gin.Recovery())

	// routers
	r.GET("/r_geo_code", func(c *gin.Context) {
		var req rgcRequest
		var resp rgcResponse
		err := c.ShouldBindWith(&req, binding.Query)
		if err != nil {
			resp.Code = -1
			resp.Message = err.Error()
			c.JSON(http.StatusOK, &resp)
			return
		}
		provider := rootProvider.GetGeoDataProvider(req.Key)
		if provider == nil {
			resp.Code = -2
			resp.Message = fmt.Sprintf("unknown key %v", req.Key)
			c.JSON(http.StatusOK, &resp)
			return
		}
		id := provider.Contains(req.Longitude, req.Latitude)
		if id == 0 {
			c.JSON(http.StatusOK, &resp)
			return
		}
		resp.Hit = true
		di := rootProvider.GetDataInfos(id)
		if di != nil {
			resp.Result += di.LocalName
			for _, pid := range di.ParentIDs {
				di = rootProvider.GetDataInfos(pid)
				if di != nil {
					resp.Result = di.LocalName + " " + resp.Result
				}
			}
		}
		c.JSON(http.StatusOK, &resp)
	})

	srv := &http.Server{
		Addr:    cfg.Web,
		Handler: r,
	}

	logrus.Infof("listen %s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logrus.Panicf("listen: %s", err)
	}
}
