package dreamhost

import (
	"context"
	"fmt"
	"sync"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
)

// MockDreamhostClient implements a mock DreamHost API client for testing
type MockDreamhostClient struct {
	mu sync.RWMutex
	
	// Storage
	records []dreamhostapi.DNSRecord
	
	// Behavior controls
	addRecordError    error
	removeRecordError error
	listRecordsError  error
	
	// Call tracking
	addRecordCalls    []dreamhostapi.DNSRecordInput
	removeRecordCalls []dreamhostapi.DNSRecordInput
	listRecordsCalls  int
	
	// Rate limiting simulation
	rateLimit      bool
	rateLimitCount int
}

// NewMockDreamhostClient creates a new mock client
func NewMockDreamhostClient() *MockDreamhostClient {
	return &MockDreamhostClient{
		records: []dreamhostapi.DNSRecord{},
	}
}

// AddDNSRecord mocks adding a DNS record
func (m *MockDreamhostClient) AddDNSRecord(ctx context.Context, record dreamhostapi.DNSRecordInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Track the call
	m.addRecordCalls = append(m.addRecordCalls, record)
	
	// Simulate rate limiting
	if m.rateLimit {
		m.rateLimitCount++
		if m.rateLimitCount%3 == 1 {
			return fmt.Errorf("rate limit exceeded")
		}
	}
	
	// Return configured error if set
	if m.addRecordError != nil {
		return m.addRecordError
	}
	
	// Check if record already exists
	for _, r := range m.records {
		if r.Record == record.Record && r.Type == record.Type && r.Value == record.Value {
			return fmt.Errorf("record already exists")
		}
	}
	
	// Add the record
	newRecord := dreamhostapi.DNSRecord{
		Record:    record.Record,
		Type:      record.Type,
		Value:     record.Value,
		Zone:      extractZone(record.Record),
		AccountID: "test-account-123",
		Comment:   "",
		Editable:  dreamhostapi.Editable,
	}
	
	m.records = append(m.records, newRecord)
	return nil
}

// RemoveDNSRecord mocks removing a DNS record
func (m *MockDreamhostClient) RemoveDNSRecord(ctx context.Context, record dreamhostapi.DNSRecordInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Track the call
	m.removeRecordCalls = append(m.removeRecordCalls, record)
	
	// Simulate rate limiting
	if m.rateLimit {
		m.rateLimitCount++
		if m.rateLimitCount%3 == 1 {
			return fmt.Errorf("rate limit exceeded")
		}
	}
	
	// Return configured error if set
	if m.removeRecordError != nil {
		return m.removeRecordError
	}
	
	// Find and remove the record
	found := false
	newRecords := []dreamhostapi.DNSRecord{}
	for _, r := range m.records {
		if r.Record == record.Record && r.Type == record.Type && r.Value == record.Value {
			found = true
			continue
		}
		newRecords = append(newRecords, r)
	}
	
	if !found {
		return fmt.Errorf("record not found")
	}
	
	m.records = newRecords
	return nil
}

// ListDNSRecords mocks listing DNS records
func (m *MockDreamhostClient) ListDNSRecords(ctx context.Context) ([]dreamhostapi.DNSRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Track the call
	m.listRecordsCalls++
	
	// Simulate rate limiting
	if m.rateLimit {
		m.rateLimitCount++
		if m.rateLimitCount%5 == 1 {
			return nil, fmt.Errorf("rate limit exceeded")
		}
	}
	
	// Return configured error if set
	if m.listRecordsError != nil {
		return nil, m.listRecordsError
	}
	
	// Return a copy of records
	result := make([]dreamhostapi.DNSRecord, len(m.records))
	copy(result, m.records)
	return result, nil
}

// Test helper methods

// SetRecords sets the mock records
func (m *MockDreamhostClient) SetRecords(records []dreamhostapi.DNSRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = records
}

// GetRecords returns the current records
func (m *MockDreamhostClient) GetRecords() []dreamhostapi.DNSRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]dreamhostapi.DNSRecord, len(m.records))
	copy(result, m.records)
	return result
}

// SetAddRecordError configures the error to return for AddDNSRecord
func (m *MockDreamhostClient) SetAddRecordError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addRecordError = err
}

// SetRemoveRecordError configures the error to return for RemoveDNSRecord
func (m *MockDreamhostClient) SetRemoveRecordError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeRecordError = err
}

// SetListRecordsError configures the error to return for ListDNSRecords
func (m *MockDreamhostClient) SetListRecordsError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listRecordsError = err
}

// SetRateLimit enables/disables rate limit simulation
func (m *MockDreamhostClient) SetRateLimit(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimit = enabled
	m.rateLimitCount = 0
}

// GetAddRecordCalls returns the list of AddDNSRecord calls
func (m *MockDreamhostClient) GetAddRecordCalls() []dreamhostapi.DNSRecordInput {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]dreamhostapi.DNSRecordInput, len(m.addRecordCalls))
	copy(result, m.addRecordCalls)
	return result
}

// GetRemoveRecordCalls returns the list of RemoveDNSRecord calls
func (m *MockDreamhostClient) GetRemoveRecordCalls() []dreamhostapi.DNSRecordInput {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]dreamhostapi.DNSRecordInput, len(m.removeRecordCalls))
	copy(result, m.removeRecordCalls)
	return result
}

// GetListRecordsCalls returns the number of ListDNSRecords calls
func (m *MockDreamhostClient) GetListRecordsCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.listRecordsCalls
}

// Reset resets all call tracking
func (m *MockDreamhostClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addRecordCalls = []dreamhostapi.DNSRecordInput{}
	m.removeRecordCalls = []dreamhostapi.DNSRecordInput{}
	m.listRecordsCalls = 0
	m.rateLimitCount = 0
}

// Helper function to extract zone from record name
func extractZone(record string) string {
	// Simple implementation - in reality would need proper domain parsing
	parts := splitDomain(record)
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}
	return record
}

func splitDomain(domain string) []string {
	var parts []string
	current := ""
	for _, ch := range domain {
		if ch == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}