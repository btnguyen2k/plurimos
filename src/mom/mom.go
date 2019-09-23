/*
Many-to-one mapping service.

@author Thanh Nguyen <btnguyen2k@gmail.com>
@since 0.1.0
*/
package mom

import (
	"github.com/btnguyen2k/prom"
	"main/src/goems"
	"main/src/itineris"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// Version of mom
	Version = "0.1.0"
)

type MyBootstrapper struct {
	name string
}

var (
	Bootstrapper = &MyBootstrapper{name: "mom"}
	mongoConnect *prom.MongoConnect
	daoMappings  IDaoMoMapping
	daoApp       IDaoApp
	startupTime  = time.Now()
)

/*
Bootstrap implements goems.IBootstrapper.Bootstrap

Bootstrapper usually does:
- register api-handlers with the global ApiRouter
- other initializing work (e.g. creating DAO, initializing database, etc)
*/
func (b *MyBootstrapper) Bootstrap() error {
	initFilters()
	initDaos()
	initApiHandlers(goems.ApiRouter)
	return nil
}

func initFilters() {
	var apiFilter itineris.IApiFilter = nil
	appName := goems.AppConfig.GetString("app.name")
	appVersion := goems.AppConfig.GetString("app.version")

	// filters are LIFO:
	// - request goes through the last filter to the first one
	// - response goes through the first filter to the last one
	// suggested order of filters:
	// - Request logger should be the last one to capture full request/response
	apiFilter = itineris.NewAddPerfInfoFilter(goems.ApiRouter, apiFilter)
	apiFilter = itineris.NewLoggingFilter(goems.ApiRouter, apiFilter, itineris.NewWriterPerfLogger(os.Stderr, appName, appVersion))
	// apiFilter = itineris.NewAuthenticationFilter(goems.ApiRouter, apiFilter, NewDummyApiAuthenticator())
	apiFilter = itineris.NewLoggingFilter(goems.ApiRouter, apiFilter, itineris.NewWriterRequestLogger(os.Stdout, appName, appVersion))
}

func initDaos() {
	dbtype := goems.AppConfig.GetString("mom.db_type")
	if strings.EqualFold("mongo", dbtype) || strings.EqualFold("mongodb", dbtype) {
		// dbtype=MongoDB
		mongoConnect = createMongoConnect()
		err := initData(dbtype)
		if err != nil {
			panic(err)
		}

		daoApp = NewMongodbDaoApp(mongoConnect, collectionApps)
		daoMappings = NewMongodbDaoMoMapping(mongoConnect, baseCollectionMom)

		return
	}
	panic("Unknown database type: [" + dbtype + "].")
}

/*
Setup API handlers: application register its api-handlers by calling router.SetHandler(apiName, apiHandlerFunc)

    - api-handler function must has the following signature: func (itineris.ApiContext, itineris.ApiAuth, itineris.ApiParams) *itineris.ApiResult
*/
func initApiHandlers(router *itineris.ApiRouter) {
	router.SetHandler("info", apiInfo)

	router.SetHandler("listApps", apiListApps)
	router.SetHandler("createApp", apiCreateApp)
	router.SetHandler("getApp", apiGetApp)
	router.SetHandler("updateApp", apiUpdateApp)
	router.SetHandler("deleteApp", apiDeleteApp)
}

/*
API handler "info"
*/
func apiInfo(_ *itineris.ApiContext, _ *itineris.ApiAuth, _ *itineris.ApiParams) *itineris.ApiResult {
	appInfo := map[string]interface{}{
		"name":        goems.AppConfig.GetString("app.name"),
		"shortname":   goems.AppConfig.GetString("app.shortname"),
		"version":     goems.AppConfig.GetString("app.version"),
		"description": goems.AppConfig.GetString("app.desc"),
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	result := map[string]interface{}{
		"app": appInfo,
		"memory": map[string]interface{}{
			"alloc":     m.Alloc,
			"alloc_str": strconv.FormatFloat(float64(m.Alloc)/1024.0/1024.0, 'f', 1, 64) + " MiB",
			"sys":       m.Sys,
			"sys_str":   strconv.FormatFloat(float64(m.Sys)/1024.0/1024.0, 'f', 1, 64) + " MiB",
			"gc":        m.NumGC,
		},
		"uptime": time.Since(startupTime).String(),
	}
	return itineris.NewApiResult(itineris.StatusOk).SetData(result)
}
