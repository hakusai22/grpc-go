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

// Package base defines a balancer base that can be used to build balancers with
// different picking algorithms.
//
// The base balancer creates a new SubConn for each resolved address. The
// provided picker will only be notified about READY SubConns.
//
// This package is the base of round_robin balancer, its purpose is to be used
// to build round_robin like balancers with complex picking algorithms.
// Balancers with more complicated logic should try to implement a balancer
// builder from scratch.
//
// All APIs in this package are experimental.
package base

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

/**
这个包定义了一个基础负载均衡器，提供了构建具有不同选择算法的负载均衡器的框架。
通过实现 PickerBuilder 接口，开发者可以定义自定义的选择算法，
并使用 NewBalancerBuilder 函数创建一个新的负载均衡器构建器。
这个构建器可以管理子连接的创建和状态，并根据选择算法选择适当的子连接来处理请求。
*/

// PickerBuilder creates balancer.Picker.
// PickerBuilder 接口定义了一个方法 Build，
// 它返回一个 balancer.Picker，该选择器将被 gRPC 用于选择一个子连接 (SubConn)。
type PickerBuilder interface {
	// Build returns a picker that will be used by gRPC to pick a SubConn.
	Build(info PickerBuildInfo) balancer.Picker
}

// PickerBuildInfo contains information needed by the picker builder to
// construct a picker.
// PickerBuildInfo 结构体包含了构建选择器所需的信息，
// 其中 ReadySCs 是一个映射，包含所有准备好的子连接及其对应的地址。
type PickerBuildInfo struct {
	// ReadySCs is a map from all ready SubConns to the Addresses used to
	// create them.
	ReadySCs map[balancer.SubConn]SubConnInfo
}

// SubConnInfo contains information about a SubConn created by the base
// balancer.
// SubConnInfo 结构体包含了一个子连接的信息，其中 Address 是用于创建此子连接的地址。
type SubConnInfo struct {
	Address resolver.Address // the address used to create this SubConn
}

// Config contains the config info about the base balancer builder.
// Config 结构体包含了基础负载均衡器构建器的配置信息，
// 其中 HealthCheck 表示是否为此特定负载均衡器启用健康检查。
type Config struct {
	// HealthCheck indicates whether health checking should be enabled for this specific balancer.
	HealthCheck bool
}

// NewBalancerBuilder returns a base balancer builder configured by the provided config.
// NewBalancerBuilder 函数返回一个配置好的基础负载均衡器构建器。
// 参数包括 name（构建器的名称）、pb（一个 PickerBuilder 实例）和 config（配置）。
// 函数返回一个 baseBuilder 实例，这是一个实现了 balancer.Builder 接口的结构体。
func NewBalancerBuilder(name string, pb PickerBuilder, config Config) balancer.Builder {
	return &baseBuilder{
		name:          name,
		pickerBuilder: pb,
		config:        config,
	}
}
