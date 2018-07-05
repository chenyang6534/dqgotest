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
	"dq/log"
	"dq/timer"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	turntableCfg = TurnTableConfig{Version: "0"}
	//锁
	turntablelock = new(sync.RWMutex)
)

func LoadTurnTableConfig() {

	timer.AddCallback(time.Second*5*60, LoadTurnTableConfig)

	ApplicationDir, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())
		return
	}

	confPath := fmt.Sprintf("%s/bin/conf/turntable.json", ApplicationDir)

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

	turntabletmp := TurnTableConfig{}

	err = json.Unmarshal(data, &turntabletmp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	//数据被改变
	if turntabletmp.Version != turntableCfg.Version {
		log.Info("turntableCfg")
		turntablelock.Lock()
		turntableCfg = turntabletmp
		turntablelock.Unlock()
		log.Info("---version:%d", turntableCfg.Version)

	}

}

func GetTurnTableConfig() TurnTableConfig {
	turntablelock.RLock()
	defer turntablelock.RUnlock()
	return turntableCfg
}

type TurnTableConfig struct {
	Version    string      //版本
	TurnTables []TurnTable //商品
	OnePrice   int         //单价
	TenPrice   int         //10次的价格
}

//"Num":0,
//            "InitLuckValue":10,
//            "Add":2,
type TurnTable struct {
	Id            int //商品ID
	Type          int //商品类型
	Time          int ////持续时间 以天为单位
	Num           int //数量
	InitLuckValue int //初始幸运值
	Add           int //增长幸运值
	Value         int //当前值
	Level         int //等级 0无意义 1表示(全服通知且中奖后初始化幸运值) 2表示(中奖后初始化值) 3表示(其他初始化的值到我身上)
}
