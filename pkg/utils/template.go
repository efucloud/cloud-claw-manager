package utils

import (
	"bytes"
	"fmt"
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/common"
	"html/template"
	"strings"
)

func TemplateRender(name, content string, params interface{}) (result string, errorData common.ErrorData) {
	var tmpl *template.Template
	tmpl, errorData.Err = template.New(name).Delims("_{{_", "_}}_").Parse(content)
	if errorData.IsNotNil() {
		config.Logger.Error(errorData.Err.Error())
		return
	}
	if tmpl == nil {
		errorData.Err = fmt.Errorf("template is nil")
		return
	}
	b := new(bytes.Buffer)
	errorData.Err = tmpl.Execute(b, params)
	if errorData.IsNil() {
		var lines []string
		for _, line := range strings.Split(b.String(), "\n") {
			// 剔除空白行（包括只含空格、tab 的行）
			if strings.TrimSpace(line) != "" {
				lines = append(lines, line)
			}
		}
		result = strings.Join(lines, "\n")
	}
	return
}
func SplitKubernetesResources(yamlContent string) (results []string) {
	lines := strings.Split(yamlContent, "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.TrimSpace(line) == "---" {
			results = append(results, strings.Join(result, "\n"))
			result = []string{}
		} else {
			result = append(result, line)
		}
	}
	if len(result) > 0 {
		results = append(results, strings.Join(result, "\n"))
	}
	return
}
func TrimEmptyLinetKubernetesResources(yamlContent string) (result string) {
	lines := strings.Split(yamlContent, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.TrimSpace(line) == "---" {
			continue
		}
		result += fmt.Sprintf("%s\n", line)
	}
	return
}

type TypeMeta struct {
	Kind       string `json:"kind" yaml:"kind" description:""`
	APIVersion string `json:"apiVersion" yaml:"apiVersion" description:""`
}

func GetResourceGroupVersion(content string) (result TypeMeta) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "apiVersion: ") {
			result.APIVersion = strings.TrimPrefix(line, "apiVersion: ")
		} else if strings.HasPrefix(strings.TrimSpace(line), "kind: ") {
			result.Kind = strings.TrimPrefix(line, "kind: ")
		}
		if len(result.Kind) > 0 && len(result.APIVersion) > 0 {
			break
		}
	}
	return
}
