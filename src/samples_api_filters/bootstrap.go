package samples_api_filters

import (
	"main/src/goems"
	"main/src/itineris"
	"os"
	"strconv"
)

type MyBootstrapper struct {
	name string
}

var Bootstrapper = &MyBootstrapper{name: "samples_api_filters"}

func (b *MyBootstrapper) Bootstrap() error {
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
	apiFilter = itineris.NewAuthenticationFilter(goems.ApiRouter, apiFilter, NewDummyApiAuthenticator())
	apiFilter = itineris.NewLoggingFilter(goems.ApiRouter, apiFilter, itineris.NewWriterRequestLogger(os.Stdout, appName, appVersion))

	goems.ApiRouter.SetApiFilter(apiFilter)
	return nil
}

/*----------------------------------------------------------------------*/

func NewDummyApiAuthenticator() *DummyApiAuthenticator {
	return &DummyApiAuthenticator{}
}

/*
DummyApiAuthenticator is a dummy "IApiAuthenticator" which checks:

	- AppId must be "dummy"
	- AccessToken must be a positive number divisible to 5
*/
type DummyApiAuthenticator struct {
}

/*
Authenticate implements IApiAuthenticator.Authenticate.
*/
func (a *DummyApiAuthenticator) Authenticate(_ *itineris.ApiContext, auth *itineris.ApiAuth) bool {
	if "dummy" != auth.GetAppId() {
		return false
	}
	v, e := strconv.Atoi(auth.GetAccessToken())
	if e != nil || v <= 0 || v%5 != 0 {
		return false
	}
	return true
}
