package netnod

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type listZoneFilter struct {
	endCustomerName *string
}

// listZonesOption configures a ListZones filter.
type listZonesOption func(*listZoneFilter)

// WithEndCustomerName filters ListZones results to the given end customer.
func WithEndCustomerName(endCustomerName string) listZonesOption {
	return func(f *listZoneFilter) {
		f.endCustomerName = &endCustomerName
	}
}

// ListZones returns all zones, automatically paginating through all results.
func (c *Client) ListZones(options ...listZonesOption) ([]Zone, error) {
	var filter listZoneFilter
	for _, apply := range options {
		apply(&filter)
	}

	var allZones []Zone
	offset := 0
	limit := 1000

	for {
		path := fmt.Sprintf("/api/v1/zones?limit=%d&offset=%d", limit, offset)
		if filter.endCustomerName != nil {
			path += "&endcustomer=" + url.QueryEscape(*filter.endCustomerName)
		}
		resp, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			err := c.handleErrorResponse(resp)
			resp.Body.Close()
			return nil, err
		}

		var result ZoneListResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		allZones = append(allZones, result.Data...)

		offset += len(result.Data)
		if offset >= result.Total || len(result.Data) == 0 {
			break
		}
	}

	return allZones, nil
}

// GetZone returns a specific zone with all RRsets
func (c *Client) GetZone(zoneID string) (*Zone, error) {
	resp, err := c.doRequest("GET", "/api/v1/zones/"+url.PathEscape(zoneID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Zone not found
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var zone Zone
	if err := json.NewDecoder(resp.Body).Decode(&zone); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &zone, nil
}

// CreateZone creates a new zone
func (c *Client) CreateZone(zone *Zone) (*Zone, error) {
	resp, err := c.doRequest("POST", "/api/v1/zones", zone)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var created Zone
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &created, nil
}

// CreateZoneFromBIND creates a new zone using BIND zone file format
func (c *Client) CreateZoneFromBIND(zone *ZoneCreateBIND) (*Zone, error) {
	resp, err := c.doRequest("POST", "/api/v1/zones", zone)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var created Zone
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &created, nil
}

// UpdateZone updates zone configuration
func (c *Client) UpdateZone(zoneID string, zone *Zone) error {
	resp, err := c.doRequest("PUT", "/api/v1/zones/"+url.PathEscape(zoneID), zone)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// DeleteZone deletes a zone
func (c *Client) DeleteZone(zoneID string) error {
	resp, err := c.doRequest("DELETE", "/api/v1/zones/"+url.PathEscape(zoneID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// ExportZone exports a zone in BIND zone file format
func (c *Client) ExportZone(zoneID string) (string, error) {
	resp, err := c.doRequest("GET", "/api/v1/zones/"+url.PathEscape(zoneID)+"/export", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.handleErrorResponse(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// NotifyZone triggers an immediate DNS NOTIFY for a zone
func (c *Client) NotifyZone(zoneID string) (*NotifyResponse, error) {
	resp, err := c.doRequest("PUT", "/api/v1/zones/"+url.PathEscape(zoneID)+"/notify", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var result NotifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
