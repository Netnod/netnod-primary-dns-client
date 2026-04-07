package netnod

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ListACME lists all ACME-enabled labels for a zone
func (c *Client) ListACME(zoneID string) ([]ACMELabel, error) {
	resp, err := c.doRequest("GET", "/api/v1/zones/"+url.PathEscape(zoneID)+"/acme", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var result ACMEListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Labels, nil
}

// CreateACME enables ACME DNS-01 challenges for a label in a zone
func (c *Client) CreateACME(zoneID, label string) (*ACMECreateResponse, error) {
	resp, err := c.doRequest("POST", "/api/v1/zones/"+url.PathEscape(zoneID)+"/acme/"+url.PathEscape(label), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var result ACMECreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteACME disables ACME for a label in a zone
func (c *Client) DeleteACME(zoneID, label string) error {
	resp, err := c.doRequest("DELETE", "/api/v1/zones/"+url.PathEscape(zoneID)+"/acme/"+url.PathEscape(label), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}

	return nil
}
