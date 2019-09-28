package mom

import (
	"fmt"
	"github.com/btnguyen2k/consu/reddo"
	"main/src/itineris"
	"regexp"
	"strings"
)

func parseParam(params *itineris.ApiParams, name string, defaultResult *itineris.ApiResult) (string, *itineris.ApiResult) {
	value := params.GetParamAsTypeUnsafe(name, reddo.TypeString)
	if value == nil || strings.TrimSpace(value.(string)) == "" {
		return "", defaultResult
	}
	return strings.TrimSpace(value.(string)), nil
}

/*
API handler "getMappingForObject"
*/
func apiGetMappingForObject(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	var ns, obj string
	var result *itineris.ApiResult
	if ns, result = parseParam(params, "ns", itineris.ResultNotFound); result != nil {
		return result
	}
	if obj, result = parseParam(params, "from", itineris.ResultNotFound); result != nil {
		return result
	}

	appId := auth.GetAppId()
	obj = normalizeMappingObject(ns, obj)
	mapping, err := daoMappings.FindTargetForObject(appId, ns, obj)
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
	var ns, obj, target string
	var result *itineris.ApiResult
	if ns, result = parseParam(params, "ns", itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [ns].")); result != nil {
		return result
	}
	if obj, result = parseParam(params, "from", itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [from].")); result != nil {
		return result
	}
	if target, result = parseParam(params, "to", itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [to].")); result != nil {
		return result
	}

	appId := auth.GetAppId()
	ns = normalizeNamespace(ns)
	obj = normalizeMappingObject(ns, obj)
	target = normalizeMappingTarget(target)
	mapping, err := daoMappings.FindTargetForObject(appId, ns, obj)
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	if mapping != nil {
		// obj has already mapped to a target
		if mapping.To != target {
			return itineris.NewApiResult(itineris.StatusConflict).
				SetMessage(fmt.Sprintf("[%s] has already mapped to another target in namespace [%s].", obj, ns))
		}
		return itineris.NewApiResult(itineris.StatusOk).SetData(mapping)
	}

	if !arbitraryTargetMode {
		reversedMappings, err := daoMappings.FindObjectsToTarget(appId, ns, target)
		if err != nil {
			return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
		}
		if reversedMappings == nil || len(reversedMappings) == 0 {
			return itineris.NewApiResult(itineris.StatusErrorClient).SetMessage(fmt.Sprintf("Target [%s] not found and arbitraryTargetMode is diabled.", target))
		}
	}

	mapping, err = daoMappings.Map(appId, ns, obj, target)
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	return itineris.NewApiResult(itineris.StatusOk).SetData(mapping)
}

/*
API handler "unmapObjectToTarget"
*/
func apiUnmapObjectToTarget(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	var ns, obj, target string
	var result *itineris.ApiResult
	if ns, result = parseParam(params, "ns", itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [ns].")); result != nil {
		return result
	}
	if obj, result = parseParam(params, "from", itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [from].")); result != nil {
		return result
	}
	if target, result = parseParam(params, "to", itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [to].")); result != nil {
		return result
	}

	appId := auth.GetAppId()
	ns = normalizeNamespace(ns)
	obj = normalizeMappingObject(ns, obj)
	target = normalizeMappingTarget(target)
	_, err := daoMappings.Unmap(appId, ns, obj, target)
	if err != nil {
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	return itineris.ResultOk
}

/*
API handler "getReverseMappinngsForTarget"
*/
func apiGetReverseMappinngsForTarget(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	var nsList, target string
	var result *itineris.ApiResult
	if nsList, result = parseParam(params, "ns", itineris.ResultNotFound); result != nil {
		return result
	}
	if target, result = parseParam(params, "to", itineris.ResultNotFound); result != nil {
		return result
	}
	resultData := map[string][]*BoMapping{}
	appId := auth.GetAppId()
	target = normalizeMappingTarget(target)
	namespaces := regexp.MustCompile(`[,;\s]+`).Split(nsList, -1)
	for _, ns := range namespaces {
		ns = normalizeNamespace(ns)
		mappings, err := daoMappings.FindObjectsToTarget(appId, ns, target)
		if err != nil {
			return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
		}
		resultData[ns] = mappings
	}
	return itineris.ResultOk.SetData(resultData)
}
