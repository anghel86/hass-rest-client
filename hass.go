package homeassistant

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HomeAssistant struct {
	baseURL string
	token   string
	client  *http.Client
}

type State struct {
	EntityID string `json:"entity_id"`
	State    string `json:"state"`
}

type Entity struct {
	EntityID   string `json:"entity_id"`
	State      string `json:"state"`
	Attributes struct {
		FriendlyName string `json:"friendly_name"`
	} `json:"attributes"`
	LastChanged string `json:"last_changed"`
	LastUpdated string `json:"last_updated"`
}

type HistoryState struct {
	EntityID    string                `json:"entity_id"`
	State       string                `json:"state"`
	LastChanged string                `json:"last_changed"`
	LastUpdated string                `json:"last_updated"`
	Attributes  HistoryStateAttribute `json:"attributes"`
}

type HistoryStateAttribute struct {
	UnitOfMeasurement string `json:"unit_of_measurement"`
	FriendlyName      string `json:"state"`
}

func NewHomeAssistant(baseURL string, token string) *HomeAssistant {
	return &HomeAssistant{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{},
	}
}

func dateFormat(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func (h *HomeAssistant) GetStates() ([]*State, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/states", h.baseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var states []*State
	err = json.NewDecoder(resp.Body).Decode(&states)
	if err != nil {
		return nil, err
	}

	return states, nil
}

func (h *HomeAssistant) GetEntity(entityID string) (*Entity, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/states/%s", h.baseURL, entityID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var entity Entity
	err = json.NewDecoder(resp.Body).Decode(&entity)
	if err != nil {
		return nil, err
	}

	return &entity, nil
}

func (h *HomeAssistant) CallService(domain string, service string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/services/%s/%s", h.baseURL, domain, service), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (h *HomeAssistant) GetHistory(startTime time.Time, entityIds []string, endTime time.Time, minimalResponse bool, noAttributes bool, significantChangesOnly bool) ([]*HistoryState, error) {

	if startTime.IsZero() {
		return nil, errors.New("startTime parameter is null or not provided")
	}

	apiUrl := fmt.Sprintf("%s/api/history/period/%s", h.baseURL, dateFormat(startTime))

	queryParams := url.Values{}

	if entityIds != nil && len(entityIds) > 0 {
		queryParams.Set("filter_entity_id", strings.Join(entityIds, ","))
	}

	if !endTime.IsZero() {
		queryParams.Set("end_time", dateFormat(endTime))
	}

	if minimalResponse {
		queryParams.Set("minimal_response", "")
	}

	if noAttributes {
		queryParams.Set("no_attributes", "")
	}

	if significantChangesOnly {
		queryParams.Set("significant_changes_only", "")
	}

	fullUrl := apiUrl + "?" + queryParams.Encode()

	fmt.Printf("Hass url: %v\n", fullUrl)

	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var history []*HistoryState
	err = json.NewDecoder(resp.Body).Decode(&history)
	if err != nil {
		return nil, err
	}

	return history, nil
}
