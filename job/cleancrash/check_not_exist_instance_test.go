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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/polarismesh/polaris-cleanup/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	instanceId = "ins-1234567"
	svcName    = "polaris.checker"
)

var (
	healthyHosts = []string{
		"127.0.0.1",
		"127.0.0.2",
		"127.0.0.3",
	}

	unhealthyHosts = []string{
		"127.0.0.4",
		"127.0.0.5",
	}

	ports = []int{
		8381, 8382,
	}
)

type mockInstance struct {
	Id        string `json:"id"`
	Service   string `json:"service"`
	Namespace string `json:"namespace"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Protocol  string `json:"protocol"`
	Healthy   bool   `json:"healthy"`
	Isolate   bool   `json:"isolate"`
}

type mockServerHandler struct {
	instances map[string]*mockInstance
	server    *http.ServeMux
}

func newMockServerHandler() *mockServerHandler {
	handler := &mockServerHandler{instances: make(map[string]*mockInstance)}
	var hosts []string
	hosts = append(hosts, healthyHosts...)
	hosts = append(hosts, unhealthyHosts...)
	for _, host := range hosts {
		for _, port := range ports {
			id := uuid.New().String()
			handler.instances[id] = &mockInstance{
				Id:        id,
				Service:   svcName,
				Namespace: "Polaris",
				Host:      host,
				Port:      port,
				Protocol:  "HTTP",
				Healthy:   true,
				Isolate:   false,
			}
		}
	}
	return handler
}

type mockResponse struct {
	Code      uint32          `json:"code"`
	Info      string          `json:"info"`
	Amount    int             `json:"amount"`
	Size      int             `json:"size"`
	Instances []*mockInstance `json:"instances"`
}

func (m *mockServerHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	response := &mockResponse{
		Code:   200000,
		Info:   "success",
		Amount: len(m.instances),
		Size:   len(m.instances),
	}
	for _, instance := range m.instances {
		response.Instances = append(response.Instances, instance)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	responseBuf, _ := json.Marshal(response)
	w.Write(responseBuf)
}

func (m *mockServerHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	str, err := ioutil.ReadAll(r.Body)
	if nil != err {
		log.Fatal(err)
	}
	instances := make([]*mockInstance, 0)
	err = json.Unmarshal(str, instances)
	if nil != err {
		log.Fatal(err)
	}
	for _, instance := range instances {
		fmt.Printf("delete instance %s, host %s\n", instance.Id, instance.Host)
		delete(m.instances, instance.Id)
	}
	w.WriteHeader(200)
}

func (m *mockServerHandler) startServer() {
	m.server = http.NewServeMux()
	m.server.HandleFunc("/naming/v1/instances", m.handleGet)
	m.server.HandleFunc("/naming/v1/instances/delete", m.handleDelete)
	go func() {
		err := http.ListenAndServe(":8090", m.server)
		if nil != err {
			log.Fatal(err)
		}
	}()
}

func TestNewCheckNotExistInstanceJob(t *testing.T) {
	fakeK8sClient := fake.NewSimpleClientset()
	var endpointAddresses []v1.EndpointAddress
	for _, healthyHost := range healthyHosts {
		endpointAddresses = append(endpointAddresses, v1.EndpointAddress{
			IP: healthyHost,
		})
	}
	_, _ = fakeK8sClient.CoreV1().Endpoints(instanceId).Create(&v1.Endpoints{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: workloadName,
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: endpointAddresses,
			},
		},
	})
	serverHandler := newMockServerHandler()
	serverHandler.startServer()

	checkJob := &CheckNotExistInstanceJob{
		appCfg: common.AppConfig{
			PostAddr:        "127.0.0.1",
			RequestId:       "11111",
			InstanceId:      instanceId,
			CheckingService: svcName,
		},
		cronSpec:    "",
		k8sClient:   fakeK8sClient,
		notReadyIps: make(map[string]bool, 0),
	}
	for i := 0; i < 2; i++ {
		checkJob.checkNotExistInstances()
	}
	healthyHostsMap := make(map[string]bool)
	for _, host := range healthyHosts {
		healthyHostsMap[host] = true
	}
	for _, instance := range serverHandler.instances {
		if _, ok := healthyHostsMap[instance.Host]; !ok {
			t.Fatalf("host %s not found", instance.Host)
		}
	}
}
