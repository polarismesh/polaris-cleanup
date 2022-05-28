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

package cleanempty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/polarismesh/polaris-cleanup/common"
)

type GetServiceInfo struct {
	Name                 string `json:"name"`
	Namespace            string `json:"namespace"`
	TotalInstanceCount   int32  `json:"total_instance_count"`
	HealthyInstanceCount int32  `json:"healthy_instance_count"`
}

type GetServiesResponse struct {
	Code     int              `json:"code"`
	Info     string           `json:"info"`
	Amount   int32            `json:"amount"`
	Size     int32            `json:"size"`
	Services []GetServiceInfo `json:"services"`
}

type DeleteServicesResponse struct {
	Code      int                     `json:"code"`
	Info      string                  `json:"info"`
	Size      int32                   `json:"size"`
	Responses []DeleteServiceResponse `json:"responses"`
}

type ServiceEntry struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type DeleteServiceResponse struct {
	Code    int          `json:"code"`
	Info    string       `json:"info"`
	Service ServiceEntry `json:"service"`
}

// DeleteEmptyServiceJob
type DeleteEmptyServiceJob struct {
	cfg common.AppConfig
}

func (job *DeleteEmptyServiceJob) Init(cfg common.AppConfig) {
	job.cfg = cfg
}

func (job *DeleteEmptyServiceJob) Name() string {
	return "DeleteEmptyService"
}

func (job *DeleteEmptyServiceJob) Destory() error {
	return nil
}

// CronSpec
func (job *DeleteEmptyServiceJob) CronSpec() string {
	return "0 0 * * * *"
}

// Execute
func (job *DeleteEmptyServiceJob) Execute() func() {
	return func() {
		job.deleteEmptyService(job.cfg)
	}
}

func (job *DeleteEmptyServiceJob) deleteEmptyService(cfg common.AppConfig) {
	emptyServices, err := job.getEmptyAutoCreatedServices()
	if err != nil {
		glog.Errorf("[DeleteEmptyService] fail to get services, %v", err)
		return
	}
	glog.Infof("[DeleteEmptyService] empty autoCreated services total count %d", len(emptyServices))

	deleteBatchSize := 10
	for i := 0; i < len(emptyServices); i += deleteBatchSize {
		j := i + deleteBatchSize
		if j > len(emptyServices) {
			j = len(emptyServices)
		}

		entries := convertServiceEntries(emptyServices[i:j])
		err := job.sendDeleteServicesRequest(entries)
		if err != nil {
			glog.Errorf("[DeleteEmptyService] failed to delete services, %v, %v", entries, err)
		}
	}
}

func convertServiceEntries(infos []GetServiceInfo) []ServiceEntry {
	var entries = make([]ServiceEntry, len(infos))
	for i, info := range infos {
		entries[i] = ServiceEntry{Namespace: info.Namespace, Name: info.Name}
	}
	return entries[:]
}

func (job *DeleteEmptyServiceJob) sendDeleteServicesRequest(entries []ServiceEntry) error {
	url := fmt.Sprintf("http://%s/naming/v1/services/delete", job.cfg.Server.ChooseOneEndpoint())

	reqBody, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	client := &http.Client{}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Request-Id", job.cfg.Server.RequestPrefix+uuid.New().String())
	request.Header.Set("Staffname", "空服务定时自动删除")
	request.Header.Set("X-Polaris-Token", job.cfg.Server.AuthToken)
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response DeleteServicesResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return err
	}

	for _, singleResp := range response.Responses {
		if singleResp.Code != 200000 {
			glog.Warningf("[DeleteEmptyService] fail to delete service, %s %s, code: %d, info:%s",
				singleResp.Service.Namespace, singleResp.Service.Name, singleResp.Code, singleResp.Info)
		} else {
			glog.Infof("[DeleteEmptyService] %s %s deleted",
				singleResp.Service.Namespace, singleResp.Service.Name)

		}

	}
	return nil
}

func (job *DeleteEmptyServiceJob) sendGetServicesRequest(query map[string]string) (*GetServiesResponse, error) {
	url := fmt.Sprintf("http://%s/naming/v1/services", job.cfg.Server.ChooseOneEndpoint())

	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := request.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	request.URL.RawQuery = q.Encode()

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Request-Id", job.cfg.Server.RequestPrefix+uuid.New().String())
	request.Header.Set("Staffname", "空服务定时自动查询")
	request.Header.Set("X-Polaris-Token", job.cfg.Server.AuthToken)
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to read response, err %v", err)
	}

	var response GetServiesResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if response.Code != 200000 {
		return nil, fmt.Errorf("fail to get services, query:%v, code:%d, info:%s",
			query, response.Code, response.Info)
	}

	return &response, nil
}

func (job *DeleteEmptyServiceJob) getEmptyAutoCreatedServices() ([]GetServiceInfo, error) {
	var emtpyServices []GetServiceInfo
	var offset int32 = 0
	var query = map[string]string{
		"keys":   "internal-auto-created",
		"values": "true",
		"limit":  "100",
		"offset": strconv.Itoa(int(offset))}

	for {
		resp, err := job.sendGetServicesRequest(query)
		if err != nil {
			return nil, err
		}

		for _, entry := range resp.Services {
			if entry.TotalInstanceCount == 0 {
				emtpyServices = append(emtpyServices, entry)
			}
		}

		nextOffset := offset + resp.Size
		if nextOffset >= resp.Amount {
			break
		}
		offset = nextOffset
		query["offset"] = strconv.Itoa(int(offset))
	}
	return emtpyServices, nil
}
