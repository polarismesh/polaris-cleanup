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
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/polarismesh/polaris-cleanup/common"
)

var darwinSignals = []os.Signal{
	syscall.SIGINT, syscall.SIGTERM,
	syscall.SIGSEGV, syscall.SIGUSR1,
}

// RunMainLoop server主循环
func RunMainLoop(sc *common.Scheduler) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, darwinSignals...)
	for {
		select {
		case s := <-ch:
			sc.Stop()
			glog.Infof("catch signal(%+v), stop servers", s)
			return
		}
	}
}
