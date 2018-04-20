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
	"bufio"
	"bytes"
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var (
	Conf = Config{}
)

func LoadConfig(Path string) {
	// Read config.
	err, data := readFileInto(Path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &Conf)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	GateInfo     gateInfo
	LoginInfo    map[string]interface{}
	HallInfo     map[string]interface{}
	Game5GInfo   map[string]interface{}
	DataBaseInfo map[string]interface{}
}
type gateInfo struct {
	ClientListenPort string
	ServerListenPort string
	MaxConnNum       int
	PendingWriteNum  int
	TimeOut          int
}

func readFileInto(path string) (error, []byte) {
	var data []byte
	buf := new(bytes.Buffer)
	f, err := os.Open(path)
	if err != nil {
		return err, data
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			if len(line) > 0 {
				buf.Write(line)
			}
			break
		}
		//处理注释
		if !strings.HasPrefix(strings.TrimLeft(string(line), "\t "), "//") {
			buf.Write(line)
		}
	}
	data = buf.Bytes()
	//log.Info(string(data))
	return nil, data
}

// If read the file has an error,it will throws a panic.
func fileToStruct(path string, ptr *[]byte) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	*ptr = data
}
