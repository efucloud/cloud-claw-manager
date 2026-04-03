package filters

import (
	"strings"

	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/common"
	"github.com/emicklei/go-restful/v3"
)

// I18n default language is zh
func I18n(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	locale := firstNonEmpty(
		req.HeaderParameter("X-Locale"),
		req.HeaderParameter("Current-Language"),
		req.Request.FormValue("lang"),
		req.HeaderParameter("Accept-Language"),
	)
	req.SetAttribute(config.RequestLanguage, normalizeI18nLanguage(locale))
	chain.ProcessFilter(req, resp)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func normalizeI18nLanguage(language string) string {
	raw := strings.TrimSpace(strings.ToLower(language))
	if raw == "" {
		return common.I18nZH
	}

	if index := strings.Index(raw, ","); index >= 0 {
		raw = strings.TrimSpace(raw[:index])
	}
	if index := strings.Index(raw, ";"); index >= 0 {
		raw = strings.TrimSpace(raw[:index])
	}

	switch {
	case raw == "en" || raw == "english" || strings.HasPrefix(raw, "en-") || strings.HasPrefix(raw, "en_"):
		return common.I18nEN
	case raw == "zh" || raw == "chinese" || raw == "中文" || strings.HasPrefix(raw, "zh-") || strings.HasPrefix(raw, "zh_"):
		return common.I18nZH
	default:
		return common.I18nZH
	}
}
