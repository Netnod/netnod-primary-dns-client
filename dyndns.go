package netnod

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ListDynDNS lists all DynDNS-enabled labels for a zone
func (c *Client) ListDynDNS(zoneID string) ([]DynDNSLabel, error) {
	resp, err := c.doRequest("GET", "/api/v1/zones/"+url.PathEscape(zoneID)+"/dyndns", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var result DynDNSListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Labels, nil
}

// CreateDynDNS enables DynDNS for a label in a zone
func (c *Client) CreateDynDNS(zoneID, label string) (*DynDNSCreateResponse, error) {
	resp, err := c.doRequest("POST", "/api/v1/zones/"+url.PathEscape(zoneID)+"/dyndns/"+url.PathEscape(label), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var result DynDNSCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteDynDNS disables DynDNS for a label in a zone
func (c *Client) DeleteDynDNS(zoneID, label string) error {
	resp, err := c.doRequest("DELETE", "/api/v1/zones/"+url.PathEscape(zoneID)+"/dyndns/"+url.PathEscape(label), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}

	return nil
}
