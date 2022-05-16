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
	"flag"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"net/http"
	"time"
)

var (
	rootCmd = &cobra.Command{
		Use:          "polaris-server-agent",
		Short:        "polaris-server-agent is used to clean the dirty instances",
		Long:         "polaris-server-agent is used to verify the node in platform and clean the dirty instances",
		SilenceUsage: true,
	}
)

/**
 * @brief 启动
 */
func Execute() error {
	_ = flag.Set("log_dir", "log")
	_ = flag.Set("stderrthreshold", "3")
	_ = flag.Set("logtostderr", "1")
	flag.Parse()

	glog.MaxSize = 1024 * 1024 * 256
	defer glog.Flush()

	go func() {
		go func() {
			//1秒刷一次日志
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					glog.Flush()
				}
			}
		}()
	}()

	go func() {
		// 默认用来调试的
		glog.Errorf("%s", http.ListenAndServe(":8088", nil))
	}()

	return rootCmd.Execute()
}
