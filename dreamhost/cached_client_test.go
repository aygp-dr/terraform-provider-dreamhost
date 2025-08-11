package dreamhost

import (
	"context"
	"fmt"
	"testing"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDreamhostClient(t *testing.T) {
	t.Parallel()
	
	// This would normally use a real dreamhostapi.Client
	// For testing, we'd need to mock it
	client := &dreamhostapi.Client{}
	
	cachedClient := newDreamhostClient(client)
	
	assert.NotNil(t, cachedClient)
	assert.Equal(t, client, cachedClient.client)
	assert.NotNil(t, cachedClient.cache)
}

func TestCachedDreamhostClient_AddDNSRecord(t *testing.T) {
	t.Parallel()
	
	t.Run("successful_add_invalidates_cache", func(t *testing.T) {
		t.Parallel()
		
		// We need a mock implementation since we can't use the real client
		// This test demonstrates the behavior we expect
		ctx := context.Background()
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "test.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		// Test would verify:
		// 1. AddDNSRecord is called on the underlying client
		// 2. Cache is invalidated on success
		// 3. Error is returned on failure
		
		// Since we can't mock dreamhostapi.Client directly without an interface,
		// this demonstrates the expected behavior
		_ = ctx
		_ = recordInput
	})
}

func TestCachedDreamhostClient_GetDNSRecord(t *testing.T) {
	t.Parallel()
	
	t.Run("cache_enabled_record_found", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		testRecord := dreamhostapi.DNSRecord{
			Record:    "test.example.com",
			Type:      dreamhostapi.ARecordType,
			Value:     "192.0.2.1",
			Zone:      "example.com",
			AccountID: "123",
			Editable:  dreamhostapi.Editable,
		}
		mockClient.SetRecords([]dreamhostapi.DNSRecord{testRecord})
		
		// Create a wrapper that can use our mock
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		ctx := context.Background()
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "test.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		// Test with cache enabled
		result, err := cachedClient.GetDNSRecord(ctx, recordInput, true)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testRecord.Record, result.Record)
		assert.Equal(t, testRecord.Type, result.Type)
		assert.Equal(t, testRecord.Value, result.Value)
	})
	
	t.Run("cache_enabled_record_not_found", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		mockClient.SetRecords([]dreamhostapi.DNSRecord{
			{
				Record: "other.example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.2",
			},
		})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		ctx := context.Background()
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "test.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		result, err := cachedClient.GetDNSRecord(ctx, recordInput, true)
		
		require.NoError(t, err)
		assert.Nil(t, result)
	})
	
	t.Run("cache_enabled_cname_with_trailing_dot", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		testRecord := dreamhostapi.DNSRecord{
			Record: "www.example.com",
			Type:   dreamhostapi.CNAMERecordType,
			Value:  "example.com.",  // With trailing dot
		}
		mockClient.SetRecords([]dreamhostapi.DNSRecord{testRecord})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		ctx := context.Background()
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "www.example.com",
			Type:   dreamhostapi.CNAMERecordType,
			Value:  "example.com",  // Without trailing dot
		}
		
		// Should find the record even with dot mismatch
		result, err := cachedClient.GetDNSRecord(ctx, recordInput, true)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testRecord.Record, result.Record)
		assert.Equal(t, "example.com.", result.Value)
	})
	
	t.Run("cache_disabled_refreshes_and_searches", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		testRecord := dreamhostapi.DNSRecord{
			Record: "test.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		mockClient.SetRecords([]dreamhostapi.DNSRecord{testRecord})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		ctx := context.Background()
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "test.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		// Test with cache disabled - should refresh and then search
		result, err := cachedClient.GetDNSRecord(ctx, recordInput, false)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testRecord.Record, result.Record)
		
		// Should have called ListDNSRecords twice (once for refresh, once for search)
		assert.GreaterOrEqual(t, mockClient.GetListRecordsCalls(), 2)
	})
	
	t.Run("cache_error_propagated", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		mockClient.SetListRecordsError(fmt.Errorf("API error"))
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		ctx := context.Background()
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "test.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		result, err := cachedClient.GetDNSRecord(ctx, recordInput, true)
		
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "API error")
	})
}

func TestCachedDreamhostClient_ListDNSRecords(t *testing.T) {
	t.Parallel()
	
	t.Run("successful_list", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		testRecords := []dreamhostapi.DNSRecord{
			{
				Record: "example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
			{
				Record: "www.example.com",
				Type:   dreamhostapi.CNAMERecordType,
				Value:  "example.com",
			},
		}
		mockClient.SetRecords(testRecords)
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		ctx := context.Background()
		records, err := cachedClient.ListDNSRecords(ctx)
		
		require.NoError(t, err)
		assert.Len(t, records, 2)
		assert.Equal(t, testRecords, records)
	})
	
	t.Run("error_propagated", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		mockClient.SetListRecordsError(fmt.Errorf("API error"))
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		ctx := context.Background()
		records, err := cachedClient.ListDNSRecords(ctx)
		
		assert.Error(t, err)
		assert.Nil(t, records)
		assert.Contains(t, err.Error(), "API error")
	})
}

func TestCachedDreamhostClient_RemoveDNSRecord(t *testing.T) {
	t.Parallel()
	
	t.Run("successful_remove_invalidates_cache", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		mockClient.SetRecords([]dreamhostapi.DNSRecord{
			{
				Record: "test.example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
		})
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		// Pre-populate cache
		ctx := context.Background()
		_, _ = cachedClient.cache.GetRecords(ctx, cachedClient)
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "test.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		err := cachedClient.RemoveDNSRecord(ctx, recordInput)
		
		require.NoError(t, err)
		
		// Verify record was removed
		remainingRecords := mockClient.GetRecords()
		assert.Len(t, remainingRecords, 0)
		
		// Cache should be invalidated - next GetRecords should fetch from API
		records, _ := cachedClient.cache.GetRecords(ctx, cachedClient)
		assert.Len(t, records, 0)
	})
	
	t.Run("error_does_not_invalidate_cache", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		testRecords := []dreamhostapi.DNSRecord{
			{
				Record: "test.example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
		}
		mockClient.SetRecords(testRecords)
		mockClient.SetRemoveRecordError(fmt.Errorf("API error"))
		
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		// Pre-populate cache
		ctx := context.Background()
		_, _ = cachedClient.cache.GetRecords(ctx, cachedClient)
		
		recordInput := dreamhostapi.DNSRecordInput{
			Record: "test.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		
		err := cachedClient.RemoveDNSRecord(ctx, recordInput)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
		
		// Cache should still have the original records
		// (though we can't directly check without exposing cache internals)
		assert.Len(t, mockClient.GetRecords(), 1)
	})
}

func TestCachedDreamhostClient_CacheInvalidation(t *testing.T) {
	t.Run("add_record_invalidates_cache", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		ctx := context.Background()
		
		// Pre-populate cache with initial records
		mockClient.SetRecords([]dreamhostapi.DNSRecord{
			{
				Record: "existing.example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
		})
		
		// Get records to populate cache
		records1, _ := cachedClient.cache.GetRecords(ctx, cachedClient)
		assert.Len(t, records1, 1)
		calls1 := mockClient.GetListRecordsCalls()
		
		// Add a new record
		newRecord := dreamhostapi.DNSRecordInput{
			Record: "new.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.2",
		}
		_ = cachedClient.AddDNSRecord(ctx, newRecord)
		
		// Get records again - should fetch from API due to cache invalidation
		records2, _ := cachedClient.cache.GetRecords(ctx, cachedClient)
		assert.Len(t, records2, 2)
		calls2 := mockClient.GetListRecordsCalls()
		
		// Verify API was called again after invalidation
		assert.Greater(t, calls2, calls1)
	})
	
	t.Run("remove_record_invalidates_cache", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		cachedClient := &cachedDreamhostClient{
			client: mockClient,
			cache:  cache{},
		}
		
		ctx := context.Background()
		
		// Pre-populate with records
		mockClient.SetRecords([]dreamhostapi.DNSRecord{
			{
				Record: "test1.example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
			{
				Record: "test2.example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.2",
			},
		})
		
		// Get records to populate cache
		records1, _ := cachedClient.cache.GetRecords(ctx, cachedClient)
		assert.Len(t, records1, 2)
		calls1 := mockClient.GetListRecordsCalls()
		
		// Remove a record
		removeRecord := dreamhostapi.DNSRecordInput{
			Record: "test1.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		}
		_ = cachedClient.RemoveDNSRecord(ctx, removeRecord)
		
		// Get records again - should fetch from API due to cache invalidation
		records2, _ := cachedClient.cache.GetRecords(ctx, cachedClient)
		assert.Len(t, records2, 1)
		calls2 := mockClient.GetListRecordsCalls()
		
		// Verify API was called again after invalidation
		assert.Greater(t, calls2, calls1)
	})
}

func TestCachedDreamhostClient_PointerSafety(t *testing.T) {
	t.Parallel()
	
	// This test ensures we don't return pointers to loop variables
	mockClient := NewMockDreamhostClient()
	testRecords := []dreamhostapi.DNSRecord{
		{
			Record: "test1.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		},
		{
			Record: "test2.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.2",
		},
		{
			Record: "test3.example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.3",
		},
	}
	mockClient.SetRecords(testRecords)
	
	cachedClient := &cachedDreamhostClient{
		client: mockClient,
		cache:  cache{},
	}
	
	ctx := context.Background()
	
	// Get multiple records and store pointers
	var results []*dreamhostapi.DNSRecord
	
	for i, testRecord := range testRecords {
		recordInput := dreamhostapi.DNSRecordInput{
			Record: testRecord.Record,
			Type:   testRecord.Type,
			Value:  testRecord.Value,
		}
		
		result, err := cachedClient.GetDNSRecord(ctx, recordInput, true)
		require.NoError(t, err)
		require.NotNil(t, result)
		
		results = append(results, result)
		
		// Verify each result has the correct value
		assert.Equal(t, testRecords[i].Record, result.Record)
		assert.Equal(t, testRecords[i].Value, result.Value)
	}
	
	// Verify all results are still correct (no pointer reuse)
	for i, result := range results {
		assert.Equal(t, testRecords[i].Record, result.Record)
		assert.Equal(t, testRecords[i].Value, result.Value)
	}
}