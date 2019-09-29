/*
Many-to-one mapping service.

@author Thanh Nguyen <btnguyen2k@gmail.com>
@since 0.1.0
*/
package mom

import (
	"github.com/btnguyen2k/prom"
	"log"
	"main/src/goems"
	"main/src/itineris"
	"main/src/utils"
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

	mongoConnect        *prom.MongoConnect
	daoMappings         IDaoMoMapping
	daoApp              IDaoApp
	startupTime         = time.Now()
	arbitraryTargetMode bool
)

/*
Bootstrap implements goems.IBootstrapper.Bootstrap

Bootstrapper usually does:
- register api-handlers with the global ApiRouter
- other initializing work (e.g. creating DAO, initializing database, etc)
*/
func (b *MyBootstrapper) Bootstrap() error {
	arbitraryTargetMode = goems.AppConfig.GetBoolean("mom.arbitrary_target_mode", false)

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
	apiFilter = itineris.NewAuthenticationFilter(goems.ApiRouter, apiFilter, NewMomApiAuthenticator())
	apiFilter = itineris.NewLoggingFilter(goems.ApiRouter, apiFilter, itineris.NewWriterRequestLogger(os.Stdout, appName, appVersion))
	goems.ApiRouter.SetApiFilter(apiFilter)
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

		appSystem = goems.AppConfig.GetString("mom.system_app_name", "system")
		app, err := daoApp.Get(appSystem)
		if err != nil {
			panic(err)
		}
		if app == nil {
			log.Printf("Creating system app...")
			secret := goems.AppConfig.GetString("mom.system_app_secret", appSystem)
			app = &BoApp{
				Id:     appSystem,
				Secret: utils.Sha1SumStr(appSystem + "." + secret),
				Time:   time.Now(),
				Config: nil,
			}
		}
		if _, err := daoApp.Create(app); err != nil {
			panic(err)
		}
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

	router.SetHandler("mapObjectToTarget", apiMapObjectToTarget)
	router.SetHandler("getMappingForObject", apiGetMappingForObject)
	router.SetHandler("unmapObjectToTarget", apiUnmapObjectToTarget)
	router.SetHandler("getReverseMappinngsForTarget", apiGetReverseMappinngsForTarget)
	router.SetHandler("allocateTargetAndMap", apiAllocateTargetAndMap)
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
