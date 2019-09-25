package mom

import (
	"context"
	"github.com/btnguyen2k/consu/reddo"
	"github.com/btnguyen2k/godal"
	"github.com/btnguyen2k/godal/mongo"
	"github.com/btnguyen2k/prom"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"main/src/goems"
	"strings"
	"time"
)

/*
MOM's DAO implementation: MongoDB

@author Thanh Nguyen <btnguyen2k@gmail.com>
@since 0.1.0
*/

const (
	collectionApps        = "apps"
	collectionTemplateMom = "${collection}_${app}"
	baseCollectionMom     = "mom"
	_fieldId              = "_id"
)

// construct an 'prom.MongoConnect' instance
func createMongoConnect() *prom.MongoConnect {
	url := goems.AppConfig.GetString("mom.mongodb.url", "mongodb://mom:mom@localhost:27017/mom")
	db := goems.AppConfig.GetString("mom.mongodb.db", "mom")
	timeoutMs := goems.AppConfig.GetInt32("mom.mongodb.timeout", 10000)
	mongoConnect, err := prom.NewMongoConnect(url, db, int(timeoutMs))
	if mongoConnect == nil || err != nil {
		if err != nil {
			log.Println(err)
		}
		panic("error creating [prom.MongoConnect] instance")
	}
	return mongoConnect
}

func NewMongodbDaoMoMapping(mongoConnect *prom.MongoConnect, baseCollectionName string) IDaoMoMapping {
	dao := &MongodbDaoMoMapping{baseCollectionName: baseCollectionName, collectionInitCache: map[string]bool{}}
	dao.GenericDaoMongo = mongo.NewGenericDaoMongo(mongoConnect, godal.NewAbstractGenericDao(dao))
	dao.SetTransactionMode(true)
	return dao
}

type MongodbDaoMoMapping struct {
	*mongo.GenericDaoMongo
	baseCollectionName  string // name of collection store data
	collectionInitCache map[string]bool
}

func (dao *MongodbDaoMoMapping) calcCollectionName(appId string) string {
	collectionName := strings.ReplaceAll(collectionTemplateMom, "${collection}", dao.baseCollectionName)
	collectionName = strings.ReplaceAll(collectionName, "${app}", strings.ToLower(appId))
	return collectionName
}

/*
InitStorage implements IDaoMoMapping.IDaoMoMapping
*/
func (dao *MongodbDaoMoMapping) InitStorage(appId string) error {
	collectionName := dao.calcCollectionName(appId)
	exists := dao.collectionInitCache[collectionName]
	if exists {
		return nil
	}
	if exists, err := dao.GetMongoConnect().HasCollection(collectionName); exists || err != nil {
		if err != nil {
			return err
		}
		dao.collectionInitCache[collectionName] = true
		return nil
	}

	// create collection if not exists
	dbResult, err := dao.GetMongoConnect().CreateCollection(collectionName)
	if err != nil || dbResult.Err() != nil {
		if err != nil {
			log.Printf("Error while creating collection %s: %e", collectionName, err)
			return err
		} else {
			log.Printf("Error while creating collection %s: %e", collectionName, dbResult.Err())
			return dbResult.Err()
		}
	} else {
		log.Printf("Created collection %s", collectionName)
	}
	dao.collectionInitCache[collectionName] = true

	// create indexes
	dbResult, err = dao.GetMongoConnect().CreateIndexes(collectionName, []interface{}{
		map[string]interface{}{
			"key": map[string]interface{}{
				fieldMapNamespace: 1,
				fieldMapFrom:      1,
			},
			"name":   "uidx_from",
			"unique": true,
		},
		map[string]interface{}{
			"key": map[string]interface{}{
				fieldMapNamespace: 1,
				fieldMapTo:        1,
			},
			"name": "idx_to",
		},
	})
	if err != nil || dbResult.Err() != nil {
		if err != nil {
			log.Printf("Error while creating indexes on collection %s: %e", collectionName, err)
			return err
		} else {
			log.Printf("Error while creating indexes on collection %s: %e", collectionName, dbResult.Err())
			return dbResult.Err()
		}
	} else {
		log.Printf("Created indexes for collection %s", collectionName)
	}

	return nil
}

/*
DestroyStorage implements IDaoMoMapping.DestroyStorage
*/
func (dao *MongodbDaoMoMapping) DestroyStorage(appId string) error {
	collectionName := dao.calcCollectionName(appId)
	db := dao.GetMongoConnect().GetDatabase()
	ctx, _ := dao.GetMongoConnect().NewBackgroundContext()
	dbResult := db.RunCommand(ctx, bson.M{"drop": collectionName})
	if dbResult.Err() != nil {
		return dbResult.Err()
	}
	delete(dao.collectionInitCache, collectionName)
	return nil
}

// GdaoCreateFilter implements godal.IGenericDao.GdaoCreateFilter.
//
//  - DAO must implement GdaoCreateFilter!
func (dao *MongodbDaoMoMapping) GdaoCreateFilter(_ string, gbo godal.IGenericBo) interface{} {
	namespace := gbo.GboGetAttrUnsafe(fieldMapNamespace, reddo.TypeString).(string)
	from := gbo.GboGetAttrUnsafe(fieldMapFrom, reddo.TypeString).(string)
	return bson.M{fieldMapNamespace: namespace, fieldMapFrom: normalizeMappingObject(namespace, from)}
}

// toBo transforms godal.IGenericBo to BoApp
func (dao *MongodbDaoMoMapping) toBo(gbo godal.IGenericBo) *BoMapping {
	if gbo == nil {
		return nil
	}
	bo := BoMapping{}
	if err := gbo.GboTransferViaJson(&bo); err != nil {
		return nil
	}
	return &bo
}

// toGbo transforms godal.IGenericBo to BoApp
func (dao *MongodbDaoMoMapping) toGbo(bo *BoMapping) godal.IGenericBo {
	if bo == nil {
		return nil
	}
	gbo := godal.NewGenericBo()
	if err := gbo.GboImportViaJson(bo); err != nil {
		return nil
	}
	return gbo
}

func (dao *MongodbDaoMoMapping) doGetMapping(ctx context.Context, appId, namespace, from string) (*BoMapping, error) {
	collectionName := dao.calcCollectionName(appId)
	filter := bson.M{fieldMapNamespace: namespace, fieldMapFrom: normalizeMappingObject(namespace, from)}
	gbo, err := dao.GdaoFetchOne(collectionName, filter)
	return dao.toBo(gbo), err
}

/*
FindTargetForObject implements IDaoMoMapping.FindTargetForObject
*/
func (dao *MongodbDaoMoMapping) FindTargetForObject(appId, namespace, from string) (*BoMapping, error) {
	return dao.doGetMapping(nil, appId, namespace, from)
}

func (dao *MongodbDaoMoMapping) doGetReversedMappings(ctx context.Context, appId, namespace, to string) (map[string]*BoMapping, error) {
	collectionName := dao.calcCollectionName(appId)
	filter := bson.M{fieldMapNamespace: namespace, fieldMapTo: to}
	gboList, err := dao.GdaoFetchMany(collectionName, filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*BoMapping)
	if gboList != nil {
		for _, gbo := range gboList {
			bo := dao.toBo(gbo)
			if bo != nil {
				result[bo.From] = bo
			}
		}
	}
	return result, nil
}

/*
FindObjectsToTarget implements IDaoMoMapping.FindObjectsToTarget
*/
func (dao *MongodbDaoMoMapping) FindObjectsToTarget(appId, namespace, to string) (map[string]*BoMapping, error) {
	return dao.doGetReversedMappings(nil, appId, namespace, to)
}

func (dao *MongodbDaoMoMapping) doInsert(ctx context.Context, bo *BoMapping) (bool, error) {
	numRows, err := dao.GdaoCreate(dao.calcCollectionName(bo.AppId), dao.toGbo(bo))
	return numRows > 0, err
}

/*
Map implements IDaoMoMapping.Map
*/
func (dao *MongodbDaoMoMapping) Map(appId, namespace, object, target string) (*BoMapping, error) {
	bo := &BoMapping{
		Namespace: namespace,
		From:      object,
		To:        target,
		Time:      time.Now(),
		AppId:     appId,
	}
	_, err := dao.doInsert(nil, bo)
	if err != nil {
		return nil, err
	}
	return bo, nil
	// ctx, _ := dao.GetMongoConnect().NewBackgroundContext()
	// err := dao.GetMongoConnect().GetMongoClient().UseSession(ctx, func(sctx mongo2.SessionContext) error {
	//     err := sctx.StartTransaction(options.Transaction().SetReadConcern(readconcern.Snapshot()).SetWriteConcern(writeconcern.New(writeconcern.WMajority())))
	//     if err != nil {
	//         return err
	//     }
	//     curMap, err := dao.doGetMapping(sctx, appId, namespace, bo.From)
	//     if err != nil {
	//         return err
	//     }
	//     if curMap != nil && curMap.To != bo.To {
	//         return errors.Errorf("Object [%s] already mapped to another target.", object)
	//     } else if curMap == nil {
	//         _, err = dao.doInsert(sctx, bo)
	//         if err != nil {
	//             return err
	//         }
	//     }
	//     return sctx.CommitTransaction(sctx)
	// })
	// return bo, err
}

/*
MapIfTargetExists implements IDaoMoMapping.MapIfTargetExists
*/
func (dao *MongodbDaoMoMapping) MapIfTargetExists(appId, namespace, object, target string) (*BoMapping, error) {
	panic("implement me")
}

/*----------------------------------------------------------------------*/

func NewMongodbDaoApp(mc *prom.MongoConnect, collectionName string) IDaoApp {
	dao := &MongodbDaoApp{collectionName: collectionName}
	dao.GenericDaoMongo = mongo.NewGenericDaoMongo(mc, godal.NewAbstractGenericDao(dao))
	dao.SetTransactionMode(true)
	return dao
}

type MongodbDaoApp struct {
	*mongo.GenericDaoMongo
	collectionName string
}

// GdaoCreateFilter implements godal.IGenericDao.GdaoCreateFilter.
//
//  - DAO must implement GdaoCreateFilter!
func (dao *MongodbDaoApp) GdaoCreateFilter(_ string, gbo godal.IGenericBo) interface{} {
	// return map[string]interface{}{fieldAppId: gbo.GboGetAttrUnsafe(fieldAppId, reddo.TypeString)}
	return map[string]interface{}{_fieldId: gbo.GboGetAttrUnsafe(_fieldId, reddo.TypeString)}
}

// toBo transforms godal.IGenericBo to BoApp
func (dao *MongodbDaoApp) toBo(gbo godal.IGenericBo) *BoApp {
	if gbo == nil {
		return nil
	}
	bo := BoApp{}
	err := gbo.GboTransferViaJson(&bo)
	if err != nil {
		return nil
	}
	bo.Id = gbo.GboGetAttrUnsafe(_fieldId, reddo.TypeString).(string)
	return &bo
}

// toGbo transforms godal.IGenericBo to BoApp
func (dao *MongodbDaoApp) toGbo(bo *BoApp) godal.IGenericBo {
	if bo == nil {
		return nil
	}
	gbo := godal.NewGenericBo()
	err := gbo.GboImportViaJson(bo)
	if err != nil {
		return nil
	}
	gbo.GboSetAttr(_fieldId, bo.Id)
	gbo.GboSetAttr(fieldAppId, nil)
	return gbo
}

// Create implements IDaoApp.Create
func (dao *MongodbDaoApp) Create(bo *BoApp) (bool, error) {
	gbo := dao.toGbo(bo)
	if gbo == nil {
		return false, nil
	}
	numRows, err := dao.GdaoCreate(dao.collectionName, gbo)
	return numRows > 0, err
}

// Get implements IDaoApp.Get
func (dao *MongodbDaoApp) Get(id string) (*BoApp, error) {
	filter := map[string]interface{}{_fieldId: id}
	gbo, err := dao.GdaoFetchOne(dao.collectionName, filter)
	if err != nil || gbo == nil {
		return nil, err
	}
	return dao.toBo(gbo), nil
}

// GetAll implements IDaoApp.GetAll
func (dao *MongodbDaoApp) GetAll() ([]*BoApp, error) {
	sorting := map[string]int{fieldAppId: 1} // sort by "id" attribute, ascending
	rows, err := dao.GdaoFetchMany(dao.collectionName, nil, sorting, 0, 0)
	if err != nil {
		return nil, err
	}
	result := make([]*BoApp, 0)
	for _, e := range rows {
		bo := dao.toBo(e)
		if bo != nil {
			result = append(result, bo)
		}
	}
	return result, nil
}

// Update implements IDaoApp.Update
func (dao *MongodbDaoApp) Update(bo *BoApp) (bool, error) {
	gbo := dao.toGbo(bo)
	if gbo == nil {
		return false, nil
	}
	numRows, err := dao.GdaoUpdate(dao.collectionName, gbo)
	return numRows > 0, err
}

// Delete implements IDaoApp.Delete
func (dao *MongodbDaoApp) Delete(bo *BoApp) (bool, error) {
	gbo := dao.toGbo(bo)
	if gbo == nil {
		return false, nil
	}
	numRows, err := dao.GdaoDelete(dao.collectionName, gbo)
	return numRows > 0, err
}
