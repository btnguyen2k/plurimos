package mom

import (
	"encoding/json"
	"github.com/pkg/errors"
	"log"
	"strings"
	"time"
)

/*
MOM's business objects.

@author Thanh Nguyen <btnguyen2k@gmail.com>
@since 0.1.0
*/

const (
	fieldMapNamespace = "ns"
	fieldMapFrom      = "frm"
	fieldMapTo        = "to"
	fieldMapTime      = "t"
	fieldMapAppId     = "app"
)

/*
BoMapping defines a mapping record
*/
type BoMapping struct {
	Namespace string    `json:"ns"`
	From      string    `json:"frm"`
	To        string    `json:"to"`
	Time      time.Time `json:"t"`
	AppId     string    `json:"app"`
}

func (bo *BoMapping) FromMap(data map[string]interface{}) *BoMapping {
	js, _ := json.Marshal(data)
	err := json.Unmarshal(js, bo)
	if err != nil {
		log.Printf("Error while unmarshal from map: %e", err)
	}
	return bo
}

func (bo *BoMapping) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	js, _ := json.Marshal(bo)
	err := json.Unmarshal(js, &data)
	if err != nil {
		log.Printf("Error while unmarshal from bo: %e", err)
	}
	return data
}

/*
IDaoMoMapping defines many-to-one mapping API.

A mapping is a direction {from/object -> to/target} (e.g. object maps to target}

	- An object can map to maximum one target
	- A target can be mapped by many objects
*/
type IDaoMoMapping interface {
	/*
	   InitStorage initializes storage to store an app's mappings.
	*/
	InitStorage(appId string) error

	/*
	   DestroyStorage cleans up storage allocated to store an app's mappings.
	*/
	DestroyStorage(appId string) error

	/*
		FindTargetForObject is given an object, finds the target of direction {object -> target}.
	*/
	FindTargetForObject(appId, namespace, object string) (*BoMapping, error)

	/*
		FindObjectsToTargets is given a target, finds all the objects of direction {target <- objects}.
	*/
	FindObjectsToTarget(appId, namespace, target string) (map[string]*BoMapping, error)

	/*
		Map maps object to target.
		Map is successful if and only if:

		    - 'object' has not mapped to any target, or
		    - 'object' had mapped to the target
	*/
	Map(appId, namespace, object, target string) (*BoMapping, error)

	/*
		MapIfTargetExists maps object to target.
		MapIfTargetExists is successful if and only if:

		    - 'target' exists, and
		      - 'object' has not mapped to any target, or
		      - 'object' had mapped to the target
	*/
	MapIfTargetExists(appId, namespace, object, target string) (*BoMapping, error)
}

/*----------------------------------------------------------------------*/

const (
	fieldAppId     = "id"
	fieldAppSecret = "sec"
	fieldAppTime   = "t"
	fieldAppConfig = "cfg"
)

var (
	appSystem = "system"
)

/*
BoApp defines an application record
*/
type BoApp struct {
	Id     string                 `json:"id"`
	Secret string                 `json:"sec"`
	Time   time.Time              `json:"t"`
	Config map[string]interface{} `json:"cfg"`
}

/*
IDaoMoApp defines API to access application storage.
*/
type IDaoApp interface {
	// Create persists a new app to database storage. If the app already existed, this function returns (false, nil)
	Create(bo *BoApp) (bool, error)

	// Get finds an app by id & fetches it from database storage.
	Get(id string) (*BoApp, error)

	// GetAll retrieves all available apps from database storage and returns them as a list.
	GetAll() ([]*BoApp, error)

	// Update modifies an existing app in the database storage. If the app does not exist in database, this function returns (false, nil).
	Update(bo *BoApp) (bool, error)

	// Delete removes a app from database storage.
	Delete(bo *BoApp) (bool, error)
}

/*----------------------------------------------------------------------*/

func initData(dbtype string) error {
	if strings.EqualFold("mongo", dbtype) || strings.EqualFold("mongodb", dbtype) {
		ok, err := mongoConnect.HasCollection(collectionApps)
		if err != nil {
			return err
		}
		if !ok {
			log.Printf("Creating collection [%s]...", collectionApps)
			dbResult, err := mongoConnect.CreateCollection(collectionApps)
			if err != nil {
				return err
			}
			if dbResult.Err() != nil {
				return dbResult.Err()
			}
			// dbResult, err = mongoConnect.CreateIndexes(collectionApps, []interface{}{
			//     map[string]interface{}{
			//         "key":    map[string]interface{}{fieldAppId: 1},
			//         "name":   "uidx_id",
			//         "unique": true,
			//     },
			// })
			// if err != nil {
			//     return err
			// }
			// if dbResult.Err() != nil {
			//     return dbResult.Err()
			// }
		}
		return nil
	}
	return errors.Errorf("Unknown database type: [%s].", dbtype)
}
