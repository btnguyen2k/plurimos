package mom

import (
	"fmt"
	"github.com/btnguyen2k/consu/reddo"
	"main/src/itineris"
	"main/src/utils"
	"strings"
	"time"
)

/*
API handler "apiListApps"
*/
func apiListApps(_ *itineris.ApiContext, _ *itineris.ApiAuth, _ *itineris.ApiParams) *itineris.ApiResult {
	apps, err := daoApp.GetAll()
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	return itineris.NewApiResult(itineris.StatusOk).SetData(apps)
}

/*
API handler "apiCreateApp"
*/
func apiCreateApp(_ *itineris.ApiContext, _ *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	_secret := params.GetParamAsTypeUnsafe("secret", reddo.TypeString)
	if _secret == nil || strings.TrimSpace(_secret.(string)) == "" {
		return itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Parameter [secret] must not be empty.")
	}
	_id := params.GetParamAsTypeUnsafe("id", reddo.TypeString)
	if _id == nil || strings.TrimSpace(_id.(string)) == "" {
		_id = utils.UniqueIdSmall()
	}
	secret := strings.TrimSpace(_secret.(string))
	id := strings.ToLower(strings.TrimSpace(_id.(string)))

	app, err := daoApp.Get(id)
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if app != nil {
		return itineris.NewApiResult(itineris.StatusConflict).SetMessage(fmt.Sprintf("App [%s] already existed.", id))
	}

	appData := params.GetAllParams()
	delete(appData, "secret")
	delete(appData, "id")
	app = &BoApp{
		Id:     id,
		Secret: utils.Sha1SumStr(id + "." + secret),
		Time:   time.Now(),
		Config: appData,
	}
	err = daoMappings.InitStorage(id)
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	ok, err := daoApp.Create(app)
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if !ok {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(fmt.Sprintf("Cannot create app [%s].", id))
	}
	return itineris.ResultOk
}

/*
API handler "apiGetApp"
*/
func apiGetApp(_ *itineris.ApiContext, _ *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	id := params.GetParamAsTypeUnsafe("id", reddo.TypeString)
	if id == nil {
		return itineris.ResultNotFound
	}
	app, err := daoApp.Get(id.(string))
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if app == nil {
		return itineris.ResultNotFound
	}
	return itineris.NewApiResult(itineris.StatusOk).SetData(app)
}

/*
API handler "apiUpdateApp"
*/
func apiUpdateApp(_ *itineris.ApiContext, _ *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	id := params.GetParamAsTypeUnsafe("id", reddo.TypeString)
	if id == nil {
		return itineris.ResultNotFound
	}
	app, err := daoApp.Get(id.(string))
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if app == nil {
		return itineris.ResultNotFound
	}
	secret := params.GetParamAsTypeUnsafe("secret", reddo.TypeString)

	appData := params.GetAllParams()
	delete(appData, "secret")
	delete(appData, "id")
	if secret != nil && strings.TrimSpace(secret.(string)) != "" {
		app.Secret = utils.Sha1SumStr(app.Id + "." + strings.TrimSpace(secret.(string)))
	}
	app.Time = time.Now()
	app.Config = appData
	ok, err := daoApp.Update(app)
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if !ok {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(fmt.Sprintf("Cannot update app [%s].", id))
	}
	return itineris.ResultOk
}

/*
API handler "apiDeleteApp"
*/
func apiDeleteApp(_ *itineris.ApiContext, _ *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	id := params.GetParamAsTypeUnsafe("id", reddo.TypeString)
	if id == nil {
		return itineris.ResultNotFound
	}
	app, err := daoApp.Get(id.(string))
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if app == nil {
		return itineris.ResultNotFound
	}
	ok, err := daoApp.Delete(app)
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if !ok {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(fmt.Sprintf("Cannot delete app [%s].", id))
	}
	err = daoMappings.DestroyStorage(app.Id)
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	return itineris.ResultOk
}
