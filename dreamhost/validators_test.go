package dreamhost

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDNSRecordName(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{"valid_domain", "example.com", false},
		{"valid_subdomain", "sub.example.com", false},
		{"valid_deep_subdomain", "a.b.c.example.com", false},
		{"valid_single_char", "a.com", false},
		{"valid_numbers", "123.example.com", false},
		{"valid_hyphen", "sub-domain.example.com", false},
		{"empty_string", "", true},
		{"too_long", string(make([]byte, 256)), true},
		{"starts_with_hyphen", "-example.com", true},
		{"ends_with_hyphen", "example-.com", true},
		{"double_dot", "example..com", true},
		{"starts_with_dot", ".example.com", true},
		{"ends_with_dot", "example.com.", true}, // Note: FQDN with trailing dot might be valid in some contexts
		{"invalid_chars", "example@.com", true},
		{"spaces", "example .com", true},
		{"non_string", 123, true},
	}
	
	validator := ValidateDNSRecordName()
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			warnings, errors := validator(tt.input, "test")
			
			if tt.expectError {
				assert.NotEmpty(t, errors, "Expected error for input: %v", tt.input)
			} else {
				assert.Empty(t, errors, "Expected no error for input: %v", tt.input)
			}
			assert.Empty(t, warnings, "No warnings expected")
		})
	}
}

func TestValidateIPv4Address(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{"valid_ipv4", "192.0.2.1", false},
		{"valid_ipv4_zero", "0.0.0.0", false},
		{"valid_ipv4_max", "255.255.255.255", false},
		{"valid_ipv4_localhost", "127.0.0.1", false},
		{"invalid_ipv6", "2001:db8::1", true},
		{"invalid_too_many_octets", "192.0.2.1.5", true},
		{"invalid_too_few_octets", "192.0.2", true},
		{"invalid_negative", "192.0.-2.1", true},
		{"invalid_over_255", "192.0.256.1", true},
		{"invalid_letters", "192.0.2.a", true},
		{"empty_string", "", true},
		{"hostname", "example.com", true},
		{"non_string", 192, true},
	}
	
	validator := ValidateIPv4Address()
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			warnings, errors := validator(tt.input, "test")
			
			if tt.expectError {
				assert.NotEmpty(t, errors, "Expected error for input: %v", tt.input)
			} else {
				assert.Empty(t, errors, "Expected no error for input: %v", tt.input)
			}
			assert.Empty(t, warnings, "No warnings expected")
		})
	}
}

func TestValidateIPv6Address(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{"valid_ipv6_full", "2001:0db8:0000:0000:0000:0000:0000:0001", false},
		{"valid_ipv6_compressed", "2001:db8::1", false},
		{"valid_ipv6_localhost", "::1", false},
		{"valid_ipv6_all_zeros", "::", false},
		{"valid_ipv6_mixed", "::ffff:192.0.2.1", false},
		{"invalid_ipv4", "192.0.2.1", true},
		{"invalid_too_many_groups", "2001:db8:0:0:0:0:0:0:1", true},
		{"invalid_chars", "2001:db8::g", true},
		{"empty_string", "", true},
		{"hostname", "example.com", true},
		{"non_string", 2001, true},
	}
	
	validator := ValidateIPv6Address()
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			warnings, errors := validator(tt.input, "test")
			
			if tt.expectError {
				assert.NotEmpty(t, errors, "Expected error for input: %v", tt.input)
			} else {
				assert.Empty(t, errors, "Expected no error for input: %v", tt.input)
			}
			assert.Empty(t, warnings, "No warnings expected")
		})
	}
}

func TestValidateMXRecord(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		errorMsg    string
	}{
		{"valid_mx", "10 mail.example.com", false, ""},
		{"valid_mx_zero_priority", "0 mail.example.com", false, ""},
		{"valid_mx_max_priority", "65535 mail.example.com", false, ""},
		{"valid_mx_subdomain", "20 mail.sub.example.com", false, ""},
		{"invalid_no_priority", "mail.example.com", true, "format"},
		{"invalid_negative_priority", "-1 mail.example.com", true, "between 0 and 65535"},
		{"invalid_priority_too_high", "65536 mail.example.com", true, "between 0 and 65535"},
		{"invalid_priority_not_number", "abc mail.example.com", true, "must be a number"},
		{"invalid_hostname", "10 mail@example.com", true, "not valid"},
		{"invalid_too_many_parts", "10 20 mail.example.com", true, "format"},
		{"empty_string", "", true, "format"},
		{"non_string", 10, true, "string"},
	}
	
	validator := ValidateMXRecord()
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			warnings, errors := validator(tt.input, "test")
			
			if tt.expectError {
				assert.NotEmpty(t, errors, "Expected error for input: %v", tt.input)
				if tt.errorMsg != "" {
					found := false
					for _, err := range errors {
						if containsSubstring(err.Error(), tt.errorMsg) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error containing '%s' for input: %v", tt.errorMsg, tt.input)
				}
			} else {
				assert.Empty(t, errors, "Expected no error for input: %v", tt.input)
			}
			assert.Empty(t, warnings, "No warnings expected")
		})
	}
}

func TestValidateTXTRecord(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{"valid_txt", "v=spf1 include:_spf.example.com ~all", false},
		{"valid_empty", "", false},
		{"valid_max_length", string(make([]byte, 255)), false},
		{"invalid_too_long", string(make([]byte, 256)), true},
		{"non_string", 123, true},
	}
	
	validator := ValidateTXTRecord()
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			warnings, errors := validator(tt.input, "test")
			
			if tt.expectError {
				assert.NotEmpty(t, errors, "Expected error for input: %v", tt.input)
			} else {
				assert.Empty(t, errors, "Expected no error for input: %v", tt.input)
			}
			assert.Empty(t, warnings, "No warnings expected")
		})
	}
}

func TestValidateSRVRecord(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		errorMsg    string
	}{
		{"valid_srv", "10 60 5060 sipserver.example.com", false, ""},
		{"valid_srv_zeros", "0 0 0 target.example.com", false, ""},
		{"valid_srv_max", "65535 65535 65535 target.example.com", false, ""},
		{"invalid_missing_parts", "10 60 5060", true, "format"},
		{"invalid_too_many_parts", "10 60 5060 80 target.example.com", true, "format"},
		{"invalid_priority_negative", "-1 60 5060 target.example.com", true, "between 0 and 65535"},
		{"invalid_weight_negative", "10 -1 5060 target.example.com", true, "between 0 and 65535"},
		{"invalid_port_negative", "10 60 -1 target.example.com", true, "between 0 and 65535"},
		{"invalid_priority_too_high", "65536 60 5060 target.example.com", true, "between 0 and 65535"},
		{"invalid_priority_not_number", "abc 60 5060 target.example.com", true, "must be a number"},
		{"invalid_target", "10 60 5060 target@example.com", true, "not valid"},
		{"empty_string", "", true, "format"},
		{"non_string", 10, true, "string"},
	}
	
	validator := ValidateSRVRecord()
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			warnings, errors := validator(tt.input, "test")
			
			if tt.expectError {
				assert.NotEmpty(t, errors, "Expected error for input: %v", tt.input)
				if tt.errorMsg != "" {
					found := false
					for _, err := range errors {
						if containsSubstring(err.Error(), tt.errorMsg) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error containing '%s' for input: %v", tt.errorMsg, tt.input)
				}
			} else {
				assert.Empty(t, errors, "Expected no error for input: %v", tt.input)
			}
			assert.Empty(t, warnings, "No warnings expected")
		})
	}
}

func TestValidateDNSRecordValue(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name        string
		recordType  string
		value       interface{}
		expectError bool
	}{
		// A records
		{"valid_a_record", "A", "192.0.2.1", false},
		{"invalid_a_record_ipv6", "A", "2001:db8::1", true},
		
		// AAAA records
		{"valid_aaaa_record", "AAAA", "2001:db8::1", false},
		{"invalid_aaaa_record_ipv4", "AAAA", "192.0.2.1", true},
		
		// CNAME records
		{"valid_cname_record", "CNAME", "example.com", false},
		{"valid_cname_at", "CNAME", "@", false},
		{"invalid_cname_ip", "CNAME", "192.0.2.1", true},
		
		// NS records
		{"valid_ns_record", "NS", "ns1.example.com", false},
		{"valid_ns_at", "NS", "@", false},
		
		// PTR records
		{"valid_ptr_record", "PTR", "example.com", false},
		
		// MX records
		{"valid_mx_record", "MX", "10 mail.example.com", false},
		{"invalid_mx_record", "MX", "mail.example.com", true},
		
		// TXT records
		{"valid_txt_record", "TXT", "v=spf1 ~all", false},
		{"valid_txt_empty", "TXT", "", false},
		
		// SRV records
		{"valid_srv_record", "SRV", "10 60 5060 sip.example.com", false},
		{"invalid_srv_record", "SRV", "sip.example.com", true},
		
		// Unknown type (no validation)
		{"unknown_type", "UNKNOWN", "anything", false},
		
		// Non-string input
		{"non_string", "A", 192, true},
	}
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			validator := ValidateDNSRecordValue(tt.recordType)
			warnings, errors := validator(tt.value, "test")
			
			if tt.expectError {
				assert.NotEmpty(t, errors, "Expected error for type %s with value: %v", tt.recordType, tt.value)
			} else {
				assert.Empty(t, errors, "Expected no error for type %s with value: %v", tt.recordType, tt.value)
			}
			assert.Empty(t, warnings, "No warnings expected")
		})
	}
}

func TestIsValidHostname(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name     string
		hostname string
		valid    bool
	}{
		{"valid_simple", "example.com", true},
		{"valid_subdomain", "sub.example.com", true},
		{"valid_deep", "a.b.c.example.com", true},
		{"valid_numbers", "123.456.example.com", true},
		{"valid_hyphen", "sub-domain.example.com", true},
		{"valid_fqdn", "example.com.", true}, // Trailing dot is valid for FQDN
		{"valid_single", "localhost", true},
		{"invalid_empty", "", false},
		{"invalid_too_long", string(make([]byte, 256)), false},
		{"invalid_label_too_long", "a" + string(make([]byte, 64)) + ".com", false},
		{"invalid_starts_hyphen", "-example.com", false},
		{"invalid_ends_hyphen", "example-.com", false},
		{"invalid_double_dot", "example..com", false},
		{"invalid_starts_dot", ".example.com", false},
		{"invalid_special_chars", "exam@ple.com", false},
		{"invalid_spaces", "exam ple.com", false},
		{"invalid_underscore", "exam_ple.com", false}, // Underscores are typically not valid in hostnames
	}
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			result := isValidHostname(tt.hostname)
			assert.Equal(t, tt.valid, result, "isValidHostname(%q) = %v, want %v", tt.hostname, result, tt.valid)
		})
	}
}

func TestIsAlphaNum(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		char  byte
		valid bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'-', false},
		{'_', false},
		{'.', false},
		{'@', false},
		{' ', false},
		{'!', false},
	}
	
	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.char), func(t *testing.T) {
			t.Parallel()
			
			result := isAlphaNum(tt.char)
			assert.Equal(t, tt.valid, result, "isAlphaNum(%c) = %v, want %v", tt.char, result, tt.valid)
		})
	}
}

func BenchmarkValidateIPv4Address(b *testing.B) {
	validator := ValidateIPv4Address()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = validator("192.0.2.1", "test")
		}
	})
}

func BenchmarkValidateDNSRecordName(b *testing.B) {
	validator := ValidateDNSRecordName()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = validator("subdomain.example.com", "test")
		}
	})
}

func BenchmarkIsValidHostname(b *testing.B) {
	hostname := "subdomain.example.com"
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = isValidHostname(hostname)
		}
	})
}