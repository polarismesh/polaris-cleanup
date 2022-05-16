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

package cleancrash

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/polarismesh/polaris-cleanup/common"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CheckNotExistInstanceJob 检查不存在的实例，并进行删除
type CheckNotExistInstanceJob struct {
	appCfg      common.AppConfig
	cronSpec    string
	k8sClient   kubernetes.Interface
	notReadyIps map[string]bool
}

// NewCheckNotExistInstanceJob 构造函数
func NewCheckNotExistInstanceJob(spec string, appCfg common.AppConfig) (*CheckNotExistInstanceJob, error) {
	if err := validateAppConfig(appCfg); nil != err {
		return nil, err
	}
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &CheckNotExistInstanceJob{
		appCfg:      appCfg,
		cronSpec:    spec,
		k8sClient:   k8sClient,
		notReadyIps: make(map[string]bool, 0),
	}, nil
}

func validateAppConfig(appCfg common.AppConfig) error {
	if len(appCfg.PostAddr) == 0 {
		return errors.New("post_addr must not be empty")
	}
	if len(appCfg.InstanceId) == 0 {
		return errors.New("instance_id must not be empty")
	}
	if len(appCfg.CheckingService) == 0 {
		return errors.New("checking_service must not be empty")
	}
	return nil
}

// CronSpec
func (job *CheckNotExistInstanceJob) CronSpec() string {
	return job.cronSpec
}

// Execute
func (job *CheckNotExistInstanceJob) Execute() func() {
	return job.checkNotExistInstances
}

const (
	polarisNamespace = "Polaris"
	workloadName     = "polaris"
)

func (job *CheckNotExistInstanceJob) checkNotExistInstances() {
	glog.Info("[Agent][CheckNotExistInstance]begin to delete not exists instance task")
	endPoints, err := job.k8sClient.CoreV1().Endpoints(job.appCfg.InstanceId).Get(workloadName, metav1.GetOptions{})
	if nil != err {
		glog.Infof("[Agent][CheckNotExistInstance]fail to list endpoints, err is %v", err)
		return
	}
	readyPodIps := make(map[string]bool, 0)
	subSets := endPoints.Subsets
	if len(subSets) > 0 {
		for _, subset := range subSets {
			addresses := subset.Addresses
			if len(addresses) == 0 {
				continue
			}
			for _, address := range addresses {
				readyPodIps[address.IP] = true
			}
		}
	}
	glog.Infof("[Agent][CheckNotExistInstance]ready pod endpoints are %v, not ready pod endpoints are %v",
		readyPodIps, job.notReadyIps)
	if len(job.notReadyIps) > 0 {
		for notReadyIp := range job.notReadyIps {
			if _, ok := readyPodIps[notReadyIp]; ok {
				delete(job.notReadyIps, notReadyIp)
			}
		}
	}
	ipToInstances, err := job.getInstancesFromServer(
		job.appCfg.PostAddr, job.appCfg.RequestId, polarisNamespace, job.appCfg.CheckingService)
	if nil != err {
		glog.Infof("[Agent][CheckNotExistInstance]fail to get instances, err is %v", err)
		return
	}
	if len(ipToInstances) == 0 {
		glog.Infof("[Agent][CheckNotExistInstance]instances from polaris server is 0")
		return
	}

	ipToInstances = job.removeNoSelfInstances(ipToInstances)

	for host, instances := range ipToInstances {
		if _, ok := readyPodIps[host]; ok {
			continue
		}
		if _, ok := job.notReadyIps[host]; !ok {
			// 首次出现不健康，则先加入列表，不做操作
			job.notReadyIps[host] = true
			glog.Infof("[Agent][CheckNotExistInstance]host %s not exists, add to pending list", host)
			continue
		}
		glog.Infof("[Agent][CheckNotExistInstance]host %s not exists, start to clean", host)
		err = job.deleteInstancesFromServer(job.appCfg.PostAddr, job.appCfg.RequestId, instances)
		if nil != err {
			glog.Infof("[Agent][CheckNotExistInstance]fail to delete instances by ip %s, err is %v", host, err)
			continue
		}
	}
	//清理not ready map里面的脏数据
	for host := range job.notReadyIps {
		_, inK8sReady := readyPodIps[host]
		if inK8sReady {
			delete(job.notReadyIps, host)
			continue
		}
		_, inServer := ipToInstances[host]
		if !inServer {
			delete(job.notReadyIps, host)
		}
	}
}

func (job *CheckNotExistInstanceJob) removeNoSelfInstances(ipToInstances map[string][]*InstanceInfo) map[string][]*InstanceInfo {
	engineID := job.appCfg.InstanceId

	metadataKey := "polaris_cluster_id"

	ret := make(map[string][]*InstanceInfo)

	for ip, instances := range ipToInstances {
		for i := range instances {
			ins := instances[i]

			val, ok := ins.Metadata[metadataKey]
			if ok && val != engineID {
				continue
			}

			if _, ok := ret[ip]; !ok {
				ret[ip] = make([]*InstanceInfo, 0)
			}

			ret[ip] = append(ret[ip], ins)
		}
	}

	return ret
}

// InstancesResponse
type InstancesResponse struct {
	Code      uint32          `json:"code"`
	Amount    int             `json:"amount"`
	Size      int             `json:"size"`
	Instances []*InstanceInfo `json:"instances"`
}

// InstanceInfo
type InstanceInfo struct {
	Id       string            `json:"id"`
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Metadata map[string]string `json:"metadata"`
}

// String
func (i InstanceInfo) String() string {
	return fmt.Sprintf("{id=%s, host=%s, port=%d}", i.Id, i.Host, i.Port)
}

//设置HTTP请求的请求参数
func buildDeleteParam(instances []*InstanceInfo) ([]byte, error) {
	params, err := json.Marshal(instances)
	if err != nil {
		return nil, fmt.Errorf("marshall failed:%s", err)
	}
	return params, nil
}

// Response 删除服务实例接口的回复
type Response struct {
	Code int    `json:"code"`
	Info string `json:"info"`
}

func (job *CheckNotExistInstanceJob) deleteInstancesFromServer(postAddr string, requestId string,
	instances []*InstanceInfo) error {

	deleteAddress := fmt.Sprintf("http://%s:8090/naming/v1/instances/delete", postAddr)
	deleteInstancesBody, err := buildDeleteParam(instances)
	if err != nil {
		return err
	}
	client := &http.Client{}
	request, err := http.NewRequest("POST", deleteAddress, bytes.NewBuffer(deleteInstancesBody))
	if err != nil {
		return fmt.Errorf("fail to create request, err %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Request-Id", requestId)
	request.Header.Set("Staffname", "自动清理不存在的pod实例")
	request.Header.Set("X-Polaris-Token", job.appCfg.XPolarisToken)
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("fail to get response, err %v", err)
	}
	defer resp.Body.Close()

	//若成功，则不返回body
	if resp.StatusCode == 200 {
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
	return fmt.Errorf("fail to delete the instances, code:%d, info:%s", jsonRsp.Code, jsonRsp.Info)
}

const limit = 100

func (job *CheckNotExistInstanceJob) getInstancesFromServer(
	postAddr string, requestId string, namespace string, service string) (map[string][]*InstanceInfo, error) {

	values := make(map[string][]*InstanceInfo)
	startOffset := 0
	for {
		nextOffset, err := func(offset int) (int, error) {
			queryAddress := fmt.Sprintf("http://%s:8090/naming/v1/instances", postAddr)
			//发送 GET 请求，获取实例上一次心跳时间
			client := &http.Client{}
			request, err := http.NewRequest(http.MethodGet, queryAddress, nil)
			if err != nil {
				return -1, fmt.Errorf("fail to create request, err %v", err)
			}

			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Request-Id", requestId)
			request.Header.Set("Staffname", "自动清理不存在的pod实例")
			request.Header.Set("X-Polaris-Token", job.appCfg.XPolarisToken)

			param := request.URL.Query()
			param.Add("service", service)
			param.Add("namespace", namespace)
			param.Add("offset", strconv.Itoa(offset))
			param.Add("limit", strconv.Itoa(limit))

			request.URL.RawQuery = param.Encode()

			resp, err := client.Do(request)
			if err != nil {
				return -1, fmt.Errorf("fail to get response, err %v", err)
			}
			defer resp.Body.Close()

			//若成功，则不返回body
			if resp.StatusCode != 200 {
				errorMsg := fmt.Sprintf("failed to get instances from service(ns=%s, name=%s), code: %d",
					namespace, service, resp.StatusCode)
				return -1, errors.New(errorMsg)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return -1, fmt.Errorf("fail to read response, err %v", err)
			}

			instancesResp := &InstancesResponse{}
			if err = json.Unmarshal(body, instancesResp); err != nil {
				return -1, err
			}
			if instancesResp.Size > 0 {
				for _, instance := range instancesResp.Instances {
					instanceInfos, ok := values[instance.Host]
					if !ok {
						values[instance.Host] = []*InstanceInfo{instance}
					} else {
						instanceInfos = append(instanceInfos, instance)
						values[instance.Host] = instanceInfos
					}
				}
			}
			totalAmount := instancesResp.Amount
			nextStart := offset + instancesResp.Size
			if totalAmount <= nextStart {
				return 0, nil
			}
			return nextStart, nil
		}(startOffset)
		if nil != err {
			return nil, err
		}
		if nextOffset <= 0 {
			return values, nil
		}
	}
}
