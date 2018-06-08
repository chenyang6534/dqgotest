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
package app

import (
	"io"
	"io/ioutil"
	"net/http"
	//"time"
	//"encoding/json"
	"flag"
	"fmt"
	"os"
	//"os/exec"
	"os/signal"
	//"path/filepath"
	"dq/conf"
	"dq/datamsg"
	"dq/db"
	"dq/game5g"
	"dq/gate"
	"dq/hall"
	"dq/log"
	"dq/login"
	"dq/model"
	"errors"
	"strings"
	"sync"
)

type DefaultApp struct {
	//module.App

	settings conf.Config

	moduleNew func(modelType string) model.BaseModel

	databaseOne sync.Once
}

func (app *DefaultApp) Init() {
	app.moduleNew = func(modelType string) model.BaseModel {

		if modelType == datamsg.GateMode {
			a := &gate.Gate{
				MaxConnNum:      conf.Conf.GateInfo.MaxConnNum,
				PendingWriteNum: conf.Conf.GateInfo.PendingWriteNum,
				//TCPAddr:         conf.Conf.GateInfo.ClientListenPort,
				WSAddr:       conf.Conf.GateInfo.ClientListenPort,
				LocalTCPAddr: conf.Conf.GateInfo.ServerListenPort,
			}
			return a
		} else if modelType == datamsg.LoginMode {
			a := &login.Login{
				TCPAddr: conf.Conf.LoginInfo["GateIp"].(string),
			}
			app.databaseOne.Do(db.CreateDB)
			return a
		} else if modelType == datamsg.HallMode {
			a := &hall.Hall{
				TCPAddr: conf.Conf.HallInfo["GateIp"].(string),
			}
			app.databaseOne.Do(db.CreateDB)
			return a
		} else if modelType == datamsg.Game5GMode {
			a := &game5g.Game5G{
				TCPAddr: conf.Conf.Game5GInfo["GateIp"].(string),
			}
			app.databaseOne.Do(db.CreateDB)
			return a
		}

		return nil

	}
}

func (app *DefaultApp) Run() error {

	app.Init()

	mods := flag.String("models", "tt", "Log file directory?")
	flag.Parse() //解析输入的参数

	allModsName := strings.Split(*mods, ",")
	//app.processId = *ProcessID

	ApplicationDir, err := os.Getwd()
	if err != nil {
		return errors.New("cannot find dir")
	}

	confPath := fmt.Sprintf("%s/bin/conf/server.json", ApplicationDir)

	f, err := os.Open(confPath)
	if err != nil {
		panic(err)
	}
	Logdir := fmt.Sprintf("%s/bin/logs/%s", ApplicationDir, *mods)

	_, err = os.Open(Logdir)
	if err != nil {
		//文件不存在
		err := os.Mkdir(Logdir, os.ModePerm) //
		if err != nil {
			fmt.Println(err)
		}
	}

	conf.LoadConfig(f.Name()) //加载配置文件

	conf.LoadTaskEveryDayConfig() //加载每日任务

	conf.LoadStoreConfig() //加载商店信息

	conf.LoadSeasonConfig() //加载赛季信息

	conf.LoadItemConfig() //加载道具信息

	log.InitBeego(true, "dq", Logdir, make(map[string]interface{}))

	log.Info("dq starting up")

	log.Info("---models:%d", len(allModsName))
	// close
	closesig := make(chan bool, len(allModsName))
	// module

	var pro sync.WaitGroup

	for i := 0; i < len(allModsName); i++ {
		mod := app.moduleNew(allModsName[i])
		log.Info("new model :%s", allModsName[i])

		pro.Add(1)
		go func() {
			mod.Run(closesig)
			pro.Done()
		}()
	}

	if false {

		//http://127.0.0.1:8080/?a=123456&b=aaa&b1=bbb
		httpserver := &http.Server{Addr: ":9090", Handler: nil}

		http.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "hello world!")

			body, _ := ioutil.ReadAll(r.Body)
			//    r.Body.Close()
			body_str := string(body)
			fmt.Println(body_str)

			r.ParseForm()
			fmt.Println("Form: ", r.Form)
			fmt.Println("Path: ", r.URL.Path)
			fmt.Println(r.Form["a"])
			fmt.Println(r.Form["b"])
			for k, v := range r.Form {
				fmt.Println(k, "=>", v, strings.Join(v, "-"))
			}

			//httpserver.Close()
		})

		httpserver.ListenAndServe()
	} else {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		sig := <-c
		log.Debug("dq closing down (signal: %v) %d", sig, len(allModsName))
	}

	for i := 0; i < len(allModsName); i++ {

		closesig <- true
	}
	pro.Wait()
	fmt.Println("app over")

	return nil
}
