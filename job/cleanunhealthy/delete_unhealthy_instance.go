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

package cleanunhealthy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/polarismesh/polaris-cleanup/common"
	"github.com/polarismesh/polaris-cleanup/job"
	"github.com/polarismesh/polaris-cleanup/store"

	//数据库操作相关库
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
)

func init() {
	job.RegisterJob(&DeleteUnHealthyInstanceJob{})
}

// DeleteUnHealthyInstanceJob
type DeleteUnHealthyInstanceJob struct {
	cfg common.AppConfig
}

func (job *DeleteUnHealthyInstanceJob) Init(cfg common.AppConfig) {
	job.cfg = cfg
}

func (job *DeleteUnHealthyInstanceJob) Name() string {
	return "DeleteUnHealthyInstance"
}

func (job *DeleteUnHealthyInstanceJob) Destory() error {
	return nil
}

// CronSpec
func (job *DeleteUnHealthyInstanceJob) CronSpec() string {
	return "0 0 2 * * ?"
}

// Execute
func (job *DeleteUnHealthyInstanceJob) Execute() func() {
	return func() {
		job.deleteUnHealthInstance(job.cfg)
	}
}

// Response 删除服务实例接口的回复
type Response struct {
	Code int    `json:"code"`
	Info string `json:"info"`
}

// PolarisInstance 用来进行序列化和反序列化的实例结构体
type PolarisInstance struct {
	Id      string `json:"id"`
	Service string `json:"service"`

	HealthCheck struct {
		HeartbeatType string `json:"type"`
		Heartbeat     struct {
			Ttl int `json:"ttl"`
		} `json:"heartbeat"`
	} `json:"healthCheck"`

	Metadata struct {
		LastHeartbeatTime string `json:"last-heartbeat-time"`
	} `json:"metadata"`
}

// InstanceHeartbeatInfo 实例心跳数据
type InstanceHeartbeatInfo struct {
	Code            int             `json:"code"`
	Info            string          `json:"info"`
	PolarisInstance PolarisInstance `json:"instance"`
}

// ResponseParam 删除服务实例接口的请求参数
type ResponseParam struct {
	ServiceToken string `json:"service_token"`
	Id           string `json:"id"`
}

// getParams 设置HTTP请求的请求参数
func getParams(ids []string, cfg common.AppConfig) ([]byte, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("delete: the id is NUll")
	}

	var responses []ResponseParam
	for _, id := range ids {
		resp := ResponseParam{
			Id: id,
		}
		responses = append(responses, resp)

	}

	params, err := json.Marshal(responses)
	if err != nil {
		return nil, fmt.Errorf("marshall failed:%s", err)
	}
	return params, nil
}

func (job *DeleteUnHealthyInstanceJob) sendHttpRequest(deleteInstances []string, cfg common.AppConfig) error {

	endpoint := cfg.Server.Endpoints[rand.Intn(len(cfg.Server.Endpoints))]

	deleteAddress := fmt.Sprintf("http://%s/naming/v1/instances/delete", endpoint)
	//发送POST请求，删除过期的实例
	deleteInstancesBody, err := getParams(deleteInstances, cfg)
	if err != nil {
		return err
	}
	client := &http.Client{}
	request, err := http.NewRequest("POST", deleteAddress, bytes.NewBuffer(deleteInstancesBody))
	if err != nil {
		return fmt.Errorf("fail to create request, err %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Request-Id", cfg.Server.RequestPrefix+uuid.New().String())
	request.Header.Set("Staffname", "异常实例定时自动删除")
	request.Header.Set("X-Polaris-Token", job.cfg.Server.AuthToken)
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("fail to get response, err %v", err)
	}
	defer resp.Body.Close()

	//若成功，则不返回body
	if resp.StatusCode == 200 {
		glog.Infof("success to delete the instance, id:%s", deleteInstances)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fail to read response, err %v", err)
	}
	var jsonRsp Response
	if err = json.Unmarshal(body, &jsonRsp); err != nil {
		return err
	}
	return fmt.Errorf("fail to delete the instance, id:%s, code:%d, info:%s",
		deleteInstances, jsonRsp.Code, jsonRsp.Info)
}

func (job *DeleteUnHealthyInstanceJob) deleteUnHealthInstance(cfg common.AppConfig) error {
	glog.Info("begin delete unhealthy instance task")
	selectReportSQL := "SELECT id FROM instance where flag=0 and enable_health_check=1 " +
		"and health_status=0 and mtime <= DATE_SUB(NOW(), INTERVAL ? MINUTE) limit ?"
	var deleteInstances []string
	var err error

	rows, err := store.GetStore().GetDB().Query(selectReportSQL, cfg.Cleanup.LimitedTime, cfg.Cleanup.LimitedNum)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var expiredId string
		err = rows.Scan(&expiredId)
		if err != nil {
			return fmt.Errorf("fail to read data from t_instance_unhealth, err is %v", err)
		}
		deleteInstances = append(deleteInstances, expiredId)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("fetch rows next err:%s", err)
	}

	counter := 0
	batchDeleteNum := cfg.Cleanup.BatchDeleteNum
	if batchDeleteNum == 0 {
		batchDeleteNum = 100
	}

	var deleteIns []string

	for _, v := range deleteInstances {
		deleteIns = append(deleteIns, v)
		counter++
		if counter >= batchDeleteNum {
			err = job.sendHttpRequest(deleteIns, cfg)
			if err != nil {
				return fmt.Errorf("id:%s, fail to delete, err is %v", deleteIns, err)
			}
			counter = 0
			deleteIns = nil
			time.Sleep(time.Second)
		}
	}
	if counter > 0 {
		err = job.sendHttpRequest(deleteIns, cfg)
		if err != nil {
			return fmt.Errorf("id:%s, fail to delete, err is %v", deleteIns, err)
		}
	}
	if len(deleteInstances) == 0 {
		return fmt.Errorf("there is no instance to delete")
	}
	glog.Infof("successful delete count %d, list: %v", len(deleteInstances), deleteInstances)
	glog.Info("delete unhealthy instance task successful end")
	return nil
}
