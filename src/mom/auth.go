package mom

import (
	"log"
	"main/src/itineris"
	"main/src/utils"
	"strings"
)

func NewMomApiAuthenticator() itineris.IApiAuthenticator {
	return &ApiAuthenticator{}
}

type ApiAuthenticator struct {
}

func (a *ApiAuthenticator) Authenticate(ctx *itineris.ApiContext, auth *itineris.ApiAuth) bool {
	app, err := daoApp.Get(auth.GetAppId())
	if err != nil {
		log.Printf("Error while loading app [%s]: %e", auth.GetAppId(), err)
		return false
	}
	if app == nil || !strings.EqualFold(app.Secret, utils.Sha1SumStr(app.Id+"."+auth.GetAccessToken())) {
		return false
	}
	return true
}
