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
	seasonCfg = SeasonConfig{Version: "0"}
	//锁
	seasonlock = new(sync.RWMutex)
)

func LoadSeasonConfig() {

	timer.AddCallback(time.Second*5, LoadSeasonConfig)

	ApplicationDir, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())
		return
	}

	confPath := fmt.Sprintf("%s/bin/conf/season.json", ApplicationDir)

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

	seasontmp := SeasonConfig{}

	err = json.Unmarshal(data, &seasontmp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	//数据被改变
	if seasontmp.Version != seasonCfg.Version {
		log.Info("seasonCfg")
		seasonlock.Lock()
		seasonCfg = seasontmp
		seasonlock.Unlock()
		log.Info("---version:%d", seasonCfg.Version)

	}

}

func GetSeasonConfig() SeasonConfig {
	seasonlock.RLock()
	defer seasonlock.RUnlock()
	return seasonCfg
}

type SeasonConfig struct {
	Version string   //版本
	Seasons []Season //赛季信息
}

//赛季奖励
type SeasonReward struct {
	RankStart int
	RankEnd   int
	Gold      int
}

//赛季信息
type Season struct {
	IdIndex    int    //商品ID
	StartTime  string //开卖时间
	EndTime    string //结束卖的时间
	RewardList []SeasonReward
}
