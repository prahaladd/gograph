package cypher

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/prahaladd/gograph/core"
)

// EdgeQueryBuilder exposes a builder pattern for building edge lookup  cypher queries
//
// # The mode of the query builder can be used to control the type of cypher query - mutating or querying
//
// The fetch mode can be used to control whether the returned edge information contains only ids of the start and end vertices or the complete
// representations of start and end vertices
//
// selectors on a start and end vertices are used to select vertices/nodes and edges with the specific set of properties
//
// filters can be specified to filter out the selected edges  based on some critieria
//
// Multiple vertex labels can be specified when constructing the query builder. In case, more than one  labels are
// specified in the builder; the resultant query would contain a MATCH/MERGE clause on a node with all the
// labels applied.
type EdgeQueryBuilder struct {
	queryMode           core.QueryMode
	edgeFetchMode       core.EdgeFetchMode
	startVertexLabels   []string
	startVertexVarName  string
	endVertexLabels     []string
	endVertexVarName    string
	labels              []string
	varName             string
	startVertexSelector core.KVMap
	endVertexSelector   core.KVMap
	selector            core.KVMap
	filters             core.KVMap
	startVertexFilters  core.KVMap
	endVertexFilters    core.KVMap
}

func NewEdgeQueryBuilder() *EdgeQueryBuilder {
	return &EdgeQueryBuilder{selector: make(core.KVMap),
		filters:             make(core.KVMap),
		startVertexSelector: make(core.KVMap),
		endVertexSelector:   make(core.KVMap),
		startVertexFilters:  make(core.KVMap),
		endVertexFilters:    make(core.KVMap),
	}
}

func (eqb *EdgeQueryBuilder) SetQueryMode(mode core.QueryMode) *EdgeQueryBuilder {
	eqb.queryMode = mode
	return eqb
}

func (eqb *EdgeQueryBuilder) SetEdgeFetchMode(edgeFetchMode core.EdgeFetchMode) *EdgeQueryBuilder {
	eqb.edgeFetchMode = edgeFetchMode
	return eqb
}

func (eqb *EdgeQueryBuilder) SetStartVertexLabels(labels []string) *EdgeQueryBuilder {
	eqb.startVertexLabels = append(eqb.startVertexLabels, labels...)
	return eqb
}

func (eqb *EdgeQueryBuilder) SetStartVertexVariableName(varName string) *EdgeQueryBuilder {
	eqb.startVertexVarName = varName
	return eqb
}

func (eqb *EdgeQueryBuilder) SetEndVertexLabels(labels []string) *EdgeQueryBuilder {
	eqb.endVertexLabels = append(eqb.endVertexLabels, labels...)
	return eqb
}

func (eqb *EdgeQueryBuilder) SetEndVertexVariableName(varName string) *EdgeQueryBuilder {
	eqb.endVertexVarName = varName
	return eqb
}

func (eqb *EdgeQueryBuilder) SetLabel(labels []string) *EdgeQueryBuilder {
	eqb.labels = append(eqb.labels, labels...)
	return eqb
}

func (eqb *EdgeQueryBuilder) SetVariableName(varName string) *EdgeQueryBuilder {
	eqb.varName = varName
	return eqb
}

func (eqb *EdgeQueryBuilder) SetStartVertexSelector(selector core.KVMap) *EdgeQueryBuilder {
	for k, v := range selector {
		eqb.startVertexSelector[k] = v
	}
	return eqb
}

func (eqb *EdgeQueryBuilder) SetEndVertexSelector(selector core.KVMap) *EdgeQueryBuilder {
	for k, v := range selector {
		eqb.endVertexSelector[k] = v
	}
	return eqb
}

func (eqb *EdgeQueryBuilder) SetSelector(selectors core.KVMap) *EdgeQueryBuilder {
	// deep copy operation
	for k, v := range selectors {
		eqb.selector[k] = v
	}
	return eqb
}

func (eqb *EdgeQueryBuilder) SetFilters(filters core.KVMap) *EdgeQueryBuilder {
	for k, v := range filters {
		eqb.filters[k] = v
	}
	return eqb
}

func (eqb *EdgeQueryBuilder) SetStartVertexFilters(filters core.KVMap) *EdgeQueryBuilder {
	for k, v := range filters {
		eqb.startVertexFilters[k] = v
	}
	return eqb
}

func (eqb *EdgeQueryBuilder) SetEndVertexFilters(filters core.KVMap) *EdgeQueryBuilder {
	for k, v := range filters {
		eqb.endVertexFilters[k] = v
	}
	return eqb
}

func (eqb *EdgeQueryBuilder) Build() (string, error) {

	err := eqb.validate()
	if err != nil {
		return "", err
	}
	operation := "MATCH"
	if eqb.queryMode == core.Write {
		operation = "MERGE"
	}

	startVertexQueryFragment, startVertexVarName := eqb.buildVertexQueryFragment(eqb.startVertexVarName, eqb.startVertexLabels, eqb.startVertexSelector)
	endVertexQueryFragment, endVertexVarName := eqb.buildVertexQueryFragment(eqb.endVertexVarName, eqb.endVertexLabels, eqb.endVertexSelector)
	edgeQueryFragment, edgeVarName := eqb.buildEdgeQueryFragment()

	allFilters := map[string]map[string]interface{}{startVertexVarName: eqb.startVertexFilters, endVertexVarName: eqb.endVertexFilters, edgeVarName: eqb.filters}

	filters := buildMultiFilters(allFilters)

	returnFragment := fmt.Sprintf("return %s", edgeVarName)
	if eqb.edgeFetchMode == core.EdgeWithCompleteVertex {
		returnFragment = fmt.Sprintf("return %s, %s, %s", startVertexVarName, edgeVarName, endVertexVarName)
	}
	return fmt.Sprintf("%s %s-[%s]-%s %s %s", operation, startVertexQueryFragment, edgeQueryFragment, endVertexQueryFragment, filters, returnFragment), nil

}

func (eqb *EdgeQueryBuilder) validate() error {
	if eqb.labels == nil || len(eqb.labels) == 0 {
		return errors.New("no edge labels specified in the query")
	}

	if len(eqb.labels) > 1 {
		return errors.New("multiple edge labels cannot be specified")
	}

	if len(eqb.startVertexLabels) == 0 && eqb.startVertexVarName == "" {
		return errors.New("either start vertex label or start vertex variable name must be specified")
	}

	if len(eqb.endVertexLabels) == 0 && eqb.endVertexVarName == "" {
		return errors.New("either end vertex label or end vertex variable name must be specified")
	}

	return nil
}

func (eqb *EdgeQueryBuilder) buildEdgeQueryFragment() (string, string) {
	variableName := strings.ToLower(eqb.labels[0])[0:2]

	if eqb.varName != "" {
		variableName = eqb.varName
	}

	edgeSelector := buildSelector(eqb.selector)
	edgeLabelSelector := bytes.Buffer{}
	for _, label := range eqb.labels {
		edgeLabelSelector.WriteString(fmt.Sprintf(":%s", label))
	}
	return fmt.Sprintf("%s%s%s", variableName, edgeLabelSelector.String(), edgeSelector), variableName
}

func (eqb *EdgeQueryBuilder) buildVertexQueryFragment(vertexVarName string, vertexlabels []string, vertexSelector core.KVMap) (string, string) {
	var variableName string
	if len(vertexlabels) > 0 {
		variableName = strings.ToLower(vertexlabels[0])[0:2]
	}

	if vertexVarName != "" {
		variableName = vertexVarName
	}

	selector := buildSelector(vertexSelector)
	labelSelectors := bytes.Buffer{}
	for _, label := range vertexlabels {
		labelSelectors.WriteString(fmt.Sprintf(":%s", label))
	}
	return fmt.Sprintf("(%s%s%s)", variableName, labelSelectors.String(), selector), variableName
}
