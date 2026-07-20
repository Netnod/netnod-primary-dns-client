package netnod

import (
	"strings"
	"testing"
)

func checkListZonesContains(t *testing.T, c *Client, zoneName string) {
	t.Helper()
	zones, err := c.ListZones()
	if err != nil {
		t.Fatalf("ListZones: %v", err)
	}
	for _, z := range zones {
		if z.Name == zoneName {
			return
		}
	}
	t.Errorf("zone %q not found in ListZones", zoneName)
}

func checkGetZone(t *testing.T, c *Client, zoneName string) *Zone {
	t.Helper()
	zone, err := c.GetZone(zoneName)
	if err != nil {
		t.Fatalf("GetZone(%q): %v", zoneName, err)
	}
	if zone == nil {
		t.Fatalf("GetZone(%q): returned nil", zoneName)
	}
	if zone.Name != zoneName {
		t.Errorf("GetZone(%q): zone.Name = %q", zoneName, zone.Name)
	}
	return zone
}

func checkGetZoneNotFound(t *testing.T, c *Client, zoneName string) {
	t.Helper()
	zone, err := c.GetZone(zoneName)
	if err != nil {
		t.Fatalf("GetZone(%q): expected nil error, got %v", zoneName, err)
	}
	if zone != nil {
		t.Errorf("GetZone(%q): expected nil, got %+v", zoneName, zone)
	}
}

func checkCreateZone(t *testing.T, c *Client, zoneName string) *Zone {
	t.Helper()
	created, err := c.CreateZone(&Zone{Name: zoneName})
	if err != nil {
		t.Fatalf("CreateZone(%q): %v", zoneName, err)
	}
	if created.Name != zoneName {
		t.Errorf("CreateZone(%q): returned zone.Name = %q", zoneName, created.Name)
	}
	return created
}

func checkDeleteZone(t *testing.T, c *Client, zoneName string) {
	t.Helper()
	if err := c.DeleteZone(zoneName); err != nil {
		t.Fatalf("DeleteZone(%q): %v", zoneName, err)
	}
}

func checkPatchZoneRRsets(t *testing.T, c *Client, zoneName string, rrsets []RRset) {
	t.Helper()
	if err := c.PatchZoneRRsets(zoneName, rrsets); err != nil {
		t.Fatalf("PatchZoneRRsets: %v", err)
	}
}

func checkGetRRset(t *testing.T, c *Client, zoneName, name, rrType, wantContent string) *RRset {
	t.Helper()
	rrset, err := c.GetRRset(zoneName, name, rrType)
	if err != nil {
		t.Fatalf("GetRRset(%q, %q, %q): %v", zoneName, name, rrType, err)
	}
	if rrset == nil {
		t.Fatalf("GetRRset(%q, %q, %q): returned nil", zoneName, name, rrType)
	}
	for _, r := range rrset.Records {
		if r.Content == wantContent {
			return rrset
		}
	}
	t.Errorf("GetRRset: content %q not found in records %+v", wantContent, rrset.Records)
	return rrset
}

func checkGetRRsetNotFound(t *testing.T, c *Client, zoneName, name, rrType string) {
	t.Helper()
	rrset, err := c.GetRRset(zoneName, name, rrType)
	if err != nil {
		t.Fatalf("GetRRset(%q, %q, %q): expected nil error, got %v", zoneName, name, rrType, err)
	}
	if rrset != nil {
		t.Errorf("GetRRset(%q, %q, %q): expected nil, got %+v", zoneName, name, rrType, rrset)
	}
}

func checkExportZone(t *testing.T, c *Client, zoneName string) string {
	t.Helper()
	export, err := c.ExportZone(zoneName)
	if err != nil {
		t.Fatalf("ExportZone(%q): %v", zoneName, err)
	}
	if export == "" {
		t.Fatalf("ExportZone(%q): returned empty string", zoneName)
	}
	if !strings.Contains(export, zoneName) {
		t.Errorf("ExportZone(%q): result does not contain zone name: %s", zoneName, export)
	}
	return export
}

func checkNotifyZone(t *testing.T, c *Client, zoneName string) {
	t.Helper()
	result, err := c.NotifyZone(zoneName)
	if err != nil {
		t.Fatalf("NotifyZone(%q): %v", zoneName, err)
	}
	if result.Result == "" {
		t.Errorf("NotifyZone(%q): Result field is empty", zoneName)
	}
}

func checkListDynDNS(t *testing.T, c *Client, zoneName string) []DynDNSLabel {
	t.Helper()
	labels, err := c.ListDynDNS(zoneName)
	if err != nil {
		t.Fatalf("ListDynDNS(%q): %v", zoneName, err)
	}
	return labels
}

func checkCreateDynDNS(t *testing.T, c *Client, zoneName, label string) *DynDNSCreateResponse {
	t.Helper()
	result, err := c.CreateDynDNS(zoneName, label)
	if err != nil {
		t.Fatalf("CreateDynDNS(%q, %q): %v", zoneName, label, err)
	}
	if result.Hostname == "" {
		t.Errorf("CreateDynDNS(%q, %q): Hostname is empty", zoneName, label)
	}
	if result.Token == "" {
		t.Errorf("CreateDynDNS(%q, %q): Token is empty", zoneName, label)
	}
	return result
}

func checkDeleteDynDNS(t *testing.T, c *Client, zoneName, label string) {
	t.Helper()
	if err := c.DeleteDynDNS(zoneName, label); err != nil {
		t.Fatalf("DeleteDynDNS(%q, %q): %v", zoneName, label, err)
	}
}

func checkListACME(t *testing.T, c *Client, zoneName string) []ACMELabel {
	t.Helper()
	labels, err := c.ListACME(zoneName)
	if err != nil {
		t.Fatalf("ListACME(%q): %v", zoneName, err)
	}
	return labels
}

func checkCreateACME(t *testing.T, c *Client, zoneName, label string) *ACMECreateResponse {
	t.Helper()
	result, err := c.CreateACME(zoneName, label)
	if err != nil {
		t.Fatalf("CreateACME(%q, %q): %v", zoneName, label, err)
	}
	if result.Hostname == "" {
		t.Errorf("CreateACME(%q, %q): Hostname is empty", zoneName, label)
	}
	if result.ChallengeHostname == "" {
		t.Errorf("CreateACME(%q, %q): ChallengeHostname is empty", zoneName, label)
	}
	if result.Token == "" {
		t.Errorf("CreateACME(%q, %q): Token is empty", zoneName, label)
	}
	return result
}

func checkDeleteACME(t *testing.T, c *Client, zoneName, label string) {
	t.Helper()
	if err := c.DeleteACME(zoneName, label); err != nil {
		t.Fatalf("DeleteACME(%q, %q): %v", zoneName, label, err)
	}
}
