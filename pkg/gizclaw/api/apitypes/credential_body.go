package apitypes

import (
	"encoding/json"
	"strings"
)

func IsZeroCredentialBody(body CredentialBody) bool {
	return len(body.union) == 0 || string(body.union) == "null"
}

func CloneCredentialBody(body CredentialBody) CredentialBody {
	if IsZeroCredentialBody(body) {
		return CredentialBody{}
	}
	return CredentialBody{union: append([]byte(nil), body.union...)}
}

func CredentialBodyMap(body CredentialBody) map[string]any {
	if IsZeroCredentialBody(body) {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(body.union, &out); err != nil {
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

func NewGeminiCredentialBody(apiKey string, baseURL ...string) CredentialBody {
	body := CredentialBody{}
	value := GeminiCredentialBody{}
	setStringPtr(&value.ApiKey, apiKey)
	if len(baseURL) > 0 && strings.TrimSpace(baseURL[0]) != "" {
		text := strings.TrimSpace(baseURL[0])
		value.BaseUrl = &text
	}
	_ = body.FromGeminiCredentialBody(value)
	return body
}

func NewDashScopeCredentialBody(apiKey string, baseURL ...string) CredentialBody {
	body := CredentialBody{}
	value := DashScopeCredentialBody{}
	setStringPtr(&value.ApiKey, apiKey)
	if len(baseURL) > 0 && strings.TrimSpace(baseURL[0]) != "" {
		text := strings.TrimSpace(baseURL[0])
		value.BaseUrl = &text
	}
	_ = body.FromDashScopeCredentialBody(value)
	return body
}

func NewMiniMaxCredentialBody(apiKey string) CredentialBody {
	body := CredentialBody{}
	value := MiniMaxCredentialBody{}
	setStringPtr(&value.ApiKey, apiKey)
	_ = body.FromMiniMaxCredentialBody(value)
	return body
}

func NewMiniMaxCredentialBodyFromStrings(values map[string]string) CredentialBody {
	value := MiniMaxCredentialBody{}
	set := func(dst **string, key string) {
		if text := strings.TrimSpace(values[key]); text != "" {
			*dst = &text
		}
	}
	set(&value.ApiKey, "api_key")
	set(&value.BaseUrl, "base_url")
	set(&value.MinimaxVoiceBaseUrl, "minimax_voice_base_url")
	set(&value.Token, "token")
	set(&value.VoiceBaseUrl, "voice_base_url")
	body := CredentialBody{}
	_ = body.FromMiniMaxCredentialBody(value)
	return body
}

func NewVolcCredentialBody(value VolcCredentialBody) CredentialBody {
	body := CredentialBody{}
	_ = body.FromVolcCredentialBody(value)
	return body
}

func NewVolcCredentialBodyFromStrings(values map[string]string) CredentialBody {
	value := VolcCredentialBody{}
	set := func(dst **string, key string) {
		if text := strings.TrimSpace(values[key]); text != "" {
			*dst = &text
		}
	}
	set(&value.AccessKey, "access_key")
	set(&value.AccessKeyId, "access_key_id")
	set(&value.AccessToken, "access_token")
	set(&value.Ak, "ak")
	set(&value.ApiKey, "api_key")
	set(&value.AppId, "app_id")
	set(&value.BaseUrl, "base_url")
	set(&value.BearerToken, "bearer_token")
	set(&value.SaucAccessKey, "sauc_access_key")
	set(&value.SecretAccessKey, "secret_access_key")
	set(&value.SecretKey, "secret_key")
	set(&value.SessionToken, "session_token")
	set(&value.Sk, "sk")
	set(&value.Token, "token")
	set(&value.VoiceBaseUrl, "voice_base_url")
	set(&value.XApiKey, "x_api_key")
	return NewVolcCredentialBody(value)
}

func setStringPtr(dst **string, value string) {
	if text := strings.TrimSpace(value); text != "" {
		*dst = &text
	}
}
