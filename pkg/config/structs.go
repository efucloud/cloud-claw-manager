package config

import (
	"context"
)

type Config struct {
	LogConfig       *LogConfig            `json:"logConfig" yaml:"logConfig"`
	OidcConfig      OidcConfig            `json:"oidcConfig" yaml:"oidcConfig" description:"认证配置"`
	OpenClawControl OpenClawControlConfig `json:"openClawControl" yaml:"openClawControl" description:"OpenClaw 管理控制配置"`
}

type OidcConfig struct {
	ClientId     string `json:"clientId" yaml:"clientId" description:"客户端ID"`
	ClientSecret string `json:"clientSecret" yaml:"clientSecret" description:"客户端密钥"`
	Issuer       string `json:"issuer" yaml:"issuer" description:"发行者"`
}
type LogConfig struct {
	Level string `json:"level,omitempty" yaml:"level,omitempty" description:"日志级别"`
	//Filename is the file to write logs to.  Backup log files will be retained
	//in the same directory.  It uses <processname>-lumberjack.log in
	//os.TempDir() if empty.
	Filename string `json:"filename,omitempty" yaml:"filename,omitempty"`

	//MaxSize is the maximum size in megabytes of the log file before it gets
	//rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize,omitempty" yaml:"maxsize,omitempty"`

	//MaxAge is the maximum number of days to retain old log files based on the
	//timestamp encoded in their filename.  Note that a day is defined as 24
	//hours and may not exactly correspond to calendar days due to daylight
	//savings, leap seconds, etc. The default is not to remove old log files
	//based on age.
	MaxAge int `json:"maxage,omitempty" yaml:"maxage,omitempty"`

	//MaxBackups is the maximum number of old log files to retain.  The default
	//is to retain all old log files (though MaxAge may still cause them to get
	//deleted.)
	MaxBackups int `json:"maxbackups,omitempty" yaml:"maxbackups,omitempty"`

	//LocalTime determines if the time used for formatting the timestamps in
	//backup files is the computer's local time.  The default is to use UTC
	//time.
	LocalTime bool `json:"localtime,omitempty" yaml:"localtime,omitempty"`

	//Compress determines if the rotated log files should be compressed
	//using gzip. The default is not to perform compression.
	Compress   bool  `json:"compress,omitempty" yaml:"compress,omitempty"`
	Production *bool `json:"production,omitempty" yaml:"production,omitempty"`
}

type OpenClawControlConfig struct {
	PreviewBaseDomain    string   `json:"previewBaseDomain" yaml:"previewBaseDomain" description:"OpenClaw 预览基础域名，例如 openclaw.example.com"`
	AdminEmails          []string `json:"adminEmails" yaml:"adminEmails" description:"OpenClaw 模板管理员邮箱白名单"`
	IngressEnabled       bool     `json:"ingressEnabled" yaml:"ingressEnabled" description:"是否为 OpenClaw 实例创建 Ingress"`
	IngressClassName     string   `json:"ingressClassName" yaml:"ingressClassName" description:"IngressClass 名称"`
	IngressPath          string   `json:"ingressPath" yaml:"ingressPath" description:"Ingress 路径"`
	IngressPathType      string   `json:"ingressPathType" yaml:"ingressPathType" description:"Ingress 路径类型（Exact/Prefix/ImplementationSpecific）"`
	IngressTLSEnabled    bool     `json:"ingressTlsEnabled" yaml:"ingressTlsEnabled" description:"是否启用 Ingress TLS"`
	IngressTLSSecretName string   `json:"ingressTlsSecretName" yaml:"ingressTlsSecretName" description:"Ingress TLS Secret 名称（启用 TLS 时生效）"`
}

func (c *Config) Default(ctx context.Context) {
	if c.OpenClawControl.IngressClassName == "" {
		c.OpenClawControl.IngressClassName = "nginx"
	}
	if c.OpenClawControl.IngressPath == "" {
		c.OpenClawControl.IngressPath = "/"
	}
	if c.OpenClawControl.IngressPathType == "" {
		c.OpenClawControl.IngressPathType = "Prefix"
	}
}
