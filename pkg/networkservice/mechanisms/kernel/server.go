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

// +build !windows

// Package kernel provides networkservice chain elements that support the kernel Mechanism
package kernel

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
	"github.com/networkservicemesh/sdk-vppagent/pkg/tools/kernelctx"
)

type kernelMechanismServer struct {
	template func(conf *configurator.Config, name, ifaceName, netnsFilename string)
}

// NewServer provides NetworkServiceServer chain elements that support the kernel Mechanism
func NewServer() networkservice.NetworkServiceServer {
	rv := &kernelMechanismServer{
		template: vethPairTemplate,
	}
	if _, err := os.Stat(vnetFilename); err == nil {
		rv.template = tapV2Template
	}
	return rv
}

func (k *kernelMechanismServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	if mechanism := kernel.ToMechanism(request.GetConnection().GetMechanism()); mechanism != nil {
		err := k.appendInterfaceConfig(ctx, request.GetConnection())
		if err != nil {
			return nil, err
		}
		linuxIfaces := vppagent.Config(ctx).GetLinuxConfig().GetInterfaces()
		ctx = kernelctx.WithServerInterface(ctx, linuxIfaces[len(linuxIfaces)-1])
	}
	return next.Server(ctx).Request(ctx, request)
}

func (k *kernelMechanismServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil {
		err := k.appendInterfaceConfig(ctx, conn)
		if err != nil {
			return nil, err
		}
		linuxIfaces := vppagent.Config(ctx).GetLinuxConfig().GetInterfaces()
		ctx = kernelctx.WithServerInterface(ctx, linuxIfaces[len(linuxIfaces)-1])
	}
	return next.Server(ctx).Close(ctx, conn)
}

func (k *kernelMechanismServer) appendInterfaceConfig(ctx context.Context, conn *networkservice.Connection) error {
	netNSURLStr := kernel.ToMechanism(conn.GetMechanism()).GetNetNSURL()
	netNSURL, err := url.Parse(netNSURLStr)
	if err != nil {
		return err
	}
	if netNSURL.Scheme != fileScheme {
		return errors.Errorf("kernel.ToMechanism(conn.GetMechanism()).GetNetNSURL() must be of scheme %q: %q", fileScheme, netNSURL)
	}
	k.template(vppagent.Config(ctx), fmt.Sprintf("server-%s", conn.GetId()), kernel.ToMechanism(conn.GetMechanism()).GetInterfaceName(conn), netNSURL.Path)
	return nil
}
