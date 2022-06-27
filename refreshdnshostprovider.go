package zk

// RefreshDNSHostProvider is a wrapper around DNSHostProvider
// that will re-resolve server addresses everytime the list of
// server IPs has been fully tried.
type RefreshDNSHostProvider struct {
	DNSHostProvider

	serverAddresses []string
}

func NewRefreshDNSHostProvider() HostProvider {
	return &RefreshDNSHostProvider{}
}

func (hp *RefreshDNSHostProvider) Init(servers []string) error {
	hp.mu.Lock()
	hp.serverAddresses = servers
	hp.mu.Unlock()

	return hp.DNSHostProvider.Init(hp.serverAddresses)
}

func (hp *RefreshDNSHostProvider) Next() (server string, retryStart bool) {
	server, retryStart = hp.DNSHostProvider.Next()
	if retryStart {
		err := hp.DNSHostProvider.Init(hp.serverAddresses)
		if err != nil {
			DefaultLogger.Printf("failed to refresh serverAddresses: %v", err)
		}
	}
	return
}
