package mom

import (
	"fmt"
	"github.com/btnguyen2k/consu/reddo"
	"main/src/itineris"
	"main/src/utils"
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
apiGetMappingForObject handles API "getMappingForObject".

Input parameters:

	- ns: (string) namespace
	- from: (string) object

Output:

	- itineris.StatusErrorServer: error on server during API call.
	- itineris.StatusNotFound: object is not mapping to any target in the namespace.
	- itineris.StatusOk: successful, mapping data is returned in `data` field as a map.
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
apiMapObjectToTarget handles API "mapObjectToTarget".

Input parameters:

	- ns: (string) namespace
	- from: (string) object
	- to: (string)target

Output:

	- itineris.StatusErrorClient: missing or invalid input parameters.
	- itineris.StatusErrorServer: error on server during API call.
	- itineris.StatusConflict: object has already mapped to another target in the namespace.
	- itineris.StatusOk: successful, mapping data is returned in `data` field as a map.

ArbitraryTargetMode:

	- false: target must exist in the namespace or API will fail with status itineris.StatusErrorClient.
	- true: server will not check for target's existence.
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
apiUnmapObjectToTarget handles API "unmapObjectToTarget".

Input parameters:

	- ns: (string) namespace
	- from: (string) object
	- to: (string)target

Output:

	- itineris.StatusErrorClient: missing or invalid input parameters.
	- itineris.StatusErrorServer: error on server during API call.
	- itineris.StatusOk: successful.
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
apiGetReverseMappinngsForTarget handles API "getReverseMappinngsForTarget".

Input parameters:

	- ns: (string) list of namespace, separated by comma (,) or semi-colon (;)
	- to: (string)target

Output:

	- itineris.StatusErrorClient: missing or invalid input parameters.
	- itineris.StatusErrorServer: error on server during API call.
	- itineris.StatusOk: successful, reversed mappings are returned in `data` field as a map {namespace: [array of mappings found in the namespace]}
*/
func apiGetReverseMappinngsForTarget(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	var nsList, target string
	var result *itineris.ApiResult
	if nsList, result = parseParam(params, "ns", itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [ns].")); result != nil {
		return result
	}
	if target, result = parseParam(params, "to", itineris.NewApiResult(itineris.StatusErrorClient).SetMessage("Required parameter [to].")); result != nil {
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
	return itineris.NewApiResult(itineris.StatusOk).SetData(resultData)
}

/*
apiAllocateTargetAndMap handles API "allocateTargetAndMap"

Input parameters:

	- a map of {namespace: object}

Output:

	- itineris.StatusErrorClient: missing or invalid input parameters.
	- itineris.StatusErrorServer: error on server during API call.
	- itineris.StatusOk: successful, reversed mappings are returned in `data` field as a map {}
*/
func apiAllocateTargetAndMap(_ *itineris.ApiContext, auth *itineris.ApiAuth, params *itineris.ApiParams) *itineris.ApiResult {
	appId := auth.GetAppId()
	mapNsObj := make(map[string]string)
	for k, v := range params.GetAllParams() {
		ns := normalizeNamespace(k)
		obj, _ := reddo.ToString(v)
		mapNsObj[ns] = normalizeMappingObject(ns, obj)
	}
	target := normalizeMappingTarget(utils.UniqueIdSmall())
	target, err := daoMappings.Allocate(appId, mapNsObj, target)
	if err != nil {
		if target != "" {
			return itineris.NewApiResult(itineris.StatusConflict).SetMessage(err.Error())
		}
		return itineris.NewApiResult(itineris.StatusErrorServer).SetMessage(err.Error())
	}
	return itineris.NewApiResult(itineris.StatusOk).SetData(target)
}
