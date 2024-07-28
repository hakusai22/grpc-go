/*
 *
 * Copyright 2017 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package roundrobin defines a roundrobin balancer. Roundrobin balancer is
// installed as one of the default balancers in gRPC, users don't need to
// explicitly install this balancer.
package roundrobin

import (
	"math/rand"
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
)

// Name is the name of round_robin balancer.
const Name = "round_robin"

var logger = grpclog.Component("roundrobin")

// newBuilder 函数创建一个新的轮询负载均衡器构建器。
// newBuilder creates a new roundrobin balancer builder.
func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &rrPickerBuilder{}, base.Config{HealthCheck: true})
}

// 在 init 函数中，将轮询负载均衡器注册为 gRPC 的一个默认负载均衡器。
func init() {
	balancer.Register(newBuilder())
}

type rrPickerBuilder struct{}

// rrPickerBuilder 结构体实现了构建器接口的 Build 方法。
// Build 方法根据准备好的子连接（ReadySCs）生成一个选择器。
// 如果没有可用的子连接，则返回一个错误选择器。
// 选择器从一个随机索引开始，以避免总是将负载集中在第一个服务器上。
func (*rrPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	logger.Infof("roundrobinPicker: Build called with info: %v", info)
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	scs := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for sc := range info.ReadySCs {
		scs = append(scs, sc)
	}
	return &rrPicker{
		subConns: scs,
		// Start at a random index, as the same RR balancer rebuilds a new
		// picker when SubConn states change, and we don't want to apply excess
		// load to the first server in the list.
		next: uint32(rand.Intn(len(scs))),
	}
}

// rrPicker 结构体包含了一个不可变的子连接列表和一个原子计数器 next。
type rrPicker struct {
	// subConns is the snapshot of the roundrobin balancer when this picker was
	// created. The slice is immutable. Each Get() will do a round robin
	// selection from it and return the selected SubConn.
	subConns []balancer.SubConn
	next     uint32
}

// Pick 方法根据当前的 next 索引进行轮询选择子连接，并返回选择结果。
func (p *rrPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	// 获取当前子连接的数量并转换为 uint32 类型
	subConnsLen := uint32(len(p.subConns))
	// 使用原子操作增加 next 计数器的值并获取增加后的值
	nextIndex := atomic.AddUint32(&p.next, 1)

	// 使用模运算获取当前应该选择的子连接索引
	sc := p.subConns[nextIndex%subConnsLen]
	// 返回选择结果，包含选中的子连接
	return balancer.PickResult{SubConn: sc}, nil
}
