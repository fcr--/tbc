package terabox

import (
	"encoding/json"
	"fmt"
	"tbc/internal/util"

	"github.com/tidwall/gjson"
)

type Quota struct {
	Free     uint64 `json:"free"`
	SBoxUsed uint64 `json:"sbox_used"`
	Total    uint64 `json:"total"`
	Used     uint64 `json:"used"`
}

func (c *Client) GetQuota() (*Quota, error) {
	body, err := util.GetResponse(c.client.R().
		SetQueryParams(map[string]string{
			"checkexpire": "1",
			"checkfree":   "1",
			"app_id":      appId,
			"jsToken":     c.jsToken,
		}).
		Get("/api/quota"))
	if err != nil {
		return nil, fmt.Errorf("Error getting quota: %v", err)
	}

	var quota Quota
	json.Unmarshal(body, &quota)

	return &quota, nil
}

func (c *Client) IsVip() (bool, error) {
	body, err := util.GetResponse(c.client.R().Get("/rest/2.0/membership/proxy/user?method=query"))
	if err != nil {
		return false, err
	}
	isVip := gjson.GetBytes(body, "data.member_info.is_vip")
	return isVip.Bool(), nil
}

func (c *Client) GetDisplayName() (string, error) {
	body, err := util.GetResponse(c.client.R().
		SetQueryParams(map[string]string{
			"app_id":  appId,
			"jsToken": c.jsToken,
		}).
		Get("/passport/get_info"))
	if err != nil {
		return "", fmt.Errorf("Failed to get display name: %w", err)
	}

	displayName := gjson.GetBytes(body, "data.display_name")
	if !displayName.Exists() {
		return "", fmt.Errorf("Failed to get display name. The Cookies may expired?")
	}
	return displayName.String(), nil
}
