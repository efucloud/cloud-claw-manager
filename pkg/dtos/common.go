package dtos

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/efucloud/common"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type ApplicationInfo struct {
	Application string `json:"application"`
	GoVersion   string `json:"goVersion"`
	Commit      string `json:"commit"`
	BuildDate   string `json:"buildDate"`
	OS          string `json:"os,omitempty"`
	Arch        string `json:"arch,omitempty"`
	CpuCores    int    `json:"cpuCores,omitempty"`
}
type ClusterAdmin struct {
	AccountId string `gorm:"column:account_id" json:"accountId" validate:"required" description:"账户ID"`
}

func (md *ClusterAdmin) Validate(ctx context.Context) (err error) {
	validate := validator.New()
	lang := common.GetLangFromCtx(ctx, "")
	validate.RegisterTagNameFunc(common.TagNameI18N(lang))
	trans := common.LoadValidateTranslator(lang, validate)
	err = validate.Struct(md)
	if err != nil {
		var lines []string
		var errs validator.ValidationErrors
		errors.As(err, &errs)
		for _, v := range errs.Translate(trans) {
			lines = append(lines, v)
		}
		if len(lines) > 0 {
			err = errors.New(strings.Join(lines, "\n"))
		}
	}
	return
}

type QueryParam struct {
	Cluster      string        `json:"-" yaml:"-" description:"指标中集群编码"`
	Time         int64         `json:"time" yaml:"time" description:"时间戳秒,query使用"`
	View         string        `json:"view" yaml:"view" validate:"oneof=apiserver" enum:"apiserver" description:"视角"`
	Code         string        `json:"code" yaml:"code" validate:"required" description:"视角中的编码"`
	Namespace    string        `json:"namespace" yaml:"namespace" description:"命名空间"`
	Node         string        `json:"node" yaml:"node" description:"节点名"`
	Pod          string        `json:"pod" yaml:"pod" description:"Pod名"`
	Workload     string        `json:"workload" yaml:"workload" description:"工作负载"`
	WorkloadType string        `json:"workloadType" yaml:"workloadType" description:"工作负载类型"`
	Start        int64         `json:"start" yaml:"start" description:"开始时间,query_range使用"`
	End          int64         `json:"end" yaml:"end" description:"结束时间,query_range使用"`
	Step         time.Duration `json:"step" yaml:"step" description:"步长,query_range使用"`
}

type PrometheusQuery struct {
	I18N        I18N   `json:"i18n,omitempty" yaml:"i18n,omitempty"`
	Query       string `json:"query" yaml:"query" description:"promql query"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type I18N struct {
	ZH string `json:"zh" yaml:"zh" description:"中文"`
	EN string `json:"en" yaml:"en" description:"英文"`
}
type ClusterVersionInfo struct {
	Major        string `json:"major"`
	Minor        string `json:"minor"`
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// PatchSubsetValue 差异更新
type PatchSubsetValue struct {
	//操作, 例如: add、remove、replace、
	Op string `json:"op" yaml:"op" validate:"required" description:"操作, 例如: add、remove、replace"`
	//到什么路径，列如下面的: /spec/subset 从 / 下开始这里的位置就是根据具体的配置
	Path string `json:"path" yaml:"path" validate:"required" description:"路径，如下面的: /spec/subset 从 / 下开始这里的位置就是根据具体的配置"`
	//路径所指向的结构体
	Value interface{} `json:"value" yaml:"value" validate:"required" description:"值:路径所指向的结构体"`
}

// TableListPagination 分页信息
type TableListPagination struct {
	//总数
	Total uint `json:"total" description:"总数"`
	//每页数量
	PageSize uint `json:"pageSize" description:"每页数量"`
	//当前页
	Current uint `json:"current" description:"当前页"`
	//名称
	Name string `json:"name" description:"名称"`
	//编码
	Code string `json:"code" description:"编码"`
	//搜索
	Search string `json:"search" description:"搜索"`
}

// ArrayString 字符串数组
type ArrayString []string

// GormDataType gorm common data type
func (m ArrayString) GormDataType() string {
	return "jsonmap"
}

// GormDBDataType gorm db data type
func (ArrayString) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "json"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	}
	return ""
}

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (m *ArrayString) Scan(value interface{}) error {
	byteValue, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal ArrayString value: ", value))
	}
	err := json.Unmarshal(byteValue, m)
	return err
}

// Value 实现 driver.Valuer 接口，Value 返回 json value
func (m ArrayString) Value() (driver.Value, error) {
	re, err := json.Marshal(m)
	return re, err
}

// ArrayUint 整数数组
type ArrayUint []uint

// GormDataType gorm common data type
func (u ArrayUint) GormDataType() string {
	return "jsonmap"
}

// GormDBDataType gorm db data type
func (ArrayUint) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "json"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	}
	return ""
}

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (u *ArrayUint) Scan(value interface{}) error {
	byteValue, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal ArrayUint value: ", value))
	}
	err := json.Unmarshal(byteValue, u)
	return err
}

// Value 实现 driver.Valuer 接口，Value 返回 json value
func (u ArrayUint) Value() (driver.Value, error) {
	re, err := json.Marshal(u)
	return re, err
}

// SystemUserJoinOrganization 用户加入组织
type SystemUserJoinOrganization struct {
	//用户ID
	UserId string `json:"userId" validate:"required" description:"用户ID"`
	//组织角色
	Role string `gorm:"type:varchar(255);column:role;default:none" json:"role" validate:"oneof=admin view edit none" enum:"admin|view|edit|none" description:"组织角色"`
	//组织ID
	OrganizationId string `json:"organizationId" validate:"required" description:"组织ID"`
	//用户名
	Username string `json:"username" description:"用户名"`
}

func (t *SystemUserJoinOrganization) Default(ctx context.Context) {
}
func (t *SystemUserJoinOrganization) Validate(ctx context.Context) (err error) {
	validate := validator.New()
	lang := common.GetLangFromCtx(ctx, "")
	validate.RegisterTagNameFunc(common.TagNameI18N(lang))
	trans := common.LoadValidateTranslator(lang, validate)
	err = validate.Struct(t)
	if err != nil {
		var lines []string
		var errs validator.ValidationErrors
		errors.As(err, &errs)
		for _, v := range errs.Translate(trans) {
			lines = append(lines, v)
		}
		if len(lines) > 0 {
			err = errors.New(strings.Join(lines, "\n"))
		}
	}
	return
}

// RequestParameter 请求参数
// 在部署时用户提交的参数，提交的是某个操作模版中使用ParameterDefinition定义的参数，
type RequestParameter struct {
	//变量名
	Key string `json:"key" description:"变量名"`
	//变量值
	Val string `json:"val" description:"变量值"`
}

// EsQuery ES查请求封装结构
type EsQuery struct {
	//查询请求数据
	Query EsQueryBody `json:"query" description:"查询请求数据"`
	//查询记录偏移
	From int `json:"from,omitempty" description:"查询记录偏移"`
	//查询数据数量
	Size int `json:"size,omitempty" description:"查询数据数量"`
	//查询排序
	Sort []interface{} `json:"sort,omitempty" description:"查询排序"`
}

// EsQueryBody 查询请求
type EsQueryBody struct {
	Bool struct {
		//必须匹配内容
		Must []interface{} `json:"must,omitempty" description:""`
		//过滤器
		Filter *Filter `json:"filter,omitempty" description:""`
	} `json:"bool,omitempty" description:""`
}

// QueryMust 匹配规则
type QueryMust struct {
	MatchPhrase map[string]interface{} `json:"match_phrase,omitempty" description:""`
	Match       map[string]interface{} `json:"match,omitempty" description:""`
}
type QueryWildcard struct {
	Wildcard map[string]interface{} `json:"wildcard" yaml:"wildcard" description:""`
}
type QueryExist struct {
	Exists map[string]interface{} `json:"exists" yaml:"exists" description:""`
}
type Filter struct {
	Range struct {
		RequestReceivedTimestamp struct {
			Gte string `json:"gte,omitempty" description:""`
			Lte string `json:"lte,omitempty" description:""`
		} `json:"requestReceivedTimestamp,omitempty" description:""`
	} `json:"range,omitempty" description:""`
	//正则匹配
	Regexp map[string]interface{} `json:"regexp,omitempty" description:""`
}

// ResponseError 错误响应
type ResponseError struct {
	//错误英文编码
	Message string `json:"message" yaml:"message" description:"错误英文编码"`
	//错误详情信息
	Detail string `json:"detail" yaml:"detail" description:"错误详情信息"`
	//支持I18N的提示信息
	Alert string `json:"alert" yaml:"alert" description:"支持I18N的提示信息"`
	//当前请求地址
	RequestURI string `json:"requestUri" description:"当前请求地址"`
}

// AccessTokenResponse 登录获取token
type AccessTokenResponse struct {
	AccessToken  string `json:"access_token,omitempty" validate:"required"`
	TokenType    string `json:"token_type,omitempty" validate:"required"`
	ExpiresIn    int64  `json:"expires_in,omitempty" validate:"required"`
	RefreshToken string `json:"refresh_token,omitempty" validate:"required"`
	IDToken      string `json:"id_token,omitempty" validate:"required"`
	Timestamp    int64  `json:"timestamp,omitempty" validate:"required" description:"token有效时间"`
}

// ClientInformation 获取浏览器等客户端信息
type ClientInformation struct {
	//平台 如"MacIntel"、"Win32"、"Linux x86_64"、"Linux armv81"
	Platform string `json:"platform"  description:"平台"`
	//供应商
	Vendor string `json:"vendor" description:"供应商"`
	//客户端代理
	UserAgent string `json:"userAgent" description:"客户端代理"`
	//CPU核数
	HardwareConcurrency int `json:"hardwareConcurrency" description:"CPU核数"`
	//浏览器
	Browser string `json:"browser" description:"浏览器"`
	//客户端地址
	Remote string `json:"remote" description:"客户端地址"`
	//平台细节
	PlatformDetail string `json:"platformDetail" description:"平台细节"`
	//内存大小(G)
	DeviceMemory uint `json:"deviceMemory" description:"内存大小(G)"`
}

func (ClientInformation) GormDataType() string {
	return "json"
}

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (ins *ClientInformation) Scan(value interface{}) error {
	byteValue, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal ProviderExtend value: ", value))
	}
	err := json.Unmarshal(byteValue, ins)
	return err
}

// Value 实现 driver.Valuer 接口，Value 返回 json value
func (ins ClientInformation) Value() (driver.Value, error) {
	re, err := json.Marshal(ins)
	return re, err
}

// BatchOperationIds 需要删除的列表,根据数据库id
type BatchOperationIds struct {
	//需要删key 可以为数据库的id
	Ids []string `json:"ids" validate:"required" description:"需要批量操作的数据库ID"`
}

// BatchOperationKeys 需要删除的列表，根据表唯一字符型字段
type BatchOperationKeys struct {
	//需要删key 可以为数据库的id
	Ids []string `json:"ids" validate:"required"  description:"需要批量操作的数据库ID"`
}

// AnyJsonData 任意json数据
type AnyJsonData map[string]interface{}
type WorkspaceBatchOperationIds struct {
	WorkspaceId string   `json:"workspaceId" description:"工作空间"`
	Ids         []string `json:"ids" validate:"required"  description:"需要批量操作的数据库ID"`
}

// AuthedUserInfo 用户信息
type AuthedUserInfo struct {
	//主键
	ID string `json:"id" validate:"required" description:"记录ID"`
	//用户名
	Username string `json:"username" validate:"required,max=255" description:"用户名"`
	//昵称，如中文名
	Nickname string `json:"nickname" validate:"max=255" description:"昵称"`
	//工号
	JobNumber string `json:"jobNumber" validate:"max=255" description:"工号"`
	//系统角色
	Role string `json:"role" validate:"oneof=admin view edit none" enum:"admin|view|edit|none" description:"系统角色"`
	//是否有效
	Enable bool `json:"enable" description:"是否有效"`
	//邮箱
	Email string `json:"email" validate:"max=255" description:"邮箱"`
	//手机号码
	Phone string `json:"phone" validate:"max=50" description:"电话"`
	//默认语言
	Language string `json:"language" validate:"oneof=zh en" enum:"zh|en" description:"默认语言"`
	//头像
	Avatar string `json:"avatar" validate:"required" description:"头像"`
	//远程地址
	RemoteAddress string `json:"remoteAddress"  validate:"required" description:"远程地址"`
}

type OidcConfig struct {
	Issuer   string `json:"issuer" validate:"required" description:"发行者"`
	ClientId string `json:"clientId" validate:"required" description:"客户端ID"`
}

// LoginByOIDC OIDC登录
type LoginByOIDC struct {
	Code        string `json:"code" description:"OIDC提供商返回的Code"`
	RedirectUri string `json:"redirectUri" description:"重定向地址"`
	//客户端信息 加密数据
	Client string `json:"client" description:"客户端信息"`
}

// OidcRequestToken oidc认证获取token请求结构
type OidcRequestToken struct {
	//客户端ID
	ClientId string `schema:"client_id" json:"client_id" validate:"required"`
	//客户端密钥
	ClientSecret string `schema:"client_secret" json:"client_secret" validate:"required"`
	//类型
	GrantType string `schema:"grant_type" json:"grant_type" validate:"required"`
	//请求码
	Code string `schema:"code" json:"code" validate:"required" `
	//重定向地址
	RedirectUri string `schema:"redirect_uri" json:"redirect_uri"`
	//客户端信息 加密数据
	Client string `json:"client" description:"客户端信息"`
}

type UserClaims struct {
	//组织用户ID
	ID      string `json:"id" description:"用户ID"`
	EAuthId string `json:"eAuthId" description:""`
	// 用户名 组织内唯一必须由DNS-1123标签格式的单元组成
	Username string `json:"username"`
	// 昵称，如中文名
	Nickname string `json:"nickname"`
	// 系统角色
	Role  string `json:"role"`
	Nonce string `json:"nonce"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	jwt.RegisteredClaims
}
type WebsocketUserInfo struct {
	UserId    string `json:"userId"`
	TimeStamp string `json:"timeStamp"`
}

// UserAgentInformation 获取浏览器等客户端信息
type UserAgentInformation struct {
	//平台 如"MacIntel"、"Win32"、"Linux x86_64"、"Linux armv81"
	Platform string `json:"platform"  description:"平台"`
	//供应商
	Vendor string `json:"vendor" description:"供应商"`
	//客户端代理
	UserAgent string `json:"userAgent" description:"客户端代理"`
	//CPU核数
	HardwareConcurrency int `json:"hardwareConcurrency" description:"CPU核数"`
	//浏览器
	Browser string `json:"browser" description:"浏览器"`
	//客户端地址
	Remote string `json:"remote" description:"客户端地址"`
	//平台细节
	PlatformDetail string `json:"platformDetail" description:"平台细节"`
	//内存大小(G)
	DeviceMemory uint `json:"deviceMemory" description:"内存大小(G)"`
}

func (UserAgentInformation) GormDataType() string {
	return "json"
}

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (ins *UserAgentInformation) Scan(value interface{}) error {
	byteValue, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal ProviderExtend value: ", value))
	}
	err := json.Unmarshal(byteValue, ins)
	return err
}

// Value 实现 driver.Valuer 接口，Value 返回 json value
func (ins UserAgentInformation) Value() (driver.Value, error) {
	re, err := json.Marshal(ins)
	return re, err
}

// Status 修改记录状态
type Status struct {
	//主键
	Ids []string `json:"ids" validate:"required" description:"主键"`
	//更新时间
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at;<-:update" json:"-" description:"更新时间"`
	//更新者
	UpdaterId string `gorm:"type:varchar(50);column:updater_id;<-:update" json:"-" validate:"required" description:"更新者"`

	//是否有效
	Enable bool `gorm:"column:enable;default:true" json:"enable" description:"是否有效"`
}

func (t *Status) Default(ctx context.Context) {
	t.UpdatedAt = time.Now()
}

// ByIds 根据id列表获取信息
type ByIds struct {
	OrganizationId string   `gorm:"column:organization_id" json:"-"  description:"组织ID"`
	Ids            []string `json:"ids" description:"ID列表"`
}

type DutyTime struct {
	StartTime time.Time `json:"startTime" description:"开始时间"`
	EndTime   time.Time `json:"endTime" description:"结束时间"`
}

// ClusterExtendConfig 集群扩展信息
type ClusterExtendConfig struct {
	//审计日志服务配置
	//AuditLogConfig *AuditLogConfig `json:"auditLogConfig" description:"审计日志服务配置"`
	//prometheus配置
	MonitorPrometheusConfig *MonitorPrometheusConfig `json:"monitorPrometheusConfig" description:"prometheus配置"`
}

// GormDataType gorm common data type
func (m ClusterExtendConfig) GormDataType() string {
	return "jsonmap"
}

// GormDBDataType gorm db data type
func (ClusterExtendConfig) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "json"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	}
	return ""
}

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (m *ClusterExtendConfig) Scan(value interface{}) error {
	byteValue, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal ArrayString value: ", value))
	}
	err := json.Unmarshal(byteValue, m)
	return err
}

// Value 实现 driver.Valuer 接口，Value 返回 json value
func (m ClusterExtendConfig) Value() (driver.Value, error) {
	re, err := json.Marshal(m)
	return re, err
}

// MonitorPrometheusConfig 监控配置
// 配置集群或者应用系统基于prometheus的监控配置信息
type MonitorPrometheusConfig struct {
	//Prometheus指标中集群编码 多集群数据共存时使用
	Cluster string `json:"cluster" yaml:"cluster" description:"指标中集群编码"`
	//prometheus地址
	Address string `json:"address" yaml:"address" description:"prometheus地址"`
	////用户访问Token
	//UserToken bool `json:"userToken" yaml:"userToken" description:"用户访问Token"`
	////Token
	//Token string `json:"token" yaml:"token" description:"Token"`
	////用户名
	//Username string `json:"username" yaml:"username" description:"用户名"`
	////密码
	//Password string `json:"password" yaml:"password" description:"密码"`
}

// AuditLogConfig 审计日志配置
type AuditLogConfig struct {
	Version   string
	Username  string
	Password  string
	Addresses []string
	//服务地址
	Address string `json:"address" yaml:"address" description:"服务地址"`
	//API Key
	ApiKey string `json:"apiKey" yaml:"apiKey" description:"API Key"`
	//API Secret
	ApiSecret string `json:"apiSecret" yaml:"apiSecret" description:"API Secret"`
}

// S3StorageConfig 对象存储配置
type S3StorageConfig struct {
	//地址
	Endpoint string `json:"endpoint" description:"地址"`
	//accessKeyId
	AccessKeyID string `json:"accessKeyId" description:"accessKeyId"`
	//secretAccessKey
	SecretAccessKey string `json:"secretAccessKey" description:"secretAccessKey"`
	//token
	Token string `json:"token" description:"token"`
	//insecure
	Insecure bool `json:"insecure" description:"insecure"`
	//region
	Region string `json:"region" description:"区域"`
}

// AuditLogSearchOption 审计日志es搜索
type AuditLogSearchOption struct {
	//集群编码
	Cluster string `json:"cluster" description:"集群编码"`
	//Namespace,带*模糊搜索
	Namespace string `json:"namespace" description:"Namespace,带*模糊搜索"`
	//用户名，包括sa账户,带*模糊搜索
	Username string `json:"username" description:"用户名，包括sa账户,带*模糊搜索"`
	//资源对象,带*模糊搜索
	Resource string `json:"resource" description:"资源对象,带*模糊搜索"`
	//资源对象名称,带*模糊搜索
	ResourceName string `json:"resourceName" description:"资源对象名称,带*模糊搜索"`
	//鉴权结果
	Decision string `json:"decision" enum:"forbid|allow" description:"鉴权结果"`
	//操作
	Verb string `json:"verb"  enum:"create|delete|deletecollection|get|list|patch|update|watch" description:"操作"`
	//从第几页开始
	From int `json:"page" description:"从第几页开始"`
	//页码大小
	Size int `json:"size" description:"页码大小"`
	//字段排序，只是上面的字段，sort: {'username':'desc or asc'}
	Sort map[string]string `json:"sort" description:"字段排序"`
	//开始时间 2023-07-23
	StartTime string `json:"startTime" description:"开始时间"`
	//结束时间，2023-07-23
	EndTime string `json:"endTime" description:"结束时间"`
}

// I18NInfo 国际化
type I18NInfo struct {
	ZH string `json:"zh" description:"中文"`
	EN string `json:"en" description:"英文"`
}

// GormDataType gorm common data type
func (m I18NInfo) GormDataType() string {
	return "jsonmap"
}

// GormDBDataType gorm db data type
func (I18NInfo) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "json"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	}
	return ""
}

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (m *I18NInfo) Scan(value interface{}) error {
	byteValue, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal I18NInfo value: ", value))
	}
	err := json.Unmarshal(byteValue, m)
	return err
}

// Value 实现 driver.Valuer 接口，Value 返回 json value
func (m I18NInfo) Value() (driver.Value, error) {
	re, err := json.Marshal(m)
	return re, err
}
