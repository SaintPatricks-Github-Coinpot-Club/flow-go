package utils

import (
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/network/p2p/internal/p2putils"
)

// PeerAddressInfo generates the libp2p peer.AddrInfo for the given Flow.Identity.
// A node in flow is defined by a flow.Identity while it is defined by a peer.AddrInfo in libp2p.
//
//	flow.Identity        ---> peer.AddrInfo
//	|-- Address          --->   |-- []multiaddr.Multiaddr
//	|-- NetworkPublicKey --->   |-- ID
func PeerAddressInfo(identity flow.Identity) (peer.AddrInfo, error) {
	ip, port, key, err := p2putils.NetworkingInfo(identity)
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("could not translate identity to networking info %s: %w", identity.NodeID.String(), err)
	}

	addr := MultiAddressStr(ip, port)
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	id, err := peer.IDFromPublicKey(key)
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("could not extract libp2p id from key:%w", err)
	}
	pInfo := peer.AddrInfo{ID: id, Addrs: []multiaddr.Multiaddr{maddr}}
	return pInfo, err
}

// PeerInfosFromIDs converts the given flow.Identities to peer.AddrInfo.
// For each identity, if the conversion succeeds, the peer.AddrInfo is included in the result else it is
// included in the error map with the corresponding error
func PeerInfosFromIDs(ids flow.IdentityList) ([]peer.AddrInfo, map[flow.Identifier]error) {
	validIDs := make([]peer.AddrInfo, 0, len(ids))
	invalidIDs := make(map[flow.Identifier]error)
	for _, id := range ids {
		peerInfo, err := PeerAddressInfo(*id)
		if err != nil {
			invalidIDs[id.NodeID] = err
			continue
		}
		validIDs = append(validIDs, peerInfo)
	}
	return validIDs, invalidIDs
}

// MultiAddressStr receives a node ip and port and returns
// its corresponding Libp2p MultiAddressStr in string format
// in current implementation IP part of the node address is
// either an IP or a dns4.
// https://docs.libp2p.io/concepts/addressing/
func MultiAddressStr(ip, port string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP != nil {
		// returns parsed ip version of the multi-address
		return fmt.Sprintf("/ip4/%s/tcp/%s", ip, port)
	}
	// could not parse it as an IP address and returns the dns version of the
	// multi-address
	return fmt.Sprintf("/dns4/%s/tcp/%s", ip, port)
}
