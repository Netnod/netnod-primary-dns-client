package netnod

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Token test-token" {
			t.Errorf("unexpected Authorization header: %s", r.Header.Get("Authorization"))
		}
		json.NewEncoder(w).Encode(ZoneListResponse{
			Data:  []Zone{{ID: "example.com.", Name: "example.com."}, {ID: "test.com.", Name: "test.com."}},
			Total: 2, Limit: 1000,
		})
	}))
	defer server.Close()

	checkListZonesContains(t, NewClient(server.URL, "test-token"), "example.com.")
}

func TestClient_ListZones_Pagination(t *testing.T) {
	all := []Zone{
		{ID: "a.com.", Name: "a.com."},
		{ID: "b.com.", Name: "b.com."},
		{ID: "c.com.", Name: "c.com."},
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		offset := 0
		if r.URL.Query().Get("offset") == "2" {
			offset = 2
		}
		end := offset + 2
		if end > len(all) {
			end = len(all)
		}
		json.NewEncoder(w).Encode(ZoneListResponse{Data: all[offset:end], Offset: offset, Limit: 2, Total: len(all)})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	zones, err := client.ListZones()
	if err != nil {
		t.Fatalf("ListZones: %v", err)
	}
	if len(zones) != 3 {
		t.Errorf("expected 3 zones, got %d", len(zones))
	}
	if requests < 2 {
		t.Errorf("expected at least 2 requests for pagination, got %d", requests)
	}
}

func TestClient_GetZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/zones/example.com." {
			t.Errorf("expected path /api/v1/zones/example.com., got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Zone{
			ID: "example.com.", Name: "example.com.",
			AlsoNotify: []string{"192.0.2.1"},
		})
	}))
	defer server.Close()

	zone := checkGetZone(t, NewClient(server.URL, "test-token"), "example.com.")
	if len(zone.AlsoNotify) != 1 {
		t.Errorf("expected 1 also_notify, got %v", zone.AlsoNotify)
	}
}

func TestClient_GetZone_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "zone not found"})
	}))
	defer server.Close()

	checkGetZoneNotFound(t, NewClient(server.URL, "test-token"), "nonexistent.com.")
}

func TestClient_CreateZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}
		var zone Zone
		if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if zone.Name != "newzone.com." {
			t.Errorf("expected zone name newzone.com., got %s", zone.Name)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Zone{ID: zone.Name, Name: zone.Name})
	}))
	defer server.Close()

	checkCreateZone(t, NewClient(server.URL, "test-token"), "newzone.com.")
}

func TestClient_CreateZoneFromBIND(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}
		var zone ZoneCreateBIND
		if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if zone.Name != "example.com." {
			t.Errorf("expected zone name example.com., got %s", zone.Name)
		}
		if zone.Zone == "" {
			t.Error("expected zone field to be non-empty")
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Zone{ID: zone.Name, Name: zone.Name})
	}))
	defer server.Close()

	result, err := NewClient(server.URL, "test-token").CreateZoneFromBIND(&ZoneCreateBIND{
		Name: "example.com.",
		Zone: "example.com.\t3600\tIN\tSOA\tns1.example.com. host.example.com. 2025110401 10800 3600 604800 3600",
	})
	if err != nil {
		t.Fatalf("CreateZoneFromBIND: %v", err)
	}
	if result.Name != "example.com." {
		t.Errorf("expected zone name example.com., got %s", result.Name)
	}
}

func TestClient_CreateZoneFromBIND_NotifyAndTransferKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		alsoNotify, ok := body["also_notify"].([]interface{})
		if !ok || len(alsoNotify) != 1 || alsoNotify[0] != "1.2.3.4" {
			t.Errorf("expected also_notify [1.2.3.4], got %v", body["also_notify"])
		}
		keys, ok := body["allow_transfer_keys"].([]interface{})
		if !ok || len(keys) != 1 || keys[0] != "key-abc" {
			t.Errorf("expected allow_transfer_keys [key-abc], got %v", body["allow_transfer_keys"])
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Zone{ID: "example.com.", Name: "example.com."})
	}))
	defer server.Close()

	_, err := NewClient(server.URL, "test-token").CreateZoneFromBIND(&ZoneCreateBIND{
		Name:              "example.com.",
		Zone:              "example.com.\t3600\tIN\tSOA\tns1.example.com. host.example.com. 2025110401 10800 3600 604800 3600",
		AlsoNotify:        []string{"1.2.3.4"},
		AllowTransferKeys: []string{"key-abc"},
	})
	if err != nil {
		t.Fatalf("CreateZoneFromBIND: %v", err)
	}
}

func TestClient_UpdateZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/zones/example.com." {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if _, ok := body["name"]; ok {
			t.Error("body must not contain 'name' key (disallowed by API)")
		}
		alsoNotify, ok := body["also_notify"].([]interface{})
		if !ok || len(alsoNotify) != 1 || alsoNotify[0] != "1.2.3.4" {
			t.Errorf("expected also_notify [1.2.3.4], got %v", body["also_notify"])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := NewClient(server.URL, "test-token").UpdateZone("example.com.", &Zone{
		Name:       "example.com.",
		AlsoNotify: []string{"1.2.3.4"},
	})
	if err != nil {
		t.Fatalf("UpdateZone: %v", err)
	}
}

func TestClient_UpdateZone_NilZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if strings.TrimSpace(string(body)) != "null" {
			t.Errorf("expected null body, got %q", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := NewClient(server.URL, "test-token").UpdateZone("example.com.", nil); err != nil {
		t.Fatalf("UpdateZone: %v", err)
	}
}

func TestClient_UpdateZone_AllowTransferKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		keys, ok := body["allow_transfer_keys"].([]interface{})
		if !ok || len(keys) != 1 || keys[0] != "key-abc" {
			t.Errorf("expected allow_transfer_keys [key-abc], got %v", body["allow_transfer_keys"])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := NewClient(server.URL, "test-token").UpdateZone("example.com.", &Zone{
		AllowTransferKeys: []string{"key-abc"},
	})
	if err != nil {
		t.Fatalf("UpdateZone: %v", err)
	}
}

func TestClient_DeleteZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	checkDeleteZone(t, NewClient(server.URL, "test-token"), "example.com.")
}

func TestClient_ExportZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/zones/example.com./export" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("example.com.\t3600\tIN\tSOA\tns1.example.com. host.example.com. 2025110401 10800 3600 604800 3600\n"))
	}))
	defer server.Close()

	checkExportZone(t, NewClient(server.URL, "test-token"), "example.com.")
}

func TestClient_NotifyZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/zones/example.com./notify" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(NotifyResponse{Result: "Notification queued"})
	}))
	defer server.Close()

	checkNotifyZone(t, NewClient(server.URL, "test-token"), "example.com.")
}

func TestClient_PatchZoneRRsets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		rrsets, ok := body["rrsets"].([]interface{})
		if !ok || len(rrsets) != 1 {
			t.Errorf("expected 1 rrset in body, got %v", body["rrsets"])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ttl := int64(3600)
	checkPatchZoneRRsets(t, NewClient(server.URL, "test-token"), "example.com.", []RRset{
		{Name: "www.example.com.", Type: "A", TTL: &ttl, ChangeType: "REPLACE",
			Records: []Record{{Content: "192.0.2.1"}}},
	})
}

func TestClient_PatchZoneRRsets_DELETE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		rrsets, ok := body["rrsets"].([]interface{})
		if !ok || len(rrsets) == 0 {
			t.Fatal("expected rrsets in body")
		}
		rrset := rrsets[0].(map[string]interface{})
		if rrset["changetype"] != "DELETE" {
			t.Errorf("expected changetype DELETE, got %v", rrset["changetype"])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	checkPatchZoneRRsets(t, NewClient(server.URL, "test-token"), "example.com.", []RRset{
		{Name: "www.example.com.", Type: "A", ChangeType: "DELETE", Records: []Record{}},
	})
}

func TestClient_PatchZoneRRsets_EXTEND(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		rrsets, ok := body["rrsets"].([]interface{})
		if !ok || len(rrsets) == 0 {
			t.Fatal("expected rrsets in body")
		}
		rrset := rrsets[0].(map[string]interface{})
		if rrset["changetype"] != "EXTEND" {
			t.Errorf("expected changetype EXTEND, got %v", rrset["changetype"])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ttl := int64(3600)
	checkPatchZoneRRsets(t, NewClient(server.URL, "test-token"), "example.com.", []RRset{
		{Name: "www.example.com.", Type: "A", TTL: &ttl, ChangeType: "EXTEND",
			Records: []Record{{Content: "192.0.2.2"}}},
	})
}

func TestClient_PatchZoneRRsets_PRUNE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		rrsets, ok := body["rrsets"].([]interface{})
		if !ok || len(rrsets) == 0 {
			t.Fatal("expected rrsets in body")
		}
		rrset := rrsets[0].(map[string]interface{})
		if rrset["changetype"] != "PRUNE" {
			t.Errorf("expected changetype PRUNE, got %v", rrset["changetype"])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ttl := int64(3600)
	checkPatchZoneRRsets(t, NewClient(server.URL, "test-token"), "example.com.", []RRset{
		{Name: "www.example.com.", Type: "A", TTL: &ttl, ChangeType: "PRUNE",
			Records: []Record{{Content: "192.0.2.1"}}},
	})
}

func TestClient_GetRRset(t *testing.T) {
	ttl := int64(3600)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Zone{
			ID: "example.com.", Name: "example.com.",
			RRsets: []RRset{
				{Name: "www.example.com.", Type: "A", TTL: &ttl, Records: []Record{{Content: "192.0.2.1"}}},
				{Name: "mail.example.com.", Type: "A", TTL: &ttl, Records: []Record{{Content: "192.0.2.2"}}},
			},
		})
	}))
	defer server.Close()

	checkGetRRset(t, NewClient(server.URL, "test-token"), "example.com.", "www.example.com.", "A", "192.0.2.1")
}

func TestClient_GetRRset_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Zone{ID: "example.com.", Name: "example.com.", RRsets: []RRset{}})
	}))
	defer server.Close()

	checkGetRRsetNotFound(t, NewClient(server.URL, "test-token"), "example.com.", "nonexistent.example.com.", "A")
}

func TestClient_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid zone name"})
	}))
	defer server.Close()

	_, err := NewClient(server.URL, "test-token").CreateZone(&Zone{Name: "invalid"})
	if err == nil {
		t.Fatal("expected error for bad request")
	}
	want := "API error (400): invalid zone name"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestClient_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid token"})
	}))
	defer server.Close()

	_, err := NewClient(server.URL, "bad-token").ListZones()
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestClient_ListZones_EndCustomerFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/zones" {
			t.Errorf("expected path /api/v1/zones, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("endcustomer") != "customer123" {
			t.Errorf("expected endcustomer=customer123, got %s", r.URL.Query().Get("endcustomer"))
		}
		json.NewEncoder(w).Encode(ZoneListResponse{
			Data:  []Zone{{ID: "example.com.", Name: "example.com.", EndCustomer: "customer123"}},
			Total: 1, Limit: 1000,
		})
	}))
	defer server.Close()

	zones, err := NewClient(server.URL, "test-token").ListZones(WithEndCustomerName("customer123"))
	if err != nil {
		t.Fatalf("ListZones: %v", err)
	}
	if len(zones) != 1 || zones[0].EndCustomer != "customer123" {
		t.Errorf("unexpected zones: %+v", zones)
	}
}

func TestClient_CreateZone_EndCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var zone Zone
		if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if zone.EndCustomer != "customer123" {
			t.Errorf("expected endcustomer customer123, got %s", zone.EndCustomer)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Zone{ID: zone.Name, Name: zone.Name, EndCustomer: zone.EndCustomer})
	}))
	defer server.Close()

	created, err := NewClient(server.URL, "test-token").CreateZone(&Zone{Name: "example.com.", EndCustomer: "customer123"})
	if err != nil {
		t.Fatalf("CreateZone: %v", err)
	}
	if created.EndCustomer != "customer123" {
		t.Errorf("expected endcustomer customer123, got %s", created.EndCustomer)
	}
}

func TestClient_GetZone_EndCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Zone{ID: "example.com.", Name: "example.com.", EndCustomer: "customer123"})
	}))
	defer server.Close()

	zone := checkGetZone(t, NewClient(server.URL, "test-token"), "example.com.")
	if zone.EndCustomer != "customer123" {
		t.Errorf("expected endcustomer customer123, got %s", zone.EndCustomer)
	}
}

func TestClient_ListDynDNS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/zones/example.com./dyndns" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(DynDNSListResponse{
			Labels: []DynDNSLabel{{Label: "home", Hostname: "home.example.com."}},
		})
	}))
	defer server.Close()

	labels := checkListDynDNS(t, NewClient(server.URL, "test-token"), "example.com.")
	if len(labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(labels))
	}
}

func TestClient_CreateDynDNS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/zones/example.com./dyndns/home" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(DynDNSCreateResponse{
			Hostname: "home.example.com.",
			Token:    "secret-token",
		})
	}))
	defer server.Close()

	checkCreateDynDNS(t, NewClient(server.URL, "test-token"), "example.com.", "home")
}

func TestClient_DeleteDynDNS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/zones/example.com./dyndns/home" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	checkDeleteDynDNS(t, NewClient(server.URL, "test-token"), "example.com.", "home")
}

func TestClient_ListACME(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/zones/example.com./acme" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(ACMEListResponse{
			Labels: []ACMELabel{{
				Label:             "www",
				Hostname:          "www.example.com.",
				ChallengeHostname: "_acme-challenge.www.example.com.",
			}},
		})
	}))
	defer server.Close()

	labels := checkListACME(t, NewClient(server.URL, "test-token"), "example.com.")
	if len(labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(labels))
	}
}

func TestClient_CreateACME(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/zones/example.com./acme/www" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ACMECreateResponse{
			Hostname:          "www.example.com.",
			ChallengeHostname: "_acme-challenge.www.example.com.",
			Token:             "secret-token",
		})
	}))
	defer server.Close()

	checkCreateACME(t, NewClient(server.URL, "test-token"), "example.com.", "www")
}

func TestClient_DeleteACME(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/zones/example.com./acme/www" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	checkDeleteACME(t, NewClient(server.URL, "test-token"), "example.com.", "www")
}
