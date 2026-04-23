package netnod

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://api.example.com", "test-token")

	if client.baseURL != "https://api.example.com" {
		t.Errorf("expected baseURL to be https://api.example.com, got %s", client.baseURL)
	}

	if client.token != "test-token" {
		t.Errorf("expected token to be test-token, got %s", client.token)
	}

	if client.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

func TestNewClient_DefaultURL(t *testing.T) {
	client := NewClient("", "test-token")

	if client.baseURL != DefaultAPIURL {
		t.Errorf("expected baseURL to be %s, got %s", DefaultAPIURL, client.baseURL)
	}
}

func TestClient_ListZones(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		if r.Header.Get("Authorization") != "Token test-token" {
			t.Errorf("expected Authorization header 'Token test-token', got %s", r.Header.Get("Authorization"))
		}

		response := ZoneListResponse{
			Data: []Zone{
				{ID: "example.com.", Name: "example.com.", NotifiedSerial: 2025010101},
				{ID: "test.com.", Name: "test.com.", NotifiedSerial: 2025010102},
			},
			Offset: 0,
			Limit:  1000,
			Total:  2,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	zones, err := client.ListZones()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(zones) != 2 {
		t.Errorf("expected 2 zones, got %d", len(zones))
	}

	if zones[0].Name != "example.com." {
		t.Errorf("expected first zone name to be example.com., got %s", zones[0].Name)
	}
}

func TestClient_ListZones_Pagination(t *testing.T) {
	allZones := []Zone{
		{ID: "a.com.", Name: "a.com.", NotifiedSerial: 1},
		{ID: "b.com.", Name: "b.com.", NotifiedSerial: 2},
		{ID: "c.com.", Name: "c.com.", NotifiedSerial: 3},
	}

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		q := r.URL.Query()
		offset := 0
		limit := 2
		if q.Get("offset") == "2" {
			offset = 2
		}

		end := offset + limit
		if end > len(allZones) {
			end = len(allZones)
		}

		response := ZoneListResponse{
			Data:   allZones[offset:end],
			Offset: offset,
			Limit:  limit,
			Total:  len(allZones),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	zones, err := client.ListZones()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(zones) != 3 {
		t.Errorf("expected 3 zones, got %d", len(zones))
	}

	if requestCount < 2 {
		t.Errorf("expected at least 2 requests for pagination, got %d", requestCount)
	}

	if zones[2].Name != "c.com." {
		t.Errorf("expected third zone name to be c.com., got %s", zones[2].Name)
	}
}

func TestClient_GetZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/zones/example.com." {
			t.Errorf("expected path /api/v1/zones/example.com., got %s", r.URL.Path)
		}

		zone := Zone{
			ID:             "example.com.",
			Name:           "example.com.",
			NotifiedSerial: 2025010101,
			AlsoNotify:     []string{"192.0.2.1"},
			RRsets: []RRset{
				{
					Name: "example.com.",
					Type: "A",
					Records: []Record{
						{Content: "192.0.2.10", Disabled: false},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(zone)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	zone, err := client.GetZone("example.com.")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if zone == nil {
		t.Fatal("expected zone to not be nil")
	}

	if zone.Name != "example.com." {
		t.Errorf("expected zone name to be example.com., got %s", zone.Name)
	}

	if len(zone.AlsoNotify) != 1 {
		t.Errorf("expected 1 also_notify, got %d", len(zone.AlsoNotify))
	}
}

func TestClient_GetZone_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "zone not found"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	zone, err := client.GetZone("nonexistent.com.")
	if err != nil {
		t.Fatalf("expected no error for 404, got %v", err)
	}

	if zone != nil {
		t.Error("expected zone to be nil for 404")
	}
}

func TestClient_CreateZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var zone Zone
		if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if zone.Name != "newzone.com." {
			t.Errorf("expected zone name newzone.com., got %s", zone.Name)
		}

		w.WriteHeader(http.StatusCreated)
		created := Zone{
			ID:             zone.Name,
			Name:           zone.Name,
			NotifiedSerial: 1,
		}
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	zone := &Zone{Name: "newzone.com."}
	created, err := client.CreateZone(zone)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.ID != "newzone.com." {
		t.Errorf("expected created zone ID to be newzone.com., got %s", created.ID)
	}
}

func TestClient_DeleteZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	err := client.DeleteZone("example.com.")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestClient_PatchZoneRRsets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH request, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		rrsets, ok := body["rrsets"].([]interface{})
		if !ok {
			t.Fatal("expected rrsets in body")
		}

		if len(rrsets) != 1 {
			t.Errorf("expected 1 rrset, got %d", len(rrsets))
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	ttl := int64(3600)
	rrsets := []RRset{
		{
			Name:       "www.example.com.",
			Type:       "A",
			TTL:        &ttl,
			ChangeType: "REPLACE",
			Records: []Record{
				{Content: "192.0.2.1", Disabled: false},
			},
		},
	}

	err := client.PatchZoneRRsets("example.com.", rrsets)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestClient_GetRRset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ttl := int64(3600)
		zone := Zone{
			ID:   "example.com.",
			Name: "example.com.",
			RRsets: []RRset{
				{
					Name: "www.example.com.",
					Type: "A",
					TTL:  &ttl,
					Records: []Record{
						{Content: "192.0.2.1", Disabled: false},
					},
				},
				{
					Name: "mail.example.com.",
					Type: "A",
					TTL:  &ttl,
					Records: []Record{
						{Content: "192.0.2.2", Disabled: false},
					},
				},
			},
		}

		json.NewEncoder(w).Encode(zone)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	rrset, err := client.GetRRset("example.com.", "www.example.com.", "A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rrset == nil {
		t.Fatal("expected rrset to not be nil")
	}

	if rrset.Name != "www.example.com." {
		t.Errorf("expected rrset name www.example.com., got %s", rrset.Name)
	}

	if len(rrset.Records) != 1 {
		t.Errorf("expected 1 record, got %d", len(rrset.Records))
	}
}

func TestClient_GetRRset_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zone := Zone{
			ID:     "example.com.",
			Name:   "example.com.",
			RRsets: []RRset{}, // No records
		}

		json.NewEncoder(w).Encode(zone)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	rrset, err := client.GetRRset("example.com.", "nonexistent.example.com.", "A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rrset != nil {
		t.Error("expected rrset to be nil for non-existent record")
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid zone name"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	_, err := client.CreateZone(&Zone{Name: "invalid"})
	if err == nil {
		t.Fatal("expected error for bad request")
	}

	expectedMsg := "API error (400): invalid zone name"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestClient_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid token"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token")

	_, err := client.ListZones()
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestClient_ListZones_EndcustomerFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("endcustomer") != "customer123" {
			t.Errorf("expected endcustomer=customer123, got %s", r.URL.Query().Get("endcustomer"))
		}

		response := ZoneListResponse{
			Data: []Zone{
				{ID: "example.com.", Name: "example.com.", EndCustomer: "customer123"},
			},
			Offset: 0,
			Limit:  1000,
			Total:  1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	zones, err := client.ListZones(WithEndCustomerName("customer123"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(zones))
	}

	if zones[0].EndCustomer != "customer123" {
		t.Errorf("expected endcustomer customer123, got %s", zones[0].EndCustomer)
	}
}

func TestClient_CreateZone_EndCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var zone Zone
		if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if zone.EndCustomer != "customer123" {
			t.Errorf("expected endcustomer customer123, got %s", zone.EndCustomer)
		}

		w.WriteHeader(http.StatusCreated)
		created := Zone{ID: zone.Name, Name: zone.Name, EndCustomer: zone.EndCustomer}
		json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	zone := &Zone{Name: "example.com.", EndCustomer: "customer123"}
	created, err := client.CreateZone(zone)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.EndCustomer != "customer123" {
		t.Errorf("expected endcustomer customer123, got %s", created.EndCustomer)
	}
}

func TestClient_GetZone_EndCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zone := Zone{
			ID:          "example.com.",
			Name:        "example.com.",
			EndCustomer: "customer123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(zone)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	zone, err := client.GetZone("example.com.")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if zone.EndCustomer != "customer123" {
		t.Errorf("expected endcustomer customer123, got %s", zone.EndCustomer)
	}
}
