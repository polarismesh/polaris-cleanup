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
	"github.com/robfig/cron/v3"
)

// Scheduler
type Scheduler struct {
	cron *cron.Cron
}

// CronJob 定时任务的通用接口
type CronJob interface {
	CronSpec() string
	Execute() func()
}

// NewDefaultScheduler
func NewDefaultScheduler() *Scheduler {
	c := cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)), cron.WithParser(cron.NewParser(
		cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
	)))
	return &Scheduler{cron: c}
}

// AddJob
func (sc *Scheduler) AddJob(job CronJob) (int, error) {
	id, err := sc.cron.AddFunc(job.CronSpec(), job.Execute())
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// AddFunc
func (sc *Scheduler) AddFunc(cron string, cmd func()) (int, error) {
	id, err := sc.cron.AddFunc(cron, cmd)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Start
func (sc *Scheduler) Start() {
	sc.cron.Start()
}

// Stop
func (sc *Scheduler) Stop() {
	sc.cron.Stop()
}
