//go:build integration

package netnod

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var (
	itClient       *Client
	itZone         string
	itMissingCreds string
)

func TestMain(m *testing.M) {
	token := os.Getenv("NETNOD_API_TOKEN")
	zone := os.Getenv("NETNOD_TEST_ZONE")

	if token == "" || zone == "" {
		itMissingCreds = "set NETNOD_API_TOKEN and NETNOD_TEST_ZONE to run integration tests"
		os.Exit(m.Run())
	}

	if !strings.HasSuffix(zone, ".") {
		zone += "."
	}

	itClient = NewClient(os.Getenv("NETNOD_API_URL"), token)
	itZone = zone

	if _, err := itClient.CreateZone(&Zone{Name: itZone}); err != nil {
		fmt.Fprintf(os.Stderr, "setup: CreateZone %q: %v\n", itZone, err)
		os.Exit(1)
	}

	code := m.Run()

	if err := itClient.DeleteZone(itZone); err != nil {
		fmt.Fprintf(os.Stderr, "teardown: DeleteZone %q: %v\n", itZone, err)
	}

	os.Exit(code)
}

func requireCredentials(t *testing.T) {
	t.Helper()
	if itMissingCreds != "" {
		t.Fatal(itMissingCreds)
	}
}

func TestIntegration_ListZones(t *testing.T) {
	requireCredentials(t)
	checkListZonesContains(t, itClient, itZone)
}

func TestIntegration_GetZone(t *testing.T) {
	requireCredentials(t)
	checkGetZone(t, itClient, itZone)
}

func TestIntegration_GetZone_NotFound(t *testing.T) {
	requireCredentials(t)
	checkGetZoneNotFound(t, itClient, "this-zone-does-not-exist.invalid.")
}

func TestIntegration_ExportZone(t *testing.T) {
	requireCredentials(t)
	checkExportZone(t, itClient, itZone)
}

func TestIntegration_NotifyZone(t *testing.T) {
	requireCredentials(t)
	checkNotifyZone(t, itClient, itZone)
}

func TestIntegration_UpdateZone(t *testing.T) {
	requireCredentials(t)

	if err := itClient.UpdateZone(itZone, &Zone{
		AlsoNotify: []string{"192.0.2.253"},
	}); err != nil {
		t.Fatalf("UpdateZone: %v", err)
	}

	zone := checkGetZone(t, itClient, itZone)
	found := false
	for _, ip := range zone.AlsoNotify {
		if ip == "192.0.2.253" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AlsoNotify does not contain 192.0.2.253: %v", zone.AlsoNotify)
	}
}

func TestIntegration_PatchZoneRRsets(t *testing.T) {
	requireCredentials(t)

	ttl := int64(300)
	recordName := "www." + itZone

	t.Cleanup(func() {
		_ = itClient.PatchZoneRRsets(itZone, []RRset{
			{Name: recordName, Type: "A", ChangeType: "DELETE", Records: []Record{}},
		})
	})

	checkPatchZoneRRsets(t, itClient, itZone, []RRset{
		{Name: recordName, Type: "A", TTL: &ttl, ChangeType: "REPLACE",
			Records: []Record{{Content: "192.0.2.1"}}},
	})
	checkGetRRset(t, itClient, itZone, recordName, "A", "192.0.2.1")
}

func TestIntegration_PatchZoneRRsets_EXTEND(t *testing.T) {
	requireCredentials(t)

	ttl := int64(300)
	recordName := "extend-test." + itZone

	t.Cleanup(func() {
		_ = itClient.PatchZoneRRsets(itZone, []RRset{
			{Name: recordName, Type: "A", ChangeType: "DELETE", Records: []Record{}},
		})
	})

	checkPatchZoneRRsets(t, itClient, itZone, []RRset{
		{Name: recordName, Type: "A", TTL: &ttl, ChangeType: "REPLACE",
			Records: []Record{{Content: "192.0.2.1"}}},
	})
	checkPatchZoneRRsets(t, itClient, itZone, []RRset{
		{Name: recordName, Type: "A", TTL: &ttl, ChangeType: "EXTEND",
			Records: []Record{{Content: "192.0.2.2"}}},
	})

	checkGetRRset(t, itClient, itZone, recordName, "A", "192.0.2.1")
	checkGetRRset(t, itClient, itZone, recordName, "A", "192.0.2.2")
}

func TestIntegration_PatchZoneRRsets_PRUNE(t *testing.T) {
	requireCredentials(t)

	ttl := int64(300)
	recordName := "prune-test." + itZone

	t.Cleanup(func() {
		_ = itClient.PatchZoneRRsets(itZone, []RRset{
			{Name: recordName, Type: "A", ChangeType: "DELETE", Records: []Record{}},
		})
	})

	checkPatchZoneRRsets(t, itClient, itZone, []RRset{
		{Name: recordName, Type: "A", TTL: &ttl, ChangeType: "REPLACE",
			Records: []Record{{Content: "192.0.2.1"}, {Content: "192.0.2.2"}}},
	})
	checkPatchZoneRRsets(t, itClient, itZone, []RRset{
		{Name: recordName, Type: "A", TTL: &ttl, ChangeType: "PRUNE",
			Records: []Record{{Content: "192.0.2.1"}}},
	})

	checkGetRRset(t, itClient, itZone, recordName, "A", "192.0.2.2")

	rrset, err := itClient.GetRRset(itZone, recordName, "A")
	if err != nil {
		t.Fatalf("GetRRset after PRUNE: %v", err)
	}
	for _, r := range rrset.Records {
		if r.Content == "192.0.2.1" {
			t.Error("pruned record 192.0.2.1 still present")
		}
	}
}

func TestIntegration_PatchZoneRRsets_DELETE(t *testing.T) {
	requireCredentials(t)

	ttl := int64(300)
	recordName := "delete-test." + itZone

	t.Cleanup(func() {
		_ = itClient.PatchZoneRRsets(itZone, []RRset{
			{Name: recordName, Type: "A", ChangeType: "DELETE", Records: []Record{}},
		})
	})

	checkPatchZoneRRsets(t, itClient, itZone, []RRset{
		{Name: recordName, Type: "A", TTL: &ttl, ChangeType: "REPLACE",
			Records: []Record{{Content: "192.0.2.1"}}},
	})
	checkPatchZoneRRsets(t, itClient, itZone, []RRset{
		{Name: recordName, Type: "A", ChangeType: "DELETE", Records: []Record{}},
	})
	checkGetRRsetNotFound(t, itClient, itZone, recordName, "A")
}

func TestIntegration_GetRRset_NotFound(t *testing.T) {
	requireCredentials(t)
	checkGetRRsetNotFound(t, itClient, itZone, "nonexistent."+itZone, "A")
}

func TestIntegration_DynDNS(t *testing.T) {
	requireCredentials(t)

	label := "integration-test"

	t.Cleanup(func() {
		_ = itClient.DeleteDynDNS(itZone, label)
	})

	checkCreateDynDNS(t, itClient, itZone, label)

	labels := checkListDynDNS(t, itClient, itZone)
	found := false
	for _, l := range labels {
		if l.Label == label {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("label %q not found in ListDynDNS", label)
	}

	checkDeleteDynDNS(t, itClient, itZone, label)

	labels = checkListDynDNS(t, itClient, itZone)
	for _, l := range labels {
		if l.Label == label {
			t.Errorf("label %q still present after DeleteDynDNS", label)
		}
	}
}

func TestIntegration_ACME(t *testing.T) {
	requireCredentials(t)

	label := "integration-test"

	t.Cleanup(func() {
		_ = itClient.DeleteACME(itZone, label)
	})

	checkCreateACME(t, itClient, itZone, label)

	labels := checkListACME(t, itClient, itZone)
	found := false
	for _, l := range labels {
		if l.Label == label {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("label %q not found in ListACME", label)
	}

	checkDeleteACME(t, itClient, itZone, label)

	labels = checkListACME(t, itClient, itZone)
	for _, l := range labels {
		if l.Label == label {
			t.Errorf("label %q still present after DeleteACME", label)
		}
	}
}
