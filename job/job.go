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

package job

import (
	"github.com/polarismesh/polaris-cleanup/common"
)

var (
	slots map[string]PolarisCleanJob = map[string]PolarisCleanJob{}
)

func RegisterJob(j PolarisCleanJob) {
	slots[j.Name()] = j
}

func GetAllRegister() map[string]PolarisCleanJob {
	ret := make(map[string]PolarisCleanJob)

	for k, v := range slots {
		ret[k] = v
	}

	return ret
}

type PolarisCleanJob interface {
	Init(cfg common.AppConfig)
	CronSpec() string
	Execute() func()
	Name() string
	Destory() error
}
