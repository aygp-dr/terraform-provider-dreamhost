package dreamhost

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_GetRecords(t *testing.T) {
	t.Parallel()
	
	t.Run("first_call_fetches_from_api", func(t *testing.T) {
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
		
		cache := &cache{}
		client := &cachedDreamhostClient{
			client: mockClient,
			cache:  *cache,
		}
		
		ctx := context.Background()
		records, err := cache.GetRecords(ctx, client)
		
		require.NoError(t, err)
		assert.Len(t, records, 2)
		assert.Equal(t, testRecords, records)
		assert.Equal(t, 1, mockClient.GetListRecordsCalls())
	})
	
	t.Run("second_call_uses_cache", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		testRecords := []dreamhostapi.DNSRecord{
			{
				Record: "example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
		}
		mockClient.SetRecords(testRecords)
		
		cache := &cache{}
		client := &cachedDreamhostClient{
			client: mockClient,
			cache:  *cache,
		}
		
		ctx := context.Background()
		
		// First call
		records1, err1 := cache.GetRecords(ctx, client)
		require.NoError(t, err1)
		assert.Len(t, records1, 1)
		
		// Second call should use cache
		records2, err2 := cache.GetRecords(ctx, client)
		require.NoError(t, err2)
		assert.Len(t, records2, 1)
		
		// API should only be called once
		assert.Equal(t, 1, mockClient.GetListRecordsCalls())
		
		// Records should be the same
		assert.Equal(t, records1, records2)
	})
	
	t.Run("error_handling", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		mockClient.SetListRecordsError(fmt.Errorf("API error"))
		
		cache := &cache{}
		client := &cachedDreamhostClient{
			client: mockClient,
			cache:  *cache,
		}
		
		ctx := context.Background()
		records, err := cache.GetRecords(ctx, client)
		
		assert.Error(t, err)
		assert.Nil(t, records)
		assert.Contains(t, err.Error(), "failed to list DNS records")
		assert.Contains(t, err.Error(), "API error")
	})
}

func TestCache_Invalidate(t *testing.T) {
	t.Parallel()
	
	t.Run("invalidate_clears_cache", func(t *testing.T) {
		t.Parallel()
		
		mockClient := NewMockDreamhostClient()
		testRecords := []dreamhostapi.DNSRecord{
			{
				Record: "example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
		}
		mockClient.SetRecords(testRecords)
		
		cache := &cache{}
		client := &cachedDreamhostClient{
			client: mockClient,
			cache:  *cache,
		}
		
		ctx := context.Background()
		
		// First call populates cache
		records1, err1 := cache.GetRecords(ctx, client)
		require.NoError(t, err1)
		assert.Len(t, records1, 1)
		assert.Equal(t, 1, mockClient.GetListRecordsCalls())
		
		// Invalidate cache
		cache.Invalidate()
		
		// Add a new record to the mock
		newRecords := []dreamhostapi.DNSRecord{
			{
				Record: "example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
			{
				Record: "new.example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.2",
			},
		}
		mockClient.SetRecords(newRecords)
		
		// Next call should fetch from API again
		records2, err2 := cache.GetRecords(ctx, client)
		require.NoError(t, err2)
		assert.Len(t, records2, 2)
		assert.Equal(t, 2, mockClient.GetListRecordsCalls())
	})
}

func TestCache_ThreadSafety(t *testing.T) {
	t.Run("concurrent_get_records", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		testRecords := []dreamhostapi.DNSRecord{
			{
				Record: "example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
		}
		mockClient.SetRecords(testRecords)
		
		cache := &cache{}
		client := &cachedDreamhostClient{
			client: mockClient,
			cache:  *cache,
		}
		
		ctx := context.Background()
		
		// Run multiple goroutines accessing cache
		var wg sync.WaitGroup
		errors := make(chan error, 100)
		
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				records, err := cache.GetRecords(ctx, client)
				if err != nil {
					errors <- err
					return
				}
				if len(records) != 1 {
					errors <- fmt.Errorf("expected 1 record, got %d", len(records))
				}
			}()
		}
		
		wg.Wait()
		close(errors)
		
		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent access error: %v", err)
		}
		
		// API should be called only once due to caching
		assert.Equal(t, 1, mockClient.GetListRecordsCalls())
	})
	
	t.Run("concurrent_invalidate_and_get", func(t *testing.T) {
		mockClient := NewMockDreamhostClient()
		testRecords := []dreamhostapi.DNSRecord{
			{
				Record: "example.com",
				Type:   dreamhostapi.ARecordType,
				Value:  "192.0.2.1",
			},
		}
		mockClient.SetRecords(testRecords)
		
		cache := &cache{}
		client := &cachedDreamhostClient{
			client: mockClient,
			cache:  *cache,
		}
		
		ctx := context.Background()
		
		// Run concurrent operations
		var wg sync.WaitGroup
		
		// Readers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = cache.GetRecords(ctx, client)
			}()
		}
		
		// Invalidators
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Millisecond * 10)
				cache.Invalidate()
			}()
		}
		
		wg.Wait()
		
		// No assertions on call count as it's non-deterministic
		// This test ensures no race conditions or panics
	})
}

func TestCache_ReturnsCopy(t *testing.T) {
	t.Parallel()
	
	mockClient := NewMockDreamhostClient()
	testRecords := []dreamhostapi.DNSRecord{
		{
			Record: "example.com",
			Type:   dreamhostapi.ARecordType,
			Value:  "192.0.2.1",
		},
	}
	mockClient.SetRecords(testRecords)
	
	cache := &cache{}
	client := &cachedDreamhostClient{
		client: mockClient,
		cache:  *cache,
	}
	
	ctx := context.Background()
	
	// Get records
	records, err := cache.GetRecords(ctx, client)
	require.NoError(t, err)
	
	// Modify returned slice
	records[0].Value = "modified"
	
	// Get records again
	records2, err := cache.GetRecords(ctx, client)
	require.NoError(t, err)
	
	// Original cache should not be modified
	assert.Equal(t, "192.0.2.1", records2[0].Value)
	assert.NotEqual(t, "modified", records2[0].Value)
}

func BenchmarkCache_GetRecords(b *testing.B) {
	mockClient := NewMockDreamhostClient()
	testRecords := make([]dreamhostapi.DNSRecord, 1000)
	for i := 0; i < 1000; i++ {
		testRecords[i] = dreamhostapi.DNSRecord{
			Record: fmt.Sprintf("subdomain%d.example.com", i),
			Type:   dreamhostapi.ARecordType,
			Value:  fmt.Sprintf("192.0.2.%d", i%256),
		}
	}
	mockClient.SetRecords(testRecords)
	
	cache := &cache{}
	client := &cachedDreamhostClient{
		client: mockClient,
		cache:  *cache,
	}
	
	ctx := context.Background()
	
	// Warm up cache
	_, _ = cache.GetRecords(ctx, client)
	
	b.ResetTimer()
	
	// Benchmark cached access
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.GetRecords(ctx, client)
		}
	})
}