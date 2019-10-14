package mom

import (
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

const (
	_testMongodbCollectionApps         = "test_apps"
	_testMongodbBaseCollectionMappings = "test_mom"
)

func _initMongodbApps() IDaoApp {
	mc := createMongoConnect()
	_, err := mc.CreateCollection(_testMongodbCollectionApps)
	if err != nil {
		panic(err)
	}
	_, err = mc.GetCollection(_testMongodbCollectionApps).DeleteMany(nil, bson.M{})
	if err != nil {
		panic(err)
	}
	return NewMongodbDaoApp(mc, _testMongodbCollectionApps)
}

func TestMongodbDaoApp_Create(t *testing.T) {
	name := "TestMongodbDaoApp_Create"
	dao := _initMongodbApps()
	ok, err := dao.Create(_testApp)
	if !ok || err != nil {
		t.Fatalf("%s failed - error creating app [%s]: %#v %e", name, _testAppId, ok, err)
	}
}

func TestMongodbDaoApp_CreateGet(t *testing.T) {
	name := "TestMongodbDaoApp_CreateGet"
	dao := _initMongodbApps()

	ok, err := dao.Create(_testApp)
	if !ok || err != nil {
		t.Fatalf("%s failed - error creating app [%s]: %#v %e", name, _testAppId, ok, err)
	}

	app, err := dao.Get(_testAppId)
	if err != nil {
		t.Fatalf("%s failed - error getting app [%s]: %e", name, _testAppId, err)
	}
	if app == nil {
		t.Fatalf("%s failed - app [%s] not found", name, _testAppId)
	}
}

func TestMongodbDaoApp_CreateGetDeleteGet(t *testing.T) {
	name := "TestMongodbDaoApp_CreateGetDeleteGet"
	dao := _initMongodbApps()

	ok, err := dao.Create(_testApp)
	if !ok || err != nil {
		t.Fatalf("%s failed - error creating app [%s]: %#v %e", name, _testAppId, ok, err)
	}

	app, err := dao.Get(_testAppId)
	if err != nil {
		t.Fatalf("%s failed - error getting app [%s]: %e", name, _testAppId, err)
	}
	if app == nil {
		t.Fatalf("%s failed - app [%s] not found", name, _testAppId)
	}

	ok, err = dao.Delete(_testApp)
	if !ok || err != nil {
		t.Fatalf("%s failed - error deleting app [%s]: %#v %e", name, _testAppId, ok, err)
	}
	app, err = dao.Get(_testAppId)
	if err != nil {
		t.Fatalf("%s failed - error getting app [%s]: %e", name, _testAppId, err)
	}
	if app != nil {
		t.Fatalf("%s failed - app [%s] should have been deleted", name, _testAppId)
	}
}

func TestMongodbDaoApp_CreateUpdateGet(t *testing.T) {
	name := "TestMongodbDaoApp_CreateUpdateGet"
	dao := _initMongodbApps()

	ok, err := dao.Create(_testApp)
	if !ok || err != nil {
		t.Fatalf("%s failed - error creating app [%s]: %#v %e", name, _testAppId, ok, err)
	}

	app := _testApp.Clone()
	app.Secret = _testAppSecret + "-cloned"
	app.Config["desc"] = _testAppDesc + "-cloned"
	ok, err = dao.Update(app)
	if !ok || err != nil {
		t.Fatalf("%s failed - error updating app [%s]: %#v %e", name, _testAppId, ok, err)
	}

	app, err = dao.Get(_testAppId)
	if err != nil {
		t.Fatalf("%s failed - error getting app [%s]: %e", name, _testAppId, err)
	}
	if app == nil {
		t.Fatalf("%s failed - app [%s] not found", name, _testAppId)
	}

	if app.Secret != _testAppSecret+"-cloned" {
		t.Fatalf("%s failed - expect app.Secret to be [%s] but received [%s]", name, _testAppSecret+"-cloned", app.Secret)
	}
	if app.Config["desc"] != _testAppDesc+"-cloned" {
		t.Fatalf("%s failed - expect app.Config['desc'] to be [%s] but received [%s]", name, _testAppDesc+"-cloned", app.Secret)
	}
}

/*----------------------------------------------------------------------*/

func _initMongodbMappings() IDaoMoMapping {
	mc := createMongoConnect()
	_, err := mc.GetCollection(_testMongodbCollectionApps).DeleteMany(nil, bson.M{})
	if err != nil {
		panic(err)
	}
	return NewMongodbDaoMoMapping(mc, _testMongodbBaseCollectionMappings)
}

func TestMongodbDaoMoMapping_InitStorage(t *testing.T) {
	name := "TestMongodbDaoMoMapping_InitStorage"
	dao := _initMongodbMappings()
	err := dao.InitStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
}

func TestMongodbDaoMoMapping_DestroyStorage(t *testing.T) {
	name := "TestMongodbDaoMoMapping_DestroyStorage"
	dao := _initMongodbMappings()
	err := dao.DestroyStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
}

func TestMongodbDaoMoMapping_Map(t *testing.T) {
	name := "TestMongodbDaoMoMapping_Map"
	dao := _initMongodbMappings()
	err := dao.DestroyStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	err = dao.InitStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	ns := "email"
	object := "btnguyen2k(at)1.email"
	target := "thanhnb"
	bo, err := dao.Map(_testAppId, ns, object, target)
	if bo == nil || err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	if bo.Namespace != normalizeNamespace(ns) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeNamespace(ns), bo.Namespace)
	}
	if bo.AppId != _testAppId {
		t.Fatalf("%s failed - expect %#v but received %#v", name, _testAppId, bo.AppId)
	}
	if bo.From != normalizeMappingObject(ns, object) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingObject(ns, object), bo.From)
	}
	if bo.To != normalizeMappingTarget(target) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingTarget(target), bo.To)
	}
}

func TestMongodbDaoMoMapping_MapFindTarget(t *testing.T) {
	name := "TestMongodbDaoMoMapping_MapFindTarget"
	dao := _initMongodbMappings()
	err := dao.DestroyStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	err = dao.InitStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	ns := "email"
	object := "btnguyen2k(at)1.email"
	target := "thanhnb"
	_, err = dao.Map(_testAppId, ns, object, target)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	bo, err := dao.FindTargetForObject(_testAppId, ns, object)
	if bo == nil || err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	if bo.Namespace != normalizeNamespace(ns) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeNamespace(ns), bo.Namespace)
	}
	if bo.AppId != _testAppId {
		t.Fatalf("%s failed - expect %#v but received %#v", name, _testAppId, bo.AppId)
	}
	if bo.From != normalizeMappingObject(ns, object) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingObject(ns, object), bo.From)
	}
	if bo.To != normalizeMappingTarget(target) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingTarget(target), bo.To)
	}
}

func TestMongodbDaoMoMapping_MapFindObjects(t *testing.T) {
	name := "TestMongodbDaoMoMapping_MapFindObjects"
	dao := _initMongodbMappings()
	err := dao.DestroyStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	err = dao.InitStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	ns := "email"
	object1 := "btnguyen2k(at)1.email"
	object2 := "thanhnb(at)2.email"
	target := "thanhnb"
	_, err = dao.Map(_testAppId, ns, object1, target)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	_, err = dao.Map(_testAppId, ns, object2, target)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}

	boList, err := dao.FindObjectsToTarget(_testAppId, ns, target)
	if err != nil || boList == nil || len(boList) != 2 {
		t.Fatalf("%s failed: %e", name, err)
	}
	for _, bo := range boList {
		if bo.Namespace != normalizeNamespace(ns) {
			t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeNamespace(ns), bo.Namespace)
		}
		if bo.AppId != _testAppId {
			t.Fatalf("%s failed - expect %#v but received %#v", name, _testAppId, bo.AppId)
		}
		if bo.From != normalizeMappingObject(ns, object1) && bo.From != normalizeMappingObject(ns, object2) {
			t.Fatalf("%s failed - expect %#v or %#v but received %#v", name, normalizeMappingObject(ns, object1), normalizeMappingObject(ns, object2), bo.From)
		}
		if bo.To != normalizeMappingTarget(target) {
			t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingTarget(target), bo.To)
		}
	}
}

func TestMongodbDaoMoMapping_MapUnmap(t *testing.T) {
	name := "TestMongodbDaoMoMapping_MapUnmap"
	dao := _initMongodbMappings()
	err := dao.DestroyStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	err = dao.InitStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	ns := "email"
	object1 := "btnguyen2k(at)1.email"
	object2 := "thanhnb(at)2.email"
	target := "thanhnb"
	_, err = dao.Map(_testAppId, ns, object1, target)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	_, err = dao.Map(_testAppId, ns, object2, target)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	dao.Unmap(_testAppId, ns, object1, target)
	boList, err := dao.FindObjectsToTarget(_testAppId, ns, target)
	if err != nil || boList == nil || len(boList) != 1 {
		t.Fatalf("%s failed: %e", name, err)
	}
	for _, bo := range boList {
		if bo.Namespace != normalizeNamespace(ns) {
			t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeNamespace(ns), bo.Namespace)
		}
		if bo.AppId != _testAppId {
			t.Fatalf("%s failed - expect %#v but received %#v", name, _testAppId, bo.AppId)
		}
		if bo.From != normalizeMappingObject(ns, object2) {
			t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingObject(ns, object2), bo.From)
		}
		if bo.To != normalizeMappingTarget(target) {
			t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingTarget(target), bo.To)
		}
	}
}

func TestMongodbDaoMoMapping_Allocate(t *testing.T) {
	name := "TestMongodbDaoMoMapping_Allocate"
	dao := _initMongodbMappings()
	err := dao.DestroyStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	err = dao.InitStorage(_testAppId)
	if err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	ns1 := "email"
	ns2 := "phone"
	object1 := "thanhnb(at)2.email"
	object2 := "09876544321"
	target := "thanhnb"
	finalTarget, err := dao.Allocate(_testAppId, map[string]string{ns1: object1, ns2: object2}, target)
	if err != nil || finalTarget != target {
		t.Fatalf("%s failed: %e", name, err)
	}

	bo1, err := dao.FindTargetForObject(_testAppId, ns1, object1)
	if bo1 == nil || err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	if bo1.Namespace != normalizeNamespace(ns1) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeNamespace(ns1), bo1.Namespace)
	}
	if bo1.AppId != _testAppId {
		t.Fatalf("%s failed - expect %#v but received %#v", name, _testAppId, bo1.AppId)
	}
	if bo1.From != normalizeMappingObject(ns1, object1) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingObject(ns1, object1), bo1.From)
	}
	if bo1.To != normalizeMappingTarget(target) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingTarget(target), bo1.To)
	}

	bo2, err := dao.FindTargetForObject(_testAppId, ns2, object2)
	if bo2 == nil || err != nil {
		t.Fatalf("%s failed: %e", name, err)
	}
	if bo2.Namespace != normalizeNamespace(ns2) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeNamespace(ns2), bo2.Namespace)
	}
	if bo2.AppId != _testAppId {
		t.Fatalf("%s failed - expect %#v but received %#v", name, _testAppId, bo2.AppId)
	}
	if bo2.From != normalizeMappingObject(ns2, object2) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingObject(ns2, object2), bo2.From)
	}
	if bo2.To != normalizeMappingTarget(target) {
		t.Fatalf("%s failed - expect %#v but received %#v", name, normalizeMappingTarget(target), bo2.To)
	}
}
