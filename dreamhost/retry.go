package dreamhost

import (
	"context"
	"fmt"
	"time"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/pkg/errors"
)

const (
	// Retry configuration
	retryTimeout  = 2 * time.Minute
	retryDelay    = 5 * time.Second
	retryMinDelay = 1 * time.Second
)

// retryOnError retries a function when it returns certain errors
func retryOnError(ctx context.Context, f func() error) error {
	return resource.RetryContext(ctx, retryTimeout, func() *resource.RetryError {
		err := f()
		if err == nil {
			return nil
		}

		// Check if error is retryable
		if isRetryableError(err) {
			return resource.RetryableError(err)
		}

		// Non-retryable error
		return resource.NonRetryableError(err)
	})
}

// isRetryableError checks if an error should be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// API rate limiting
	if contains(errMsg, "rate limit") || contains(errMsg, "too many requests") {
		return true
	}

	// Temporary network issues
	if contains(errMsg, "timeout") || contains(errMsg, "connection refused") {
		return true
	}

	// API temporary unavailability
	if contains(errMsg, "service unavailable") || contains(errMsg, "bad gateway") {
		return true
	}

	return false
}

// waitForDNSRecord waits for a DNS record to appear in the API
func waitForDNSRecord(ctx context.Context, client *cachedDreamhostClient, record dreamhostapi.DNSRecordInput) (*dreamhostapi.DNSRecord, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"available"},
		Refresh:    dnsRecordStateRefreshFunc(ctx, client, record),
		Timeout:    retryTimeout,
		Delay:      retryDelay,
		MinTimeout: retryMinDelay,
	}

	result, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error waiting for DNS record")
	}

	if result == nil {
		return nil, fmt.Errorf("DNS record not found after waiting")
	}

	dnsRecord, ok := result.(*dreamhostapi.DNSRecord)
	if !ok {
		return nil, fmt.Errorf("unexpected type from state refresh: %T", result)
	}

	return dnsRecord, nil
}

// waitForDNSRecordDeletion waits for a DNS record to be deleted
func waitForDNSRecordDeletion(ctx context.Context, client *cachedDreamhostClient, record dreamhostapi.DNSRecordInput) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"deleting"},
		Target:     []string{"deleted"},
		Refresh:    dnsRecordDeletionStateRefreshFunc(ctx, client, record),
		Timeout:    retryTimeout,
		Delay:      retryDelay,
		MinTimeout: retryMinDelay,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return errors.Wrap(err, "error waiting for DNS record deletion")
	}

	return nil
}

// dnsRecordStateRefreshFunc returns a function that checks if a DNS record exists
func dnsRecordStateRefreshFunc(ctx context.Context, client *cachedDreamhostClient, recordInput dreamhostapi.DNSRecordInput) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Invalidate cache to get fresh data
		client.cache.Invalidate()

		record, err := client.GetDNSRecord(ctx, recordInput, false)
		if err != nil {
			return nil, "", err
		}

		if record == nil {
			return nil, "pending", nil
		}

		return record, "available", nil
	}
}

// dnsRecordDeletionStateRefreshFunc returns a function that checks if a DNS record has been deleted
func dnsRecordDeletionStateRefreshFunc(ctx context.Context, client *cachedDreamhostClient, recordInput dreamhostapi.DNSRecordInput) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Invalidate cache to get fresh data
		client.cache.Invalidate()

		record, err := client.GetDNSRecord(ctx, recordInput, false)
		if err != nil {
			return nil, "", err
		}

		if record != nil {
			return record, "deleting", nil
		}

		return nil, "deleted", nil
	}
}

