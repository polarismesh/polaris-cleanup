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

package cleandeleted

import (
	"fmt"
	//数据库操作相关库
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/polarismesh/polaris-cleanup/common"
	"github.com/polarismesh/polaris-cleanup/job"
	"github.com/polarismesh/polaris-cleanup/store"
)

func init() {
	job.RegisterJob(&DeleteSoftDeleteInstanceJob{})
}

// DeleteSoftDeleteInstanceJob
type DeleteSoftDeleteInstanceJob struct {
	cfg common.AppConfig
}

func (job *DeleteSoftDeleteInstanceJob) Init(cfg common.AppConfig) {
	job.cfg = cfg
}

func (job *DeleteSoftDeleteInstanceJob) Name() string {
	return "DeleteSoftDeleteInstance"
}

func (job *DeleteSoftDeleteInstanceJob) Destory() error {
	return nil
}

// CronSpec
func (job *DeleteSoftDeleteInstanceJob) CronSpec() string {
	return "0 0 1 * * ?"
}

// Execute
func (job *DeleteSoftDeleteInstanceJob) Execute() func() {
	return func() {
		err := deleteSoftDeleteInstance(job.cfg)
		if nil != err {
			glog.Errorf("fail to soft  delete instance, err: %s", err.Error())
		}
	}
}

func deleteSoftDeleteInstance(cfg common.AppConfig) error {
	glog.Info("begin delete soft delete instance task")
	dbSource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Store.DbUser, cfg.Store.DbPwd,
		cfg.Store.DbHost, cfg.Store.DbPort, cfg.Store.DbName)
	db, err := store.NewPolarisDB(dbSource)
	if err != nil {
		glog.Errorf("[ERROR] new polaris db err: %s", err.Error())
	}
	defer db.Close()

	instances, err := db.LoadAllInvalidInstances(cfg.Cleanup.LimitedTime, cfg.Cleanup.LimitedNum)
	if err != nil {
		glog.Errorf("database load all invalid instances err: %s", err.Error())
		return err
	}

	glog.Infof("instances count: %d", len(instances))

	_ = iteratorInstance(db, cfg, instances)
	glog.Infof("successful delete %v", instances)
	glog.Info("delete soft delete instance task successful end")
	return nil
}

func iteratorInstance(db *store.PolarisDB, cfg common.AppConfig, deleteInstances []string) error {

	counter := 0
	batchDeleteNum := cfg.Cleanup.BatchDeleteNum
	if batchDeleteNum == 0 {
		batchDeleteNum = 100
	}

	var deleteIns []string

	for _, insId := range deleteInstances {
		deleteIns = append(deleteIns, insId)
		counter++
		if counter >= batchDeleteNum {
			err := db.CleanInvalidInstanceList(deleteIns)
			if err != nil {
				return fmt.Errorf("id:%s, fail to delete, err is %v", deleteIns, err)
			}
			counter = 0
			deleteIns = nil
			time.Sleep(time.Second)
		}
	}
	if counter > 0 {
		err := db.CleanInvalidInstanceList(deleteIns)
		if err != nil {
			return fmt.Errorf("id:%s, fail to delete, err is %v", deleteIns, err)
		}
	}

	return nil
}
