/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package common

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// AppConfig agent configuration on startup
type AppConfig struct {
	InstanceId string  `yaml:"instanceId"`
	Store      Store   `yaml:"store"`
	Server     Server  `yaml:"server"`
	Cleanup    Cleanup `yaml:"cleanUp"`
}

type Server struct {
	Endpoints     []string `yaml:"endpoints"`
	AuthToken     string   `yaml:"authToken"`
	RequestPrefix string   `yaml:"requestPrefix"`
}

type Cleanup struct {
	LimitedTime    int `yaml:"deleteLimitedTime"`
	LimitedNum     int `yaml:"deleteLimitedNum"`
	BatchDeleteNum int `yaml:"batchDeleteNum"`
}

type Store struct {
	DbHost string `yaml:"dbHost"`
	DbPort int    `yaml:"dbPort"`
	DbName string `yaml:"dbName"`
	DbUser string `yaml:"dbUser"`
	DbPwd  string `yaml:"dbPwd"`
}

func LoadConfig(filePath string) (*AppConfig, error) {
	if filePath == "" {
		err := errors.New("invalid config file path")
		fmt.Printf("[ERROR] %v\n", err)
		return nil, err
	}

	fmt.Printf("[INFO] load config from %v\n", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return nil, err
	}
	defer file.Close()

	config := &AppConfig{}
	err = yaml.NewDecoder(file).Decode(config)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return nil, err
	}

	return config, nil
}
