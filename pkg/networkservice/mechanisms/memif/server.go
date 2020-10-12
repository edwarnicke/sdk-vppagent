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

package memif

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/golang/protobuf/ptypes/empty"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vppinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/memif"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type memifServer struct {
	baseDir string
}

// NewServer provides a NetworkServiceServer chain elements that support the memif Mechanism
func NewServer(baseDir string) networkservice.NetworkServiceServer {
	return &memifServer{baseDir: baseDir}
}

func (m *memifServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	m.appendInterfaceConfig(ctx, request.GetConnection())
	return next.Server(ctx).Request(ctx, request)
}

func (m *memifServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	m.appendInterfaceConfig(ctx, conn)
	return next.Server(ctx).Close(ctx, conn)
}

func (m *memifServer) appendInterfaceConfig(ctx context.Context, conn *networkservice.Connection) {
	if mechanism := memif.ToMechanism(conn.GetMechanism()); mechanism != nil {
		conf := vppagent.Config(ctx)
		socketFile := filepath.Join(m.baseDir, fmt.Sprintf("%s.memif.socket", conn.GetId()))
		mechanism.SetSocketFileURL((&url.URL{Scheme: "file", Path: socketFile}).String())
		conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vpp.Interface{
			Name:    fmt.Sprintf("server-%s", conn.GetId()),
			Type:    vppinterfaces.Interface_MEMIF,
			Enabled: true,
			Link: &vppinterfaces.Interface_Memif{
				Memif: &vppinterfaces.MemifLink{
					Master:         true,
					SocketFilename: socketFile,
				},
			},
		})
	}
}
