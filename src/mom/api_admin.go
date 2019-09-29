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
apiListApps handles API call "listApps".

Input parameters: none.

Output:

	- itineris.StatusErrorServer: error on server during API call.
	- itineris.StatusOk: successful, available apps are returned in `data` field as an array.

Authorization: only "system" app can call this API.
*/
func apiListApps(_ *itineris.ApiContext, auth *itineris.ApiAuth, _ *itineris.ApiParams) *itineris.ApiResult {
	if auth.GetAppId() != appSystem {
		return itineris.ResultNoPermission
	}

	apps, err := daoApp.GetAll()
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	return itineris.NewApiResult(itineris.StatusOk).SetData(apps)
}

/*
apiCreateApp handles API call "createApp".

Input parameters:

	- id: (optional, string) app's unique id. If not provided, a unique id will be generated.
	- secret: (string) app's secret key, used for authentication.
	- other arbitrary fields/values.

Output:

	- itineris.StatusErrorClient: missing or invalid input parameters.
	- itineris.StatusErrorServer: error on server during API call.
	- itineris.StatusConflict: app with specified id already existed.
	- itineris.StatusOk: successful, app's id is returned in `data` field.

Authorization: only "system" app can call this API.
*/
func apiCreateApp(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	if auth.GetAppId() != appSystem {
		return itineris.ResultNoPermission
	}

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

	err = daoMappings.InitStorage(id)
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

	ok, err := daoApp.Create(app)
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if !ok {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(fmt.Sprintf("Cannot create app [%s].", id))
	}
	return itineris.NewApiResult(itineris.StatusOk).SetData(id)
}

/*
apiGetApp handles API call "getApp".

Input parameters: none.

Output:

	- itineris.StatusErrorServer: error on server during API call.
	- itineris.StatusNotFound: app does not exist.
	- itineris.StatusOk: successful, app's info is returned in `data` field as a map.

Authorization: only "system" and owner app can call this API.
*/
func apiGetApp(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	id := params.GetParamAsTypeUnsafe("id", reddo.TypeString)
	if id == nil {
		return itineris.ResultNotFound
	}
	if auth.GetAppId() != appSystem && auth.GetAppId() != id.(string) {
		return itineris.ResultNoPermission
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
apiUpdateApp handles API call "updateApp".

Input parameters:

	- id: (string) app's id.
	- secret: (optional, string) new app's secret key.
	- other arbitrary fields/values.

Output:

	- itineris.StatusErrorServer: error on server during API call.
	- itineris.StatusNotFound: app does not exist.
	- itineris.StatusOk: successful.

Authorization: only "system" and owner app can call this API.
*/
func apiUpdateApp(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	id := params.GetParamAsTypeUnsafe("id", reddo.TypeString)
	if id == nil {
		return itineris.ResultNotFound
	}
	if auth.GetAppId() != appSystem && auth.GetAppId() != id.(string) {
		return itineris.ResultNoPermission
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
apiDeleteApp handles API call "deleteApp".

Input parameters:

	- id: (string) app's id.

Output:

	- itineris.StatusErrorServer: error on server during API call.
	- itineris.StatusNotFound: app does not exist.
	- itineris.StatusOk: successful.

Authorization: only "system" app can call this API.
*/
func apiDeleteApp(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	if auth.GetAppId() != appSystem {
		return itineris.ResultNoPermission
	}

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
