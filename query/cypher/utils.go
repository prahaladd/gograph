package cypher

import (
	"bytes"
	"fmt"
)

func buildSelector(selector map[string]interface{}) string {
	if len(selector) == 0 {
		return ""
	}
	buffer := bytes.Buffer{}
	buffer.WriteString("{")
	firstFilterProcessed := false
	for k, v := range selector {
		if firstFilterProcessed {
			buffer.WriteString(",")
		}
		switch v.(type) {
		case string:
			buffer.WriteString(fmt.Sprintf("%s:'%s'", k, v))
		default:
			buffer.WriteString(fmt.Sprintf("%s: %v", k, v))
		}
		firstFilterProcessed = true
	}
	buffer.WriteString("}")
	return buffer.String()
}

func buildFilterConditions(varName string, filters map[string]interface{}) string {
	if len(filters) == 0 {
		return ""
	}
	buffer := bytes.Buffer{}
	firstFilterProcessed := false
	for k, v := range filters {
		if firstFilterProcessed {
			buffer.WriteString(" AND ")
		}
		switch v.(type) {
		case string:
			buffer.WriteString(fmt.Sprintf("%s.%s='%s'", varName, k, v))
		default:
			buffer.WriteString(fmt.Sprintf("%s.%s=%v", varName, k, v))
		}
		firstFilterProcessed = true
	}
	return buffer.String()
}

func buildMultiFilters(multiFilters map[string]map[string]interface{}) string {
	if len(multiFilters) == 0 {
		return ""
	}

	buffer := bytes.Buffer{}
	
	firstFilterProcessed := false
	for k, v := range multiFilters {
		if len(v) == 0 {
			continue
		}
		if !firstFilterProcessed {
			buffer.WriteString(" WHERE ")
		}

		if firstFilterProcessed {
			buffer.WriteString(" AND ")
		}
		
		buffer.WriteString(buildFilterConditions(k, v))
		if !firstFilterProcessed {
			firstFilterProcessed = true
		}
	}
	return buffer.String()
}
