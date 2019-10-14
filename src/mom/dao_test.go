package mom

import (
	"time"
)

const (
	_testAppId     = "test"
	_testAppSecret = "tests3cr3t"
	_testAppDesc   = "Test application"
)

var _testApp = &BoApp{
	Id:     _testAppId,
	Secret: _testAppSecret,
	Time:   time.Now(),
	Config: map[string]interface{}{
		"desc":        _testAppDesc,
		"config_bool": true,
		"config_int":  103,
	},
}
