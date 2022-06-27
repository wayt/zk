package zk

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"
)

// WithEventCallback returns a connection option that specifies an event
// callback.
// The callback must not block - doing so would delay the ZK go routines.
func WithRefreshDNSHostProvider() connOption {
	return func(c *Conn) {
		c.hostProvider = &RefreshDNSHostProvider{
			done:   c.shouldQuit,
			logger: c.logger,
		}
	}
}

// RefreshDNSHostProvider is a wrapper around DNSHostProvider
// that will re-resolve server addresses everytime the list of
// server IPs has been fully tried.
type RefreshDNSHostProvider struct {
	DNSHostProvider

	done   chan struct{}
	logger Logger
}

func (hp *RefreshDNSHostProvider) Init(servers []string) error {
	go hp.refreshLoop(servers)
	return hp.DNSHostProvider.Init(servers)
}

func (hp *RefreshDNSHostProvider) refreshLoop(servers []string) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-hp.done:
			return
		case <-t.C:
			before := hp.safeHash(hp.servers)
			if err := hp.DNSHostProvider.Init(servers); err != nil {
				hp.logger.Printf("refreshdnshostprovider: failed to refresh server addresses: %v", err)
				continue
			}
			after := hp.safeHash(hp.servers)
			if before != after {
				hp.logger.Printf("refreshdnshostprovider: addresses updated: %v", hp.servers)
			}

		}
	}
}

func (hp *RefreshDNSHostProvider) safeHash(a []string) string {
	cp := make([]string, len(a))
	copy(cp, a)
	return fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(cp, ","))))
}
