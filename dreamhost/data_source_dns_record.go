package dreamhost

import (
	"context"
	"fmt"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDNSRecord() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSRecordRead,
		Schema: map[string]*schema.Schema{
			"record": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The DNS record name to look up",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The DNS record type (A, AAAA, CNAME, MX, NS, PTR, TXT, SRV, NAPTR)",
			},
			"value": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The DNS record value (optional for lookup, will be populated from the found record)",
			},
			// Computed fields
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Unique identifier for the record",
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
	}
}

func dataSourceDNSRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, ok := m.(*cachedDreamhostClient)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve dreamhost API client")
	}

	var diags diag.Diagnostics

	recordName := d.Get("record").(string)
	recordType := d.Get("type").(string)
	recordValue, hasValue := d.GetOk("value")

	// Get all DNS records
	records, err := api.ListDNSRecords(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// Find matching record
	var foundRecord *dreamhostapi.DNSRecord
	for _, record := range records {
		if record.Record == recordName && string(record.Type) == recordType {
			// If value is specified, must match exactly
			if hasValue && record.Value != recordValue.(string) {
				continue
			}
			// If we already found a record and no value filter, this is ambiguous
			if foundRecord != nil && !hasValue {
				return diag.Errorf("multiple DNS records found for %s (type: %s). Please specify 'value' to disambiguate", recordName, recordType)
			}
			foundRecord = &record
			if hasValue {
				break // Exact match found
			}
		}
	}

	if foundRecord == nil {
		if hasValue {
			return diag.Errorf("DNS record not found: %s (type: %s, value: %s)", recordName, recordType, recordValue)
		}
		return diag.Errorf("DNS record not found: %s (type: %s)", recordName, recordType)
	}

	// Set all fields
	id := recordInputToID(dreamhostapi.DNSRecordInput{
		Record: foundRecord.Record,
		Type:   foundRecord.Type,
		Value:  foundRecord.Value,
	})
	d.SetId(id)

	if err := d.Set("id", id); err != nil {
		return diag.Errorf("failed to set id: %v", err)
	}
	if err := d.Set("record", foundRecord.Record); err != nil {
		return diag.Errorf("failed to set record: %v", err)
	}
	if err := d.Set("type", string(foundRecord.Type)); err != nil {
		return diag.Errorf("failed to set type: %v", err)
	}
	if err := d.Set("value", foundRecord.Value); err != nil {
		return diag.Errorf("failed to set value: %v", err)
	}
	if err := d.Set("zone", foundRecord.Zone); err != nil {
		return diag.Errorf("failed to set zone: %v", err)
	}
	if err := d.Set("comment", foundRecord.Comment); err != nil {
		return diag.Errorf("failed to set comment: %v", err)
	}
	if err := d.Set("account_id", foundRecord.AccountID); err != nil {
		return diag.Errorf("failed to set account_id: %v", err)
	}
	if err := d.Set("editable", string(foundRecord.Editable)); err != nil {
		return diag.Errorf("failed to set editable: %v", err)
	}

	return diags
}

