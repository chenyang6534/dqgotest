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
	taskEveryDayCfg = TaskEveryDayConfig{Version: "0"}
	//锁
	lock = new(sync.RWMutex)
)

func LoadTaskEveryDayConfig() {

	timer.AddCallback(time.Second*5, LoadTaskEveryDayConfig)

	ApplicationDir, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())
		return
	}

	confPath := fmt.Sprintf("%s/bin/conf/taskeveryday.json", ApplicationDir)

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

	ted := TaskEveryDayConfig{}

	err = json.Unmarshal(data, &ted)
	if err != nil {
		log.Error(err.Error())
		return
	}
	//数据被改变
	if ted.Version != taskEveryDayCfg.Version {
		log.Info("LoadTaskEveryDayConfig")
		lock.Lock()
		taskEveryDayCfg = ted
		lock.Unlock()
		log.Info("---version:%d", taskEveryDayCfg.Version)
	}

}

func GetTaskEveryDayCfg() *TaskEveryDayConfig {
	lock.RLock()
	defer lock.RUnlock()
	return &taskEveryDayCfg
}

type TaskEveryDayConfig struct {
	Version string //版本
	Task    []TaskConfig
}

type TaskConfig struct {
	Id        int
	Type      int
	DestValue int //目标

	DBTableName         string //数据库表名
	GetTagDBFieldName   string //是否已经领取
	ProgressDBFieldName string //进度在数据库的字段名字
	InitialValue        int    //初始值
}
