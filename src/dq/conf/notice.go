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
	noticeCfg = NoticeConfig{Version: "0"}
	//锁
	noticelock = new(sync.RWMutex)
)

func LoadNoticeConfig() {

	timer.AddCallback(time.Second*6, LoadNoticeConfig)

	ApplicationDir, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())
		return
	}

	confPath := fmt.Sprintf("%s/bin/conf/notice.json", ApplicationDir)

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

	noticetmp := NoticeConfig{}

	err = json.Unmarshal(data, &noticetmp)
	if err != nil {
		log.Error(err.Error())
		return
	}
	//数据被改变
	if noticetmp.Version != noticeCfg.Version {
		log.Info("noticeCfg")
		noticelock.Lock()
		noticeCfg = noticetmp
		noticelock.Unlock()
		log.Info("---version:%d----GameOverShare:%d", noticeCfg.Version, noticeCfg.GameOverShare)

	}

}

func GetNoticeConfig() NoticeConfig {
	noticelock.RLock()
	defer noticelock.RUnlock()
	return noticeCfg
}

type NoticeConfig struct {
	Version       string //版本
	Notice        string
	GameOverShare int //游戏结束后分享 1表示可以分享 0表示不能分享
	LookViewCount int //观看视频免费抽次数
}
