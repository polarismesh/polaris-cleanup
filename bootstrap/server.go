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

package bootstrap

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/polarismesh/polaris-cleanup/common"
	"github.com/polarismesh/polaris-cleanup/job"
)

func Run(filePath string) error {

	// 从环境变量中获取配置
	appConfig, err := common.LoadConfig(filePath)
	if err != nil {
		return err
	}

	glog.Infof("get config %v", appConfig)

	sc := common.NewDefaultScheduler()
	sc.Start()

	jobs := job.GetAllRegister()
	openJobs := appConfig.OpenJob

	for i := range openJobs {
		task, ok := jobs[openJobs[i]]
		if !ok {
			panic("JOB NOT FOUND - " + openJobs[i])
		}

		task.Init(*appConfig)
		if _, err = sc.AddJob(task); err != nil {
			return fmt.Errorf("add job=[%s] fail %+v", task.Name(), err)
		}
		glog.Infof("start job=[%s]", task.Name())
	}

	RunMainLoop(sc)
	return nil
}
