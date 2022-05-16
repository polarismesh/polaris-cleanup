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

// AppConfig agent configuration on startup
type AppConfig struct {
	//polaris_server数据库
	ServerDbHost    string   `yaml:"server_db_host"`
	ServerDbPort    int      `yaml:"server_db_port"`
	ServerUser      string   `yaml:"server_user"`
	ServerPwd       string   `yaml:"server_pwd"`
	ServerDbName    string   `yaml:"server_db_name"`
	ServiceToken    string   `yaml:"service_token"`
	PostAddr        string   `yaml:"Post-Address"`
	RequestId       string   `yaml:"Request-Id"`
	TtlTimes        int      `yaml:"ttl_times"`
	SoftLimitedTime int      `yaml:"soft_limited_time"`
	LimitedTime     int      `yaml:"delete_limitedTime"`
	LimitedNum      int      `yaml:"delete_limitedNum"`
	BatchDeleteNum  int      `yaml:"batch_delete_num"`
	InstanceId      string   `yaml:"instanceId"`
	CheckingService []string `yaml:"checkingService"`
}
