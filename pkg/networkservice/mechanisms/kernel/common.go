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

package kernel

import (
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux"
	linuxinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linuxnamespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vppinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

const (
	// MECHANISM string
	MECHANISM = kernel.MECHANISM

	fileScheme   = "file"
	vnetFilename = "/dev/vhost-net"
)

func vethPairTemplate(conf *configurator.Config, name, ifaceName, netnsFilename string) {
	conf.GetLinuxConfig().Interfaces = append(conf.GetLinuxConfig().Interfaces,
		&linuxinterfaces.Interface{
			Name:       name + "-veth",
			Type:       linuxinterfaces.Interface_VETH,
			Enabled:    true,
			HostIfName: linuxIfaceName(name),
			Link: &linuxinterfaces.Interface_Veth{
				Veth: &linuxinterfaces.VethLink{
					PeerIfName:           name,
					RxChecksumOffloading: linuxinterfaces.VethLink_CHKSM_OFFLOAD_DISABLED,
					TxChecksumOffloading: linuxinterfaces.VethLink_CHKSM_OFFLOAD_DISABLED,
				},
			},
		},
		&linuxinterfaces.Interface{
			Name:       name,
			Type:       linuxinterfaces.Interface_VETH,
			Enabled:    true,
			HostIfName: linuxIfaceName(ifaceName),
			Namespace: &linuxnamespace.NetNamespace{
				Type:      linuxnamespace.NetNamespace_FD,
				Reference: netnsFilename,
			},
			Link: &linuxinterfaces.Interface_Veth{
				Veth: &linuxinterfaces.VethLink{
					PeerIfName:           name + "-veth",
					RxChecksumOffloading: linuxinterfaces.VethLink_CHKSM_OFFLOAD_DISABLED,
					TxChecksumOffloading: linuxinterfaces.VethLink_CHKSM_OFFLOAD_DISABLED,
				},
			},
		})
	conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vppinterfaces.Interface{
		Name:    name,
		Type:    vppinterfaces.Interface_AF_PACKET,
		Enabled: true,
		Link: &vppinterfaces.Interface_Afpacket{
			Afpacket: &vppinterfaces.AfpacketLink{
				HostIfName: linuxIfaceName(name),
			},
		},
	})
}

func tapV2Template(conf *configurator.Config, name, ifaceName, netnsFilename string) {
	// We append an Interfaces.  Interfaces creates the vpp side of an interface.
	//   In this case, a Tapv2 interface that has one side in vpp, and the other
	//   as a Linux kernel interface
	conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().GetInterfaces(), &vppinterfaces.Interface{
		Name:    name,
		Type:    vppinterfaces.Interface_TAP,
		Enabled: true,
		Link: &vppinterfaces.Interface_Tap{
			Tap: &vppinterfaces.TapLink{
				Version: 2,
			},
		},
	})
	// We apply configuration to LinuxInterfaces
	// Important details:
	//    - LinuxInterfaces.HostIfName - must be no longer than 15 chars (linux limitation)
	conf.GetLinuxConfig().Interfaces = append(conf.GetLinuxConfig().Interfaces, &linux.Interface{
		Name:       name,
		Type:       linuxinterfaces.Interface_TAP_TO_VPP,
		Enabled:    true,
		HostIfName: linuxIfaceName(ifaceName),
		Namespace: &linuxnamespace.NetNamespace{
			Type:      linuxnamespace.NetNamespace_FD,
			Reference: netnsFilename,
		},
		Link: &linuxinterfaces.Interface_Tap{
			Tap: &linuxinterfaces.TapLink{
				VppTapIfName: name,
			},
		},
	})
}

func linuxIfaceName(ifaceName string) string {
	if len(ifaceName) <= kernel.LinuxIfMaxLength {
		return ifaceName
	}
	return ifaceName[:kernel.LinuxIfMaxLength]
}
