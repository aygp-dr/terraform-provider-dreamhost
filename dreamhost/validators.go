package dreamhost

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ValidateDNSRecordName validates a DNS record name
func ValidateDNSRecordName() schema.SchemaValidateFunc {
	return validation.All(
		validation.StringLenBetween(1, 255),
		validation.StringMatch(
			regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-\.]*[a-zA-Z0-9])?$`),
			"must be a valid DNS hostname",
		),
	)
}

// ValidateIPv4Address validates an IPv4 address
func ValidateIPv4Address() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		ip := net.ParseIP(v)
		if ip == nil || ip.To4() == nil {
			errors = append(errors, fmt.Errorf("%s is not a valid IPv4 address", v))
		}

		return warnings, errors
	}
}

// ValidateIPv6Address validates an IPv6 address
func ValidateIPv6Address() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		ip := net.ParseIP(v)
		if ip == nil || ip.To4() != nil {
			errors = append(errors, fmt.Errorf("%s is not a valid IPv6 address", v))
		}

		return warnings, errors
	}
}

// ValidateMXRecord validates an MX record value (priority hostname)
func ValidateMXRecord() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		parts := strings.Fields(v)
		if len(parts) != 2 {
			errors = append(errors, fmt.Errorf("MX record must be in format 'priority hostname', got: %s", v))
			return warnings, errors
		}

		// Validate priority is a number
		var priority int
		if _, err := fmt.Sscanf(parts[0], "%d", &priority); err != nil {
			errors = append(errors, fmt.Errorf("MX priority must be a number, got: %s", parts[0]))
			return warnings, errors
		}

		if priority < 0 || priority > 65535 {
			errors = append(errors, fmt.Errorf("MX priority must be between 0 and 65535, got: %d", priority))
		}

		// Validate hostname
		if !isValidHostname(parts[1]) {
			errors = append(errors, fmt.Errorf("MX hostname is not valid: %s", parts[1]))
		}

		return warnings, errors
	}
}

// ValidateTXTRecord validates a TXT record value
func ValidateTXTRecord() schema.SchemaValidateFunc {
	return validation.StringLenBetween(0, 255)
}

// ValidateSRVRecord validates an SRV record value
func ValidateSRVRecord() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		// SRV format: priority weight port target
		parts := strings.Fields(v)
		if len(parts) != 4 {
			errors = append(errors, fmt.Errorf("SRV record must be in format 'priority weight port target', got: %s", v))
			return warnings, errors
		}

		// Validate priority, weight, and port are numbers
		var priority, weight, port int
		if _, err := fmt.Sscanf(parts[0], "%d", &priority); err != nil {
			errors = append(errors, fmt.Errorf("SRV priority must be a number, got: %s", parts[0]))
		}
		if _, err := fmt.Sscanf(parts[1], "%d", &weight); err != nil {
			errors = append(errors, fmt.Errorf("SRV weight must be a number, got: %s", parts[1]))
		}
		if _, err := fmt.Sscanf(parts[2], "%d", &port); err != nil {
			errors = append(errors, fmt.Errorf("SRV port must be a number, got: %s", parts[2]))
		}

		// Validate ranges
		if priority < 0 || priority > 65535 {
			errors = append(errors, fmt.Errorf("SRV priority must be between 0 and 65535, got: %d", priority))
		}
		if weight < 0 || weight > 65535 {
			errors = append(errors, fmt.Errorf("SRV weight must be between 0 and 65535, got: %d", weight))
		}
		if port < 0 || port > 65535 {
			errors = append(errors, fmt.Errorf("SRV port must be between 0 and 65535, got: %d", port))
		}

		// Validate target hostname
		if !isValidHostname(parts[3]) {
			errors = append(errors, fmt.Errorf("SRV target hostname is not valid: %s", parts[3]))
		}

		return warnings, errors
	}
}

// ValidateDNSRecordValue validates the value based on the record type
func ValidateDNSRecordValue(recordType string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		value, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		switch recordType {
		case "A":
			return ValidateIPv4Address()(i, k)
		case "AAAA":
			return ValidateIPv6Address()(i, k)
		case "CNAME", "NS", "PTR":
			if !isValidHostname(value) && value != "@" {
				errors = append(errors, fmt.Errorf("%s record value must be a valid hostname or '@', got: %s", recordType, value))
			}
		case "MX":
			return ValidateMXRecord()(i, k)
		case "TXT":
			return ValidateTXTRecord()(i, k)
		case "SRV":
			return ValidateSRVRecord()(i, k)
		}

		return warnings, errors
	}
}

// isValidHostname checks if a string is a valid hostname
func isValidHostname(hostname string) bool {
	if len(hostname) > 255 {
		return false
	}

	// Allow trailing dot for FQDN
	hostname = strings.TrimSuffix(hostname, ".")

	if hostname == "" {
		return false
	}

	// Check each label
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}

		// Label must start with alphanumeric
		if !isAlphaNum(label[0]) {
			return false
		}

		// Label must end with alphanumeric
		if len(label) > 1 && !isAlphaNum(label[len(label)-1]) {
			return false
		}

		// Check all characters are valid
		for _, ch := range label {
			if !isAlphaNum(byte(ch)) && ch != '-' {
				return false
			}
		}
	}

	return true
}

func isAlphaNum(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

