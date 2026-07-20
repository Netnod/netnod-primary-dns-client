// Package netnod provides a Go client for the Netnod Primary DNS API.
package netnod

// Zone represents a DNS zone
type Zone struct {
	ID                string   `json:"id,omitempty"`
	Name              string   `json:"name,omitempty"`
	NotifiedSerial    int64    `json:"notified_serial,omitempty"`
	AlsoNotify        []string `json:"also_notify,omitempty"`
	AllowTransferKeys []string `json:"allow_transfer_keys,omitempty"`
	EndCustomer       string   `json:"endcustomer,omitempty"`
	RRsets            []RRset  `json:"rrsets,omitempty"`
}

// RRset represents a resource record set
type RRset struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	TTL        *int64   `json:"ttl,omitempty"`
	ChangeType string   `json:"changetype,omitempty"`
	Records    []Record `json:"records"`
}

// Record represents a single DNS record
type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

// ZoneListResponse represents the paginated zone list response
type ZoneListResponse struct {
	Data   []Zone `json:"data"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Total  int    `json:"total"`
}

// ZoneCreateBIND represents a zone creation request using BIND zone file format
type ZoneCreateBIND struct {
	Name              string   `json:"name"`
	Zone              string   `json:"zone"`
	AlsoNotify        []string `json:"also_notify,omitempty"`
	AllowTransferKeys []string `json:"allow_transfer_keys,omitempty"`
	EndCustomer       string   `json:"endcustomer,omitempty"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error string `json:"error"`
}

// NotifyResponse represents the response from a notify request
type NotifyResponse struct {
	Result string `json:"result"`
}

// DynDNSLabel represents a DynDNS-enabled label
type DynDNSLabel struct {
	Label    string `json:"label"`
	Hostname string `json:"hostname"`
}

// DynDNSListResponse represents the response from listing DynDNS labels
type DynDNSListResponse struct {
	Labels []DynDNSLabel `json:"labels"`
}

// DynDNSCreateResponse represents the response from enabling DynDNS
type DynDNSCreateResponse struct {
	Hostname string `json:"hostname"`
	Token    string `json:"token"`
}

// ACMELabel represents an ACME-enabled label
type ACMELabel struct {
	Label             string `json:"label"`
	Hostname          string `json:"hostname"`
	ChallengeHostname string `json:"challenge_hostname"`
}

// ACMEListResponse represents the response from listing ACME labels
type ACMEListResponse struct {
	Labels []ACMELabel `json:"labels"`
}

// ACMECreateResponse represents the response from enabling ACME
type ACMECreateResponse struct {
	Hostname          string `json:"hostname"`
	ChallengeHostname string `json:"challenge_hostname"`
	Token             string `json:"token"`
}
