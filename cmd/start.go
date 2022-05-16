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

package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/kelseyhightower/envconfig"
	"github.com/polarismesh/polaris-cleanup/common"
	"github.com/polarismesh/polaris-cleanup/job/cleancrash"
	"github.com/polarismesh/polaris-cleanup/job/cleandeleted"
	"github.com/polarismesh/polaris-cleanup/job/cleanunhealthy"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start polaris agent",
	Long:  "this command start polaris agent",
	Run:   startCommand,
}

func startCommand(cmd *cobra.Command, args []string) {
	// 从环境变量中获取配置
	var appConfig common.AppConfig
	if err := envconfig.Process("", &appConfig); err != nil {
		log.Fatal("exit with loading config error. ", err)
	}

	glog.Infof("get config %v", appConfig)

	sc := common.NewDefaultScheduler()
	sc.Start()
	defer sc.Stop()

	// 初始化任务
	err := initJob(sc, appConfig)
	if err != nil {
		log.Fatal("add micro service metrics job error. ", err)
	}
	glog.Info("polaris agent starting...")

	// wait kill
	waitKillSignal()
}

func initJob(sc *common.Scheduler, cfg common.AppConfig) error {

	var err error
	// 清理正常实例的接口有问题，先下掉了
	// 10 min 运行一次
	// var cron = "0 */10 * * * ?"
	// deleteNotExistHealthyInstanceJob := job.NewDeleteNotExistHealthyInstanceJob(cron, cfg)

	// 每天一点运行一次
	cron := "0 0 1 * * ?"
	deleteSoftDeleteInstanceJob := cleandeleted.NewDeleteSoftDeleteInstanceJob(cron, cfg)

	// 每天两点运行一次
	cron = "0 0 2 * * ?"
	deleteUnhealthyInstanceJob := cleanunhealthy.NewDeleteUnHealthyInstanceJob(cron, cfg)

	// 每10秒运行一次
	cron = "*/30 * * * * ?"
	checkNotExistInstanceJob, err := cleancrash.NewCheckNotExistInstanceJob(cron, cfg)
	if err != nil {
		return fmt.Errorf("create check not exist instance job error %s", err)
	}

	//_, err := sc.AddJob(deleteNotExistHealthyInstanceJob)
	//if err != nil {
	//	return fmt.Errorf("add delete not exist healthy instance job error %s", err)
	//}

	_, err = sc.AddJob(deleteSoftDeleteInstanceJob)
	if err != nil {
		return fmt.Errorf("delete soft delete instance job %s", err)
	}

	_, err = sc.AddJob(deleteUnhealthyInstanceJob)
	if err != nil {
		return fmt.Errorf("delete unhealthy instance job %s", err)
	}
	_, err = sc.AddJob(checkNotExistInstanceJob)
	if err != nil {
		return fmt.Errorf("add check not exist instance job error %s", err)
	}
	return nil
}

func waitKillSignal() {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown polaris agent server ...")
}
