# DreamHost Terraform Provider Architecture

## Table of Contents
- [Overview](#overview)
- [Component Architecture](#component-architecture)
- [Data Flow](#data-flow)
- [Sequence Diagrams](#sequence-diagrams)
- [Module Descriptions](#module-descriptions)
- [Design Patterns](#design-patterns)

## Overview

The DreamHost Terraform Provider is built using the Terraform Plugin SDK v2 and follows HashiCorp's best practices for provider development. It interfaces with the DreamHost API to manage DNS records.

## Component Architecture

```mermaid
graph TB
    subgraph "Terraform Core"
        TC[Terraform CLI]
        TS[Terraform State]
    end
    
    subgraph "DreamHost Provider"
        P[provider.go<br/>Provider Configuration]
        
        subgraph "Resources"
            R[resource_dns_record.go<br/>DNS Record Resource]
        end
        
        subgraph "Data Sources"
            DS1[data_source_dns_record.go<br/>Single Record Lookup]
            DS2[data_source_dns_records.go<br/>Multiple Records Query]
        end
        
        subgraph "Core Components"
            CC[cached_client.go<br/>API Client Wrapper]
            C[cache.go<br/>Cache Management]
            RT[retry.go<br/>Retry Logic]
            V[validators.go<br/>Input Validation]
        end
    end
    
    subgraph "External"
        API[DreamHost API]
        GD[go-dreamhost<br/>Client Library]
    end
    
    TC --> P
    P --> R
    P --> DS1
    P --> DS2
    R --> CC
    DS1 --> CC
    DS2 --> CC
    CC --> C
    CC --> RT
    R --> V
    CC --> GD
    GD --> API
    TS <--> R
    TS <--> DS1
    TS <--> DS2
```

## Data Flow

### Provider Initialization Flow

```mermaid
sequenceDiagram
    participant User
    participant Terraform
    participant Provider
    participant Config
    participant APIClient
    
    User->>Terraform: terraform init
    Terraform->>Provider: Initialize Provider
    Provider->>Config: Load Configuration
    Config->>Config: Check API Key (env/config)
    Config->>APIClient: Create Client
    APIClient->>Provider: Return Configured Client
    Provider->>Terraform: Provider Ready
```

### DNS Record Creation Flow

```mermaid
sequenceDiagram
    participant Terraform
    participant Resource
    participant Validator
    participant CachedClient
    participant RetryLogic
    participant Cache
    participant DreamHostAPI
    
    Terraform->>Resource: Create DNS Record
    Resource->>Validator: Validate Input
    Validator-->>Resource: Validation Result
    
    alt Invalid Input
        Resource-->>Terraform: Return Error
    else Valid Input
        Resource->>CachedClient: AddDNSRecord()
        CachedClient->>RetryLogic: Execute with Retry
        RetryLogic->>DreamHostAPI: POST /dns-add_record
        
        alt API Error
            DreamHostAPI-->>RetryLogic: Error Response
            RetryLogic->>RetryLogic: Check if Retryable
            alt Retryable
                RetryLogic->>RetryLogic: Wait & Retry
                RetryLogic->>DreamHostAPI: POST /dns-add_record
            else Not Retryable
                RetryLogic-->>Resource: Return Error
            end
        else Success
            DreamHostAPI-->>RetryLogic: Success Response
            RetryLogic-->>CachedClient: Success
            CachedClient->>Cache: Invalidate Cache
            CachedClient->>CachedClient: Wait for Record
            loop Check Record Exists
                CachedClient->>DreamHostAPI: GET /dns-list_records
                DreamHostAPI-->>CachedClient: Records List
            end
            CachedClient-->>Resource: Record Created
            Resource->>Terraform: Update State
        end
    end
```

### DNS Record Read Flow

```mermaid
sequenceDiagram
    participant Terraform
    participant Resource
    participant CachedClient
    participant Cache
    participant DreamHostAPI
    
    Terraform->>Resource: Read DNS Record
    Resource->>CachedClient: GetDNSRecord()
    CachedClient->>Cache: Check Cache
    
    alt Cache Hit
        Cache-->>CachedClient: Return Cached Records
    else Cache Miss
        CachedClient->>DreamHostAPI: GET /dns-list_records
        DreamHostAPI-->>CachedClient: Records List
        CachedClient->>Cache: Store in Cache
    end
    
    CachedClient->>CachedClient: Find Matching Record
    alt Record Found
        CachedClient-->>Resource: Return Record
        Resource->>Terraform: Update State
    else Record Not Found
        CachedClient-->>Resource: Return Nil
        Resource->>Terraform: Remove from State
    end
```

### DNS Record Deletion Flow

```mermaid
sequenceDiagram
    participant Terraform
    participant Resource
    participant CachedClient
    participant RetryLogic
    participant Cache
    participant DreamHostAPI
    
    Terraform->>Resource: Delete DNS Record
    Resource->>CachedClient: RemoveDNSRecord()
    CachedClient->>RetryLogic: Execute with Retry
    RetryLogic->>DreamHostAPI: POST /dns-remove_record
    
    alt Success
        DreamHostAPI-->>RetryLogic: Success Response
        RetryLogic-->>CachedClient: Success
        CachedClient->>Cache: Invalidate Cache
        CachedClient->>CachedClient: Wait for Deletion
        loop Check Record Deleted
            CachedClient->>DreamHostAPI: GET /dns-list_records
            DreamHostAPI-->>CachedClient: Records List
        end
        CachedClient-->>Resource: Record Deleted
        Resource->>Terraform: Remove from State
    else Error
        DreamHostAPI-->>RetryLogic: Error Response
        RetryLogic-->>Resource: Return Error
    end
```

### Data Source Query Flow

```mermaid
sequenceDiagram
    participant Terraform
    participant DataSource
    participant CachedClient
    participant Cache
    participant DreamHostAPI
    participant Filter
    
    Terraform->>DataSource: Query DNS Records
    DataSource->>CachedClient: ListDNSRecords()
    CachedClient->>Cache: Get All Records
    
    alt Cache Empty
        CachedClient->>DreamHostAPI: GET /dns-list_records
        DreamHostAPI-->>CachedClient: All Records
        CachedClient->>Cache: Store Records
    end
    
    Cache-->>CachedClient: Return Records
    CachedClient-->>DataSource: All Records
    DataSource->>Filter: Apply Filters
    Filter->>Filter: Match Criteria
    Filter-->>DataSource: Filtered Records
    DataSource->>Terraform: Return Results
```

## Module Descriptions

### Core Provider Module (`provider.go`)

**Responsibilities:**
- Provider configuration and initialization
- API key management
- Resource and data source registration
- Client instantiation

**Key Functions:**
- `Provider()`: Returns configured provider schema
- `providerConfigure()`: Initializes API client with credentials

### DNS Record Resource (`resource_dns_record.go`)

**Responsibilities:**
- CRUD operations for DNS records
- State management
- Import functionality
- Input validation coordination

**Key Functions:**
- `resourceDNSRecordCreate()`: Creates new DNS record
- `resourceDNSRecordRead()`: Reads existing record
- `resourceDNSRecordDelete()`: Removes DNS record
- `recordInputToID()`: Generates unique resource ID
- `idToRecordInput()`: Parses ID for import

### Data Sources

#### Single Record Lookup (`data_source_dns_record.go`)

**Responsibilities:**
- Query specific DNS record
- Handle ambiguous matches
- Populate computed fields

**Key Functions:**
- `dataSourceDNSRecordRead()`: Finds and returns single record

#### Multiple Records Query (`data_source_dns_records.go`)

**Responsibilities:**
- List all DNS records
- Apply filters
- Support partial matching

**Key Functions:**
- `dataSourceDNSRecordsRead()`: Returns filtered record list
- `filterDNSRecords()`: Applies filter criteria
- `matchesFilter()`: Evaluates individual record against filters

### Caching Layer

#### Cache Management (`cache.go`)

**Responsibilities:**
- Thread-safe record caching
- Cache invalidation
- Memory management

**Key Functions:**
- `GetRecords()`: Returns cached records or fetches new
- `Invalidate()`: Clears cache after modifications

#### Cached Client (`cached_client.go`)

**Responsibilities:**
- Wraps DreamHost API client
- Manages cache lifecycle
- Coordinates API calls

**Key Functions:**
- `AddDNSRecord()`: Adds record and invalidates cache
- `GetDNSRecord()`: Retrieves with cache support
- `RemoveDNSRecord()`: Removes and invalidates cache
- `ListDNSRecords()`: Lists all records

### Reliability Components

#### Retry Logic (`retry.go`)

**Responsibilities:**
- Retry transient failures
- Handle rate limiting
- Wait for eventual consistency

**Key Functions:**
- `retryOnError()`: Wraps operations with retry logic
- `isRetryableError()`: Determines retry eligibility
- `waitForDNSRecord()`: Polls until record appears
- `waitForDNSRecordDeletion()`: Polls until record removed

#### Validators (`validators.go`)

**Responsibilities:**
- DNS record type validation
- IP address validation
- Hostname validation
- Record-specific format validation

**Key Functions:**
- `ValidateDNSRecordName()`: Validates DNS names
- `ValidateIPv4Address()`: Validates IPv4 format
- `ValidateIPv6Address()`: Validates IPv6 format
- `ValidateMXRecord()`: Validates MX record format
- `ValidateSRVRecord()`: Validates SRV record format
- `ValidateDNSRecordValue()`: Type-specific validation

## Design Patterns

### 1. **Cached Wrapper Pattern**
The `cachedDreamhostClient` wraps the base API client to add caching capabilities transparently.

### 2. **Retry Pattern with Exponential Backoff**
All API operations use configurable retry logic to handle transient failures gracefully.

### 3. **Repository Pattern**
DNS records are accessed through a consistent interface regardless of cache state.

### 4. **Factory Pattern**
Provider configuration creates appropriate client instances based on configuration.

### 5. **Strategy Pattern**
Validators use different validation strategies based on DNS record type.

### 6. **Observer Pattern**
Cache invalidation occurs automatically when data modifications happen.

## Error Handling Strategy

```mermaid
graph TD
    A[API Operation] --> B{Error Occurred?}
    B -->|No| C[Return Success]
    B -->|Yes| D{Retryable?}
    D -->|Yes| E[Wait with Backoff]
    E --> F{Max Retries?}
    F -->|No| A
    F -->|Yes| G[Return Error]
    D -->|No| G
    G --> H{Critical Error?}
    H -->|Yes| I[Fail Resource Operation]
    H -->|No| J[Log Warning & Continue]
```

## Performance Optimizations

1. **Intelligent Caching**: Reduces API calls by caching list operations
2. **Parallel Operations**: Data sources can query concurrently
3. **Lazy Loading**: Cache populated only when needed
4. **Automatic Invalidation**: Cache cleared on modifications
5. **Efficient Filtering**: In-memory filtering reduces API load

## Security Considerations

1. **API Key Protection**: Sensitive field, supports environment variables
2. **No Logging of Secrets**: Error messages sanitized
3. **Input Validation**: All inputs validated before API calls
4. **Secure Defaults**: No hardcoded credentials

## Future Architecture Considerations

### Potential Enhancements

1. **Distributed Caching**: Redis/Memcached for multi-instance deployments
2. **Batch Operations**: Bulk create/delete capabilities
3. **Webhook Support**: Real-time updates via webhooks
4. **Metrics Collection**: Prometheus metrics for monitoring
5. **Circuit Breaker**: Prevent cascade failures
6. **Rate Limit Management**: Adaptive rate limiting

### Scalability Path

```mermaid
graph LR
    A[Current: In-Memory Cache] --> B[Phase 1: Persistent Cache]
    B --> C[Phase 2: Distributed Cache]
    C --> D[Phase 3: Multi-Region Support]
    D --> E[Phase 4: Event-Driven Updates]
```

## Dependencies

| Component | Version | Purpose |
|-----------|---------|---------|
| Terraform SDK | v2.26.1 | Provider framework |
| go-dreamhost | v0.1.1 | DreamHost API client |
| Go | 1.19+ | Runtime |

## Testing Architecture

```mermaid
graph TB
    subgraph "Test Types"
        UT[Unit Tests]
        IT[Integration Tests]
        AT[Acceptance Tests]
    end
    
    subgraph "Test Targets"
        V[Validators]
        C[Cache Logic]
        R[Retry Logic]
        RES[Resources]
        DS[Data Sources]
    end
    
    subgraph "Test Infrastructure"
        MT[Mock DreamHost API]
        TF[Test Fixtures]
        TC[Test Configuration]
    end
    
    UT --> V
    UT --> C
    UT --> R
    IT --> RES
    IT --> DS
    AT --> RES
    AT --> DS
    MT --> IT
    TF --> UT
    TC --> AT
```

---

This architecture ensures reliability, performance, and maintainability while providing a clear separation of concerns and extensibility for future enhancements.