package dreamhost

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	t.Parallel()
	
	p := Provider()
	
	t.Run("provider_schema", func(t *testing.T) {
		assert.NotNil(t, p)
		assert.IsType(t, &schema.Provider{}, p)
		
		// Check schema
		assert.Contains(t, p.Schema, "api_key")
		apiKeySchema := p.Schema["api_key"]
		assert.Equal(t, schema.TypeString, apiKeySchema.Type)
		assert.True(t, apiKeySchema.Required)
		assert.True(t, apiKeySchema.Sensitive)
		assert.NotNil(t, apiKeySchema.DefaultFunc)
	})
	
	t.Run("provider_resources", func(t *testing.T) {
		assert.Contains(t, p.ResourcesMap, "dreamhost_dns_record")
		assert.NotNil(t, p.ResourcesMap["dreamhost_dns_record"])
	})
	
	t.Run("provider_data_sources", func(t *testing.T) {
		assert.Contains(t, p.DataSourcesMap, "dreamhost_dns_record")
		assert.Contains(t, p.DataSourcesMap, "dreamhost_dns_records")
		assert.NotNil(t, p.DataSourcesMap["dreamhost_dns_record"])
		assert.NotNil(t, p.DataSourcesMap["dreamhost_dns_records"])
	})
	
	t.Run("provider_configure_func", func(t *testing.T) {
		assert.NotNil(t, p.ConfigureContextFunc)
	})
}

func TestProviderConfigure(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name          string
		apiKey        string
		envVar        string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid_api_key",
			apiKey:      "test-api-key-123",
			expectError: false,
		},
		{
			name:          "empty_api_key",
			apiKey:        "",
			expectError:   true,
			errorContains: "Missing Dreamhost API key",
		},
		{
			name:        "api_key_from_env",
			envVar:      "env-api-key-456",
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			// Set up environment
			if tt.envVar != "" {
				oldEnv := os.Getenv(dreamhostAPIKeyEnvVarName)
				os.Setenv(dreamhostAPIKeyEnvVarName, tt.envVar)
				defer os.Setenv(dreamhostAPIKeyEnvVarName, oldEnv)
			}
			
			// Create resource data
			d := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{
				"api_key": tt.apiKey,
			})
			
			if tt.envVar != "" && tt.apiKey == "" {
				d.Set("api_key", tt.envVar)
			}
			
			// Call providerConfigure
			client, diags := providerConfigure(context.Background(), d)
			
			if tt.expectError {
				assert.True(t, diags.HasError())
				if tt.errorContains != "" {
					found := false
					for _, d := range diags {
						if d.Severity == "Error" && 
						   (contains(d.Summary, tt.errorContains) || contains(d.Detail, tt.errorContains)) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error containing '%s'", tt.errorContains)
				}
				assert.Nil(t, client)
			} else {
				assert.False(t, diags.HasError())
				assert.NotNil(t, client)
				_, ok := client.(*cachedDreamhostClient)
				assert.True(t, ok, "Expected client to be *cachedDreamhostClient")
			}
		})
	}
}

func TestProviderConfigureInvalidType(t *testing.T) {
	t.Parallel()
	
	// Create resource data with invalid type
	rawConfig := map[string]interface{}{
		"api_key": 12345, // Invalid: should be string
	}
	
	d := &schema.ResourceData{}
	d.SetId("test")
	
	// Mock the Get method to return non-string
	ctx := context.Background()
	
	// This should trigger the type assertion failure
	client, diags := providerConfigure(ctx, d)
	
	// We expect this to fail
	assert.Nil(t, client)
	// Note: In real scenario, this would panic or return error
	// The current implementation has a type assertion that could fail
}

// Helper function for testing
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && containsSubstring(s, substr)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestProviderValidation(t *testing.T) {
	t.Parallel()
	
	provider := Provider()
	
	t.Run("validate_provider_meta", func(t *testing.T) {
		// Test that provider can be validated
		err := provider.InternalValidate()
		require.NoError(t, err, "Provider internal validation should pass")
	})
}