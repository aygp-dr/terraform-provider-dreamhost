package dreamhost

import (
	"context"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/pkg/errors"
)

type cachedDreamhostClient struct {
	client *dreamhostapi.Client
	cache  cache
}

func newDreamhostClient(client *dreamhostapi.Client) *cachedDreamhostClient {
	return &cachedDreamhostClient{
		client: client,
	}
}

func (c *cachedDreamhostClient) AddDNSRecord(ctx context.Context, recordInput dreamhostapi.DNSRecordInput) error {
	err := c.client.AddDNSRecord(ctx, recordInput)
	if err == nil {
		c.cache.Invalidate()
	}
	return err
}

func (c *cachedDreamhostClient) GetDNSRecord(
	ctx context.Context, recordInput dreamhostapi.DNSRecordInput, enableCache bool,
) (*dreamhostapi.DNSRecord, error) {
	if enableCache {
		records, err := c.cache.GetRecords(ctx, c)
		if err != nil {
			return nil, err
		}
		for i := range records {
			if records[i].Record == recordInput.Record &&
				records[i].Type == recordInput.Type &&
				(records[i].Value == recordInput.Value || records[i].Value+"." == recordInput.Value) {
				// record found - return a copy to avoid issues
				recordCopy := records[i]
				return &recordCopy, nil
			}
		}
		return nil, nil
	}
	if _, err := c.ListDNSRecords(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to refresh cache")
	}
	return c.GetDNSRecord(ctx, recordInput, true)
}

func (c *cachedDreamhostClient) ListDNSRecords(ctx context.Context) ([]dreamhostapi.DNSRecord, error) {
	records, err := c.client.ListDNSRecords(ctx)
	if err != nil {
		return nil, err
	}

	return records, err
}

func (c *cachedDreamhostClient) RemoveDNSRecord(ctx context.Context, recordInput dreamhostapi.DNSRecordInput) error {
	err := c.client.RemoveDNSRecord(ctx, recordInput)
	if err == nil {
		c.cache.Invalidate()
	}
	return err
}
