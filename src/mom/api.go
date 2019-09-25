package mom

import (
	"fmt"
	"github.com/btnguyen2k/consu/reddo"
	"main/src/itineris"
)

/*
API handler "getMappingForObject"
*/
func apiGetMappingForObject(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	ns := params.GetParamAsTypeUnsafe("ns", reddo.TypeString)
	if ns == nil || ns.(string) == "" {
		return itineris.ResultNotFound
	}
	obj := params.GetParamAsTypeUnsafe("from", reddo.TypeString)
	if obj == nil || obj.(string) == "" {
		return itineris.ResultNotFound
	}
	appId := auth.GetAppId()
	obj = normalizeMappingObject(ns.(string), obj.(string))
	mapping, err := daoMappings.FindTargetForObject(appId, ns.(string), obj.(string))
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if mapping == nil {
		return itineris.ResultNotFound
	}
	return itineris.NewApiResult(itineris.StatusOk).SetData(mapping)
}

/*
API handler "mapObjectToTarget"
*/
func apiMapObjectToTarget(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	ns := params.GetParamAsTypeUnsafe("ns", reddo.TypeString)
	if ns == nil || ns.(string) == "" {
		return itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [ns].")
	}
	obj := params.GetParamAsTypeUnsafe("from", reddo.TypeString)
	if obj == nil || obj.(string) == "" {
		return itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [from].")
	}
	target := params.GetParamAsTypeUnsafe("to", reddo.TypeString)
	if target == nil || target.(string) == "" {
		return itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [to].")
	}

	appId := auth.GetAppId()
	ns = normalizeNamespace(ns.(string))
	obj = normalizeMappingObject(ns.(string), obj.(string))
	target = normalizeMappingTarget(target.(string))
	mapping, err := daoMappings.FindTargetForObject(appId, ns.(string), obj.(string))
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if mapping != nil {
		// obj has already mapped to a target
		if mapping.To != target.(string) {
			return itineris.NewApiResult(itineris.StatusConflict).
				SetMessage(fmt.Sprintf("[%s] has already mapped to another target in namespace [%s].", obj, ns))
		}
		return itineris.NewApiResult(itineris.StatusOk).SetData(mapping)
	}

	if !arbitraryTargetMode {
		reversedMappings, err := daoMappings.FindObjectsToTarget(appId, ns.(string), target.(string))
		if err != nil {
			return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
		}
		if reversedMappings == nil || len(reversedMappings) == 0 {
			return itineris.NewApiResult(itineris.StatusErrorClient).SetMessage(fmt.Sprintf("Target [%s] not found and arbitraryTargetMode is diabled.", target))
		}
	}

	mapping, err = daoMappings.Map(appId, ns.(string), obj.(string), target.(string))
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	return itineris.NewApiResult(itineris.StatusOk).SetData(mapping)
}
