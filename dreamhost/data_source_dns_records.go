package dreamhost

import (
	"context"
	"strconv"
	"time"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDNSRecords() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSRecordsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Filter criteria for DNS records",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"record": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter by record name (supports partial match)",
						},
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter by record type (A, AAAA, CNAME, MX, NS, PTR, TXT, SRV, NAPTR)",
						},
						"value": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter by record value (supports partial match)",
						},
						"zone": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter by zone",
						},
					},
				},
			},
			"records": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of DNS records",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unique identifier for the record",
						},
						"record": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The DNS record name",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The DNS record type",
						},
						"value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The DNS record value",
						},
						"zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The DNS zone",
						},
						"comment": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Comment associated with the record",
						},
						"account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Account ID for the record",
						},
						"editable": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Whether the record is editable",
						},
					},
				},
			},
		},
	}
}

func dataSourceDNSRecordsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, ok := m.(*cachedDreamhostClient)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve dreamhost API client")
	}

	var diags diag.Diagnostics

	// Get all DNS records
	records, err := api.ListDNSRecords(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// Apply filters if provided
	if v, ok := d.GetOk("filter"); ok {
		filters := v.([]interface{})
		if len(filters) > 0 && filters[0] != nil {
			filterMap := filters[0].(map[string]interface{})
			records = filterDNSRecords(records, filterMap)
		}
	}

	// Convert records to list of maps
	recordList := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		recordMap := map[string]interface{}{
			"id":         recordInputToID(dreamhostapi.DNSRecordInput{Record: record.Record, Type: record.Type, Value: record.Value}),
			"record":     record.Record,
			"type":       string(record.Type),
			"value":      record.Value,
			"zone":       record.Zone,
			"comment":    record.Comment,
			"account_id": record.AccountID,
			"editable":   string(record.Editable),
		}
		recordList = append(recordList, recordMap)
	}

	if err := d.Set("records", recordList); err != nil {
		return diag.Errorf("failed to set records: %v", err)
	}

	// Use current timestamp as ID to ensure the data source is always refreshed
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func filterDNSRecords(records []dreamhostapi.DNSRecord, filters map[string]interface{}) []dreamhostapi.DNSRecord {
	var filtered []dreamhostapi.DNSRecord

	for _, record := range records {
		if matchesFilter(record, filters) {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

func matchesFilter(record dreamhostapi.DNSRecord, filters map[string]interface{}) bool {
	if recordFilter, ok := filters["record"].(string); ok && recordFilter != "" {
		if !contains(record.Record, recordFilter) {
			return false
		}
	}

	if typeFilter, ok := filters["type"].(string); ok && typeFilter != "" {
		if string(record.Type) != typeFilter {
			return false
		}
	}

	if valueFilter, ok := filters["value"].(string); ok && valueFilter != "" {
		if !contains(record.Value, valueFilter) {
			return false
		}
	}

	if zoneFilter, ok := filters["zone"].(string); ok && zoneFilter != "" {
		if record.Zone != zoneFilter {
			return false
		}
	}

	return true
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

