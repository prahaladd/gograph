package cypher

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/prahaladd/gograph/core"
)

// VertexQueryBuilder exposes a builder pattern for building vertex lookup  cypher queries
//
// # The mode of the query builder can be used to control the type of cypher query - mutating or querying
//
// selectors on a vertex are used to select vertices/nodes with the specific set of properties
//
// filters can be specified to filter out selected nodes based on some critieria
//
// Multiple vertex labels can be specified when constructing the query builder. In case, more than one  labels are
// specified in the builder; the resultant query would contain a MATCH/MERGE clause on a node with all the
// labels applied.
type VertexQueryBuilder struct {
	queryMode core.QueryMode
	labels    []string
	varName   string
	selector  core.KVMap
	filters   core.KVMap
	writeMode core.WriteMode
}

func NewVertexQueryBuilder() *VertexQueryBuilder {
	return &VertexQueryBuilder{selector: core.KVMap{}, filters: core.KVMap{}, writeMode: core.Merge}
}

func (vqb *VertexQueryBuilder) SetQueryMode(mode core.QueryMode) *VertexQueryBuilder {
	vqb.queryMode = mode
	return vqb
}

func (vqb *VertexQueryBuilder) SetLabel(labels []string) *VertexQueryBuilder {
	vqb.labels = labels
	return vqb
}

func (vqb *VertexQueryBuilder) SetVarName(varName string) *VertexQueryBuilder {
	vqb.varName = varName
	return vqb
}

func (vqb *VertexQueryBuilder) SetSelector(selectors core.KVMap) *VertexQueryBuilder {
	// deep copy operation
	for k, v := range selectors {
		vqb.selector[k] = v
	}
	return vqb
}

func (vqb *VertexQueryBuilder) SetFilters(filters core.KVMap) *VertexQueryBuilder {
	for k, v := range filters {
		vqb.filters[k] = v
	}
	return vqb
}

func (vqb *VertexQueryBuilder) SetWriteMode(writeMode core.WriteMode) *VertexQueryBuilder {
	vqb.writeMode = writeMode
	return vqb
}

func (vqb *VertexQueryBuilder) Build() (string, error) {

	err := vqb.validate()
	if err != nil {
		return "", err
	}
	operation := "MATCH"

	if vqb.queryMode == core.Write {
		operation = "MERGE"
		if vqb.writeMode == core.Create {
			operation = "CREATE"
		}
	}

	variableName := strings.ToLower(vqb.labels[0])[0:2]
	if vqb.varName != "" {
		variableName = vqb.varName
	}
	selectors := buildSelector(vqb.selector)
	filters := buildMultiFilters(map[string]map[string]interface{}{variableName: vqb.filters})

	labelSelectors := bytes.Buffer{}
	for _, label := range vqb.labels {
		labelSelectors.WriteString(fmt.Sprintf(":%s", label))
	}
	return fmt.Sprintf("%s (%s%s%s) %s return %s", operation, variableName, labelSelectors.String(), selectors, filters, variableName), nil

}

func (vqb *VertexQueryBuilder) validate() error {
	if vqb.labels == nil || len(vqb.labels) == 0 {
		return errors.New("no vertex labels specified in the query")
	}
	return nil
}
