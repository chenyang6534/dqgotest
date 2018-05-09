// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conf

import (
	//"bufio"
	//"bytes"
	"dq/log"
	"dq/timer"
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"os"
	//"strings"
	"sync"
	"time"
)

var (
	storeCfg = StoreConfig{Version: "0"}
	//锁
	storelock = new(sync.RWMutex)
)

func LoadStoreConfig() {

	timer.AddCallback(time.Second*5, LoadStoreConfig)

	ApplicationDir, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())
		return
	}

	confPath := fmt.Sprintf("%s/bin/conf/store.json", ApplicationDir)

	f, err := os.Open(confPath)
	if err != nil {
		log.Error(err.Error())
		return
	}

	// Read config.
	err, data := readFileInto(f.Name())
	if err != nil {
		log.Error(err.Error())
		return
	}

	storetmp := StoreConfig{}

	err = json.Unmarshal(data, &storetmp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	//数据被改变
	if storetmp.Version != storeCfg.Version {
		log.Info("storeCfg")
		storelock.Lock()
		storeCfg = storetmp
		storelock.Unlock()
		log.Info("---version:%d", storeCfg.Version)

	}

}

func GetStoreConfig() StoreConfig {
	storelock.RLock()
	defer storelock.RUnlock()
	return storeCfg
}

type StoreConfig struct {
	Version    string      //版本
	Commoditys []Commodity //商品
}

type Commodity struct {
	Id           int    //商品ID
	Type         int    //商品类型
	Time         int    ////持续时间 以天为单位
	SaleGoldType int    //售价类型 1表示金币
	SalePrice    int    //售价
	SaleDiscount int    //折扣 10折表示不打折
	StartTime    string //开卖时间
	EndTime      string //结束卖的时间
}
