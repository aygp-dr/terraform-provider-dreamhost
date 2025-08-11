package dreamhost

import (
	"context"
	"fmt"
	"testing"
	"time"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryOnError(t *testing.T) {
	t.Parallel()
	
	t.Run("success_on_first_try", func(t *testing.T) {
		t.Parallel()
		
		calls := 0
		err := retryOnError(context.Background(), func() error {
			calls++
			return nil
		})
		
		assert.NoError(t, err)
		assert.Equal(t, 1, calls)
	})
	
	t.Run("retry_on_rate_limit", func(t *testing.T) {
		t.Parallel()
		
		calls := 0
		err := retryOnError(context.Background(), func() error {
			calls++
			if calls < 3 {
				return fmt.Errorf("rate limit exceeded")
			}
			return nil
		})
		
		assert.NoError(t, err)
		assert.Equal(t, 3, calls)
	})
	
	t.Run("retry_on_timeout", func(t *testing.T) {
		t.Parallel()
		
		calls := 0
		err := retryOnError(context.Background(), func() error {
			calls++
			if calls < 2 {
				return fmt.Errorf("request timeout")
			}
			return nil
		})
		
		assert.NoError(t, err)
		assert.Equal(t, 2, calls)
	})
	
	t.Run("non_retryable_error", func(t *testing.T) {
		t.Parallel()
		
		calls := 0
		expectedErr := fmt.Errorf("invalid credentials")
		err := retryOnError(context.Background(), func() error {
			calls++
			return expectedErr
		})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
		assert.Equal(t, 1, calls, "Should not retry for non-retryable errors")
	})
	
	t.Run("context_cancellation", func(t *testing.T) {
		t.Parallel()
		
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		
		// Cancel after a short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		
		err := retryOnError(ctx, func() error {
			calls++
			return fmt.Errorf("rate limit exceeded") // Always fail to trigger retry
		})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestIsRetryableError(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"nil_error", nil, false},
		{"rate_limit", fmt.Errorf("rate limit exceeded"), true},
		{"too_many_requests", fmt.Errorf("too many requests"), true},
		{"timeout", fmt.Errorf("request timeout"), true},
		{"connection_refused", fmt.Errorf("connection refused"), true},
		{"service_unavailable", fmt.Errorf("service unavailable"), true},
		{"bad_gateway", fmt.Errorf("bad gateway"), true},
		{"invalid_credentials", fmt.Errorf("invalid credentials"), false},
		{"not_found", fmt.Errorf("record not found"), false},
		{"permission_denied", fmt.Errorf("permission denied"), false},
	}
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.retryable, result, "isRetryableError(%v) = %v, want %v", tt.err, result, tt.retryable)
		})
	}
}

func TestWaitForDNSRecord(t *testing.T) {
	t.Run("record_appears_immediately", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		record := dreamhostapi.DNSRecord{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		mockClient.SetRecords([]dreamhostapi.DNSRecord{record})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		ctx := context.Background()
		result, err := waitForDNSRecord(ctx, cachedClient, recordInput)
		
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, record.Record, result.Record)
		assert.Equal(t, record.Type, result.Type)
		assert.Equal(t, record.Value, result.Value)
	})
	
	t.Run("record_appears_after_delay", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		// Simulate record appearing after a delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			mockClient.SetRecords([]dreamhostapi.DNSRecord{
				{
					Record: "example.com",
					Type:   dreamhostapi.ARecordType,
					Value:  "192.0.2.1",
				},
			})
		}()
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		result, err := waitForDNSRecord(ctx, cachedClient, recordInput)
		
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
	
	t.Run("timeout_waiting_for_record", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		// Use a very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		
		result, err := waitForDNSRecord(ctx, cachedClient, recordInput)
		
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "waiting for DNS record")
	})
}

func TestWaitForDNSRecordDeletion(t *testing.T) {
	t.Run("record_deleted_immediately", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		// No records - already deleted
		mockClient.SetRecords([]dreamhostapi.DNSRecord{})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		ctx := context.Background()
		err := waitForDNSRecordDeletion(ctx, cachedClient, recordInput)
		
		assert.NoError(t, err)
	})
	
	t.Run("record_deleted_after_delay", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		mockClient.SetRecords([]dreamhostapi.DNSRecord{
			{
				Record: "example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
		})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		// Simulate record deletion after a delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			mockClient.SetRecords([]dreamhostapi.DNSRecord{})
		}()
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		err := waitForDNSRecordDeletion(ctx, cachedClient, recordInput)
		
		assert.NoError(t, err)
	})
	
	t.Run("timeout_waiting_for_deletion", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		mockClient.SetRecords([]dreamhostapi.DNSRecord{
			{
				Record: "example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
		})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		// Use a very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		
		err := waitForDNSRecordDeletion(ctx, cachedClient, recordInput)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "waiting for DNS record deletion")
	})
}

func TestDNSRecordStateRefreshFunc(t *testing.T) {
	t.Run("record_exists", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		record := dreamhostapi.DNSRecord{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		mockClient.SetRecords([]dreamhostapi.DNSRecord{record})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		refreshFunc := dnsRecordStateRefreshFunc(context.Background(), cachedClient, recordInput)
		result, state, err := refreshFunc()
		
		require.NoError(t, err)
		assert.Equal(t, "available", state)
		assert.NotNil(t, result)
		
		dnsRecord, ok := result.(*dreamhostapi.DNSRecord)
		require.True(t, ok)
		assert.Equal(t, record.Record, dnsRecord.Record)
	})
	
	t.Run("record_not_exists", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		mockClient.SetRecords([]dreamhostapi.DNSRecord{})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		refreshFunc := dnsRecordStateRefreshFunc(context.Background(), cachedClient, recordInput)
		result, state, err := refreshFunc()
		
		require.NoError(t, err)
		assert.Equal(t, "pending", state)
		assert.Nil(t, result)
	})
	
	t.Run("api_error", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		mockClient.SetListRecordsError(fmt.Errorf("API error"))
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		refreshFunc := dnsRecordStateRefreshFunc(context.Background(), cachedClient, recordInput)
		result, state, err := refreshFunc()
		
		assert.Error(t, err)
		assert.Empty(t, state)
		assert.Nil(t, result)
	})
}

func TestDNSRecordDeletionStateRefreshFunc(t *testing.T) {
	t.Run("record_still_exists", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		record := dreamhostapi.DNSRecord{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		mockClient.SetRecords([]dreamhostapi.DNSRecord{record})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		refreshFunc := dnsRecordDeletionStateRefreshFunc(context.Background(), cachedClient, recordInput)
		result, state, err := refreshFunc()
		
		require.NoError(t, err)
		assert.Equal(t, "deleting", state)
		assert.NotNil(t, result)
	})
	
	t.Run("record_deleted", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		mockClient.SetRecords([]dreamhostapi.DNSRecord{})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		refreshFunc := dnsRecordDeletionStateRefreshFunc(context.Background(), cachedClient, recordInput)
		result, state, err := refreshFunc()
		
		require.NoError(t, err)
		assert.Equal(t, "deleted", state)
		assert.Nil(t, result)
	})
}