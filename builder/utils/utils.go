package utils

import (
	"fmt"
	"regexp"
)

// GetStringMap converts map[string]interface{} → map[string]string
func GetStringMap(raw interface{}) map[string]string {
	if raw == nil {
		return make(map[string]string)
	}
	in := raw.(map[string]interface{})
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

// GetStringSlice converts []interface{} → []string
func GetStringSlice(raw interface{}) []string {
	if raw == nil {
		return []string{}
	}
	in := raw.([]interface{})
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = fmt.Sprintf("%v", v)
	}
	return out
}

// PrefixList adds a prefix to each string in a slice
func PrefixList(prefix string, list []string) []string {
	out := make([]string, len(list))
	for i, v := range list {
		out[i] = prefix + v
	}
	return out
}

// FilterExisting returns only items that exist as keys in the map
func FilterExisting[T any](list []string, m map[string]T) []string {
	var result []string
	for _, item := range list {
		if _, ok := m[item]; ok {
			result = append(result, item)
		}
	}
	return result
}

// ResolveVars substitutes ${var} using the given map
func ResolveVars(cmd map[string]string, vars map[string]string, stageName string) map[string]string {
	varPattern := regexp.MustCompile(`\$\{(\w+)\}`)
	result := make(map[string]string)
	for k, v := range cmd {
		newVal := varPattern.ReplaceAllStringFunc(v, func(m string) string {
			match := varPattern.FindStringSubmatch(m)
			if val, ok := vars[match[1]]; ok {
				return val
			}
			panic(fmt.Sprintf("Missing variable '%s' in '%s'", match[1], stageName))
		})
		result[k] = newVal
	}
	return result
}

func CopyStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
