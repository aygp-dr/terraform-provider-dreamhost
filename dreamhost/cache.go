package dreamhost

import (
	"context"
	"sync"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/pkg/errors"
)

// DNSRecordLister interface to avoid circular dependency
type DNSRecordLister interface {
	ListDNSRecords(ctx context.Context) ([]dreamhostapi.DNSRecord, error)
}

type cache struct {
	sync.Mutex

	cachedRecords []dreamhostapi.DNSRecord
}

func (c *cache) GetRecords(ctx context.Context, client DNSRecordLister) ([]dreamhostapi.DNSRecord, error) {
	c.Lock()
	defer c.Unlock()

	if c.cachedRecords == nil {
		records, err := client.ListDNSRecords(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list DNS records")
		}
		c.cachedRecords = records
	}

	// Return a copy to prevent external modification
	result := make([]dreamhostapi.DNSRecord, len(c.cachedRecords))
	copy(result, c.cachedRecords)
	return result, nil
}

func (c *cache) Invalidate() {
	c.Lock()
	defer c.Unlock()
	c.cachedRecords = nil
}
