package casdist

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/andyleap/casfs/cas"
	"github.com/serialx/hashring"
)

type CASDist struct {
	nodes        []string
	hr           *hashring.HashRing
	casClients   map[string]*cas.Client
	casClientsMu sync.RWMutex
	rep          int
}

func New(nodes []string, rep int) *CASDist {

	return &CASDist{
		nodes:      nodes,
		hr:         hashring.New(nodes),
		casClients: map[string]*cas.Client{},
		rep:        rep,
	}
}

func (cd *CASDist) Write(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	hexHash := hex.EncodeToString(hash[:])
	nodes, _ := cd.hr.GetNodes(hexHash, cd.rep)
	for _, node := range nodes {
		cd.casClientsMu.RLock()
		casClient := cd.casClients[node]
		cd.casClientsMu.RUnlock()
		for {
			if casClient == nil {
				var err error
				casClient, err = cas.Dial(node)
				if err != nil {
					break
				}
				cd.casClientsMu.Lock()
				cd.casClients[node] = casClient
				cd.casClientsMu.Unlock()
			}
			_, err := casClient.Write(data)
			if err != nil {
				casClient = nil
				continue
			}
			break
		}
	}
	return hash[:], nil
}

func (cd *CASDist) Read(hash []byte) ([]byte, error) {
	nodes, _ := cd.hr.GetNodes(hex.EncodeToString(hash[:]), len(cd.nodes))
	for _, node := range nodes {
		cd.casClientsMu.RLock()
		casClient := cd.casClients[node]
		cd.casClientsMu.RUnlock()
		for {
			if casClient == nil {
				casClient, _ = cas.Dial(node)
				cd.casClientsMu.Lock()
				cd.casClients[node] = casClient
				cd.casClientsMu.Unlock()
			}
			if casClient == nil {
				break
			}
			data, err := casClient.Read(hash)
			if err == nil {
				return data, nil
			}
			if err == cas.ErrLostConn {
				casClient = nil
				continue
			}
			break
		}
	}
	return nil, fmt.Errorf("Could not find data")
}

func (cd *CASDist) Mark() {
	for _, node := range cd.nodes {
		cd.casClientsMu.RLock()
		casClient := cd.casClients[node]
		cd.casClientsMu.RUnlock()
		for {
			if casClient == nil {
				casClient, _ = cas.Dial(node)
				cd.casClientsMu.Lock()
				cd.casClients[node] = casClient
				cd.casClientsMu.Unlock()
			}
			if casClient == nil {
				break
			}
			err := casClient.Mark()
			if err == nil {
				break
			}
			if err == cas.ErrLostConn {
				casClient = nil
				continue
			}
			break
		}
	}
}

func (cd *CASDist) Sweep() {
	for _, node := range cd.nodes {
		cd.casClientsMu.RLock()
		casClient := cd.casClients[node]
		cd.casClientsMu.RUnlock()
		for {
			if casClient == nil {
				casClient, _ = cas.Dial(node)
				cd.casClientsMu.Lock()
				cd.casClients[node] = casClient
				cd.casClientsMu.Unlock()
			}
			if casClient == nil {
				break
			}
			err := casClient.Sweep()
			if err == nil {
				break
			}
			if err == cas.ErrLostConn {
				casClient = nil
				continue
			}
			break
		}
	}
}
