/**
* Licensed to the Apache Software Foundation (ASF) under one
* or more contributor license agreements. See the NOTICE file
* distributed with this work for additional information
* regarding copyright ownership. The ASF licenses this file
* to you under the Apache License, Version 2.0 (the
* "License"); you may not use this file except in compliance
* with the License. You may obtain a copy of the License at
*
*    http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing,
* software distributed under the License is distributed on an
* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
* KIND, either express or implied. See the License for the
* specific language governing permissions and limitations
* under the License.
*/

package dm303_go

import (
	"fmt"
	"os"
	"time"

	"github.com/DrWrong/monica/thrift"
	"github.com/DrWrong/monica/dm303_go/stats"
	"github.com/DrWrong/monica/config"

)

type DomobBase struct {
	Name    string
	Alive   int64
}

func NewDomobBase(name string) *DomobBase {
	return &DomobBase{name, time.Now().Unix()}
}

func (this *DomobBase) GetName() (string, error) {
	return this.Name, nil
}

func (this *DomobBase) GetVersion() (string, error) {
	return os.Getenv("VERSION"), nil
}

func (this *DomobBase) GetStatus() (DmStatus, error) {
	return DmStatus_ALIVE, nil
}

func (this *DomobBase) GetStatusDetails() (string, error) {
	return "everything is ok, you ho ho ho ho ho~~", nil
}

func (this *DomobBase) GetCounters() (map[string]int64, error) {
	return stats.GetStats(), nil
}

func (this *DomobBase) GetCounter(key string) (int64, error) {
	return stats.GetValue(key), nil
}

func (this *DomobBase) SetOption(key string, value string) error {
	return nil
}
func (this *DomobBase) GetOption(key string) (string, error) {
	return "", nil
}

func (this *DomobBase) GetOptions() (map[string]string, error) {
	return nil, nil
}

func (this *DomobBase) GetCpuProfile(profileDurationInSec int32) (string, error) {
	return "", nil
}

func (this *DomobBase) AliveSince() (int64, error) {
	return this.Alive, nil
}
func (this *DomobBase) Reinitialize() error {
	return nil
}

func (this *DomobBase) Shutdown() error {
	return nil
}

//启动dm303
func Start(port int, serviceName string) {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	networkAddr := fmt.Sprintf(":%d", port)
	serverTransport, err := thrift.NewTServerSocket(networkAddr)
	if err != nil {
		panic(err.Error())
	}
	handler := NewDomobBase(serviceName)
	processor := NewDomobServiceProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, serverTransport, transportFactory, protocolFactory)
	err = server.Serve()
	if err != nil {
		panic(err.Error())
	}
}


func MonicaDm303Start() {
	service_name := config.GlobalConfiger.String("dm303::service_name")
	port, _ := config.GlobalConfiger.Int("dm303::port")
	go Start(port, service_name)
}
