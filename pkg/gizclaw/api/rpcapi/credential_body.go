package rpcapi

import (
	"encoding/json"
	"strings"
)

func CredentialBodyMap(body CredentialBody) map[string]any {
	var out map[string]any
	data, err := body.MarshalJSON()
	if err != nil || string(data) == "null" {
		return nil
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}

func CredentialBodyString(body CredentialBody, keys ...string) string {
	values := CredentialBodyMap(body)
	for _, key := range keys {
		if value, ok := values[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func NewOpenAICredentialBody(apiKey string, baseURL ...string) CredentialBody {
	body := CredentialBody{}
	value := OpenAICredentialBody{}
	setStringPtr(&value.ApiKey, apiKey)
	if len(baseURL) > 0 && strings.TrimSpace(baseURL[0]) != "" {
		text := strings.TrimSpace(baseURL[0])
		value.BaseUrl = &text
	}
	_ = body.FromOpenAICredentialBody(value)
	return body
}

func setStringPtr(dst **string, value string) {
	if text := strings.TrimSpace(value); text != "" {
		*dst = &text
	}
}
