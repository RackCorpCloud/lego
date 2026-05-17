package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DNSRecord struct {
	ID       json.Number `json:"id"`
	Lookup   string      `json:"lookup"`
	Type     string      `json:"type"`
	Data     string      `json:"data"`
	TTL      json.Number `json:"ttl"`
	DomainID json.Number `json:"domainid"`
}

type DNSDomain struct {
	ID      json.Number `json:"id"`
	Name    string      `json:"name"`
	Records []DNSRecord `json:"records"`
}

type response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const responseCodeOK = "OK"
const recordTypeTXT = "TXT"

type dnsDomainGetResponse struct {
	response
	Data DNSDomain `json:"data"`
}

type dnsDomainGetAllResponse struct {
	response
	Data []DNSDomain `json:"data"`
}

const DefaultURL = "https://api.rackcorp.com/api/rest/v2.9/json.php"

type RCClient struct {
	client    *http.Client
	url       string
	apiUUID   string
	apiSecret string
	userAgent string
}

func NewRCClient(client *http.Client, url, apiUUID, apiSecret, userAgent string) *RCClient {
	return &RCClient{
		client:    client,
		url:       url,
		apiUUID:   apiUUID,
		apiSecret: apiSecret,
		userAgent: userAgent,
	}
}

func (c *RCClient) apiReq(payload map[string]any) (*http.Response, error) {
	reqBodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.url,
		bytes.NewReader(reqBodyBytes),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.SetBasicAuth(c.apiUUID, c.apiSecret)

	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func doApi[T any](c *RCClient, payload map[string]any, result *T) error {
	resp, err := c.apiReq(payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(result)
	return err
}

func (c *RCClient) DNSDomainGetAll() ([]DNSDomain, error) {
	reqPayload := map[string]any{
		"cmd": "dns.domain.getall",
	}

	var apiResp dnsDomainGetAllResponse
	err := doApi(c, reqPayload, &apiResp)
	if err != nil {
		return nil, err
	}

	if apiResp.Code != responseCodeOK {
		return nil, fmt.Errorf("api error: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

func (c *RCClient) DNSDomainGet(domainID json.Number) (*DNSDomain, error) {
	reqPayload := map[string]any{
		"cmd": "dns.domain.get",
		"id":  domainID,
	}

	var apiResp dnsDomainGetResponse
	err := doApi(c, reqPayload, &apiResp)
	if err != nil {
		return nil, err
	}

	if apiResp.Code != responseCodeOK {
		return nil, fmt.Errorf("api error: %s", apiResp.Message)
	}

	return &apiResp.Data, nil
}

func (c *RCClient) DNSRecordCreateTXT(domainID json.Number, lookup, data string, ttl int) error {
	record := DNSRecord{
		ID:       "0",
		Lookup:   lookup,
		Type:     recordTypeTXT,
		Data:     data,
		TTL:      json.Number(fmt.Sprintf("%d", ttl)),
		DomainID: domainID,
	}

	reqPayload := map[string]any{
		"cmd": "dns.record.create",
		"data": map[json.Number]DNSRecord{
			record.ID: record,
		},
	}

	var apiResp response
	err := doApi(c, reqPayload, &apiResp)
	if err != nil {
		return err
	}

	if apiResp.Code != responseCodeOK {
		return fmt.Errorf("api error: %s", apiResp.Message)
	}

	return nil
}

func (c *RCClient) DNSRecordUpdate(record DNSRecord) error {
	reqPayload := map[string]any{
		"cmd": "dns.record.update",
		"data": map[json.Number]DNSRecord{
			record.ID: record,
		},
	}

	var apiResp response
	err := doApi(c, reqPayload, &apiResp)
	if err != nil {
		return err
	}

	if apiResp.Code != responseCodeOK {
		return fmt.Errorf("api error: %s", apiResp.Message)
	}

	return nil
}

func (c *RCClient) DNSRecordDelete(recordID json.Number) error {
	reqPayload := map[string]any{
		"cmd": "dns.record.delete",
		"id":  recordID,
	}

	var apiResp response
	err := doApi(c, reqPayload, &apiResp)
	if err != nil {
		return err
	}

	if apiResp.Code != responseCodeOK {
		return fmt.Errorf("api error: %s", apiResp.Message)
	}

	return nil
}

func FindDomain(domains []DNSDomain, domainName string) *DNSDomain {
	for _, domain := range domains {
		if domain.Name == domainName {
			return &domain
		}
	}
	return nil
}

func FindTXTRecord(records []DNSRecord, lookup string) *DNSRecord {
	for _, record := range records {
		if record.Lookup == lookup && record.Type == recordTypeTXT {
			return &record
		}
	}
	return nil
}
