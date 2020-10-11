// Copyright (c) 2020 Cisco Systems, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ipaddress

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type setVppIPServer struct{}

// NewServer creates a NetworkServiceServer chain element to set the ip address on a vpp interface
// It sets the IP Address on the *vpp* side of an interface plugged into the
// Endpoint.
//                                         Endpoint
//                              +---------------------------+
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//          +-------------------+ ipaddress.NewServer()     |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              +---------------------------+
//
func NewServer() networkservice.NetworkServiceServer {
	return &setVppIPServer{}
}

func (s *setVppIPServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	conf := vppagent.Config(ctx)
	if index := len(conf.GetVppConfig().GetInterfaces()) - 1; index >= 0 {
		dstIP := request.GetConnection().GetContext().GetIpContext().GetDstIpAddr()
		if dstIP != "" {
			conf.GetVppConfig().GetInterfaces()[index].IpAddresses = []string{dstIP}
		}
	}
	return next.Server(ctx).Request(ctx, request)
}

func (s *setVppIPServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	conf := vppagent.Config(ctx)
	if index := len(conf.GetVppConfig().GetInterfaces()) - 1; index >= 0 {
		dstIP := conn.GetContext().GetIpContext().GetDstIpAddr()
		if dstIP != "" {
			conf.GetVppConfig().GetInterfaces()[index].IpAddresses = []string{dstIP}
		}
	}
	return next.Server(ctx).Close(ctx, conn)
}
