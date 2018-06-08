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
	itemCfg = ItemConfig{Version: "0"}
	//锁
	itemlock = new(sync.RWMutex)

	itemmap = make(map[int]ItemsInfo)
)

func LoadItemConfig() {

	timer.AddCallback(time.Second*6, LoadItemConfig)

	ApplicationDir, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())
		return
	}

	confPath := fmt.Sprintf("%s/bin/conf/item.json", ApplicationDir)

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

	itemtmp := ItemConfig{}

	err = json.Unmarshal(data, &itemtmp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	//数据被改变
	if itemtmp.Version != itemCfg.Version {
		log.Info("itemCfg")
		itemlock.Lock()
		itemCfg = itemtmp
		FreshItems()
		itemlock.Unlock()
		log.Info("---version:%d", itemCfg.Version)

	}

}

func GetItemConfig() ItemConfig {
	itemlock.RLock()
	defer itemlock.RUnlock()
	return itemCfg
}

func GetItemAddTime(k int) int {
	itemlock.RLock()
	defer itemlock.RUnlock()

	log.Info("---k:%d", k)

	if v, ok := itemmap[k]; ok {
		log.Info("---addtime:%d", v.AddTime)
		return v.AddTime
	}

	return 0
}

func FreshItems() {
	itemmap = make(map[int]ItemsInfo)
	for _, v := range itemCfg.Items {

		itemmap[v.Id] = v
	}
}

type ItemsInfo struct {
	Id      int //商品ID
	AddTime int //增加时间
}

type ItemConfig struct {
	Version string //版本
	Items   []ItemsInfo
}
