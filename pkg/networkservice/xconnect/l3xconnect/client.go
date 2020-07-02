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

// Package l3xconnect provides a NetworkServiceClient chain element for an l3 cross connect
package l3xconnect

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice/payloads/ip"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type l3XconnectClient struct{}

// newClient - creates a NetworkServiceClient chain element for an l3 cross connect
func NewClient() networkservice.NetworkServiceClient {
	return &l3XconnectClient{}
}

func (l *l3XconnectClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	rv, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	l.appendl3XConnect(ctx, rv)
	return rv, nil
}

func (l *l3XconnectClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	rv, err := next.Client(ctx).Close(ctx, conn, opts...)
	if err != nil {
		return nil, err
	}
	l.appendl3XConnect(ctx, conn)
	return rv, nil
}

func (l *l3XconnectClient) appendl3XConnect(ctx context.Context, conn *networkservice.Connection) {
	if !(conn.Payload == ip.PAYLOAD || conn.Payload == "") {
		return
	}
	conf := vppagent.Config(ctx)
	if len(conf.GetVppConfig().GetInterfaces()) >= 2 {
		ifaces := conf.GetVppConfig().GetInterfaces()[len(conf.GetVppConfig().Interfaces)-2:]

		// Create l3 cross connect
		conf.GetVppConfig().L3Xconnects = append(conf.GetVppConfig().L3Xconnects,
			&l3.L3XConnect{
				Interface: ifaces[0].Name,
				Paths: []*l3.L3XConnect_Path{
					{
						OutgoingInterface: ifaces[1].Name,
					},
				},
			},
			&l3.L3XConnect{
				Interface: ifaces[1].Name,
				Paths: []*l3.L3XConnect_Path{
					{
						OutgoingInterface: ifaces[0].Name,
					},
				},
			},
		)

		// Setup Proper Arp behavior for IPv4
		// TODO - handle IPv6 ND, also handle properly mismatch between IPv4 and IPv6 src/dst
		// TODO - should we simply proxy arp *everything* ?
		if conn.GetContext().GetIpContext().GetDstIpAddr() != "" || conn.GetContext().GetIpContext().GetSrcIpAddr() != "" {
			if conf.GetVppConfig().GetProxyArp() == nil {
				conf.GetVppConfig().ProxyArp = &l3.ProxyARP{}
			}
			conf.GetVppConfig().GetProxyArp().Interfaces = append(conf.GetVppConfig().GetProxyArp().GetInterfaces(),
				&l3.ProxyARP_Interface{
					Name: ifaces[0].Name,
				},
			)
			conf.GetVppConfig().GetProxyArp().Interfaces = append(conf.GetVppConfig().GetProxyArp().GetInterfaces(),
				&l3.ProxyARP_Interface{
					Name: ifaces[1].Name,
				},
			)
			l3xconns := conf.GetVppConfig().GetL3Xconnects()[len(conf.GetVppConfig().GetL3Xconnects())-2:]
			if conn.GetContext().GetIpContext().GetDstIpAddr() != "" {
				l3xconns[1].GetPaths()[0].NextHopAddr = conn.GetContext().GetIpContext().GetDstIpAddr()
				//conf.GetVppConfig().Routes = append(conf.GetVppConfig().Routes, &l3.Route{
				//	Type:              l3.Route_INTER_VRF,
				//	DstNetwork:        conn.GetContext().GetIpContext().GetDstIpAddr(),
				//	OutgoingInterface: ifaces[1].Name,
				//})
				conf.GetVppConfig().GetProxyArp().Ranges = append(conf.GetVppConfig().GetProxyArp().GetRanges(),
					&l3.ProxyARP_Range{
						// TODO - handle more than /32
						FirstIpAddr: conn.GetContext().GetIpContext().GetDstIpAddr(),
						LastIpAddr:  conn.GetContext().GetIpContext().GetDstIpAddr(),
					},
				)
			}
			if conn.GetContext().GetIpContext().GetDstIpAddr() != "" {
				l3xconns[0].GetPaths()[0].NextHopAddr = conn.GetContext().GetIpContext().GetSrcIpAddr()
				//conf.GetVppConfig().Routes = append(conf.GetVppConfig().Routes, &l3.Route{
				//	Type:              l3.Route_INTER_VRF,
				//	DstNetwork:        conn.GetContext().GetIpContext().GetSrcIpAddr(),
				//	OutgoingInterface: ifaces[0].Name,
				//})
				conf.GetVppConfig().GetProxyArp().Ranges = append(conf.GetVppConfig().GetProxyArp().GetRanges(),
					&l3.ProxyARP_Range{
						// TODO - handle more than /32
						FirstIpAddr: conn.GetContext().GetIpContext().GetSrcIpAddr(),
						LastIpAddr:  conn.GetContext().GetIpContext().GetSrcIpAddr(),
					},
				)
			}
		}
	}
}
