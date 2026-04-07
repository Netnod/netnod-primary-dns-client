package netnod

import (
	"fmt"
	"net/http"
	"net/url"
)

// PatchZoneRRsets updates specific RRsets in a zone
func (c *Client) PatchZoneRRsets(zoneID string, rrsets []RRset) error {
	body := map[string]interface{}{
		"rrsets": rrsets,
	}

	resp, err := c.doRequest("PATCH", "/api/v1/zones/"+url.PathEscape(zoneID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// GetRRset finds a specific RRset in a zone
func (c *Client) GetRRset(zoneID, name, rrType string) (*RRset, error) {
	zone, err := c.GetZone(zoneID)
	if err != nil {
		return nil, err
	}
	if zone == nil {
		return nil, fmt.Errorf("zone %s not found", zoneID)
	}

	for _, rrset := range zone.RRsets {
		if rrset.Name == name && rrset.Type == rrType {
			return &rrset, nil
		}
	}

	return nil, nil // RRset not found
}
