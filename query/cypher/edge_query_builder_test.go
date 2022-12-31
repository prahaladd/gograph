package cypher

import (
	"strings"
	"testing"

	"github.com/prahaladd/gograph/core"
	"github.com/stretchr/testify/suite"
)

type EdgeQueryBuilderTestSuite struct {
	suite.Suite
	edgeQueryBuilder *EdgeQueryBuilder
}

func (suite *EdgeQueryBuilderTestSuite) SetupTest() {
	suite.edgeQueryBuilder = NewEdgeQueryBuilder()
}

func (suite *EdgeQueryBuilderTestSuite) TestBuildWithNoVarNameEdgeFetchModeIds() {
	suite.edgeQueryBuilder.SetLabel([]string{"TestEdgeLabel"})
	suite.edgeQueryBuilder.SetStartVertexLabels([]string{"StartVertex"})
	suite.edgeQueryBuilder.SetEndVertexLabels([]string{"EndVertex"})
	suite.edgeQueryBuilder.SetQueryMode(core.Read)
	suite.edgeQueryBuilder.SetStartVertexSelector(core.KVMap{"name": "SV1"})
	suite.edgeQueryBuilder.SetEndVertexSelector(core.KVMap{"name": "EV1"})
	suite.edgeQueryBuilder.SetSelector(core.KVMap{"weight": 10})
	suite.edgeQueryBuilder.SetEdgeFetchMode(core.EdgeWithVertexIds)

	queryString, err := suite.edgeQueryBuilder.Build()
	suite.NoError(err)
	expectedQueryString := "MATCH (st:StartVertex{name:'SV1'})-[te:TestEdgeLabel{weight: 10}]-(en:EndVertex{name:'EV1'})  return te"
	suite.Equal(expectedQueryString, queryString)
}

func (suite *EdgeQueryBuilderTestSuite) TestBuildWithNoVarNameEdgeFetchModeComplete() {
	suite.edgeQueryBuilder.SetLabel([]string{"TestEdgeLabel"})
	suite.edgeQueryBuilder.SetStartVertexLabels([]string{"StartVertex"})
	suite.edgeQueryBuilder.SetEndVertexLabels([]string{"EndVertex"})
	suite.edgeQueryBuilder.SetQueryMode(core.Read)
	suite.edgeQueryBuilder.SetStartVertexSelector(core.KVMap{"name": "SV1"})
	suite.edgeQueryBuilder.SetEndVertexSelector(core.KVMap{"name": "EV1"})
	suite.edgeQueryBuilder.SetSelector(core.KVMap{"weight": 10})
	suite.edgeQueryBuilder.SetEdgeFetchMode(core.EdgeWithCompleteVertex)

	queryString, err := suite.edgeQueryBuilder.Build()
	suite.NoError(err)
	expectedQueryString := "MATCH (st:StartVertex{name:'SV1'})-[te:TestEdgeLabel{weight: 10}]-(en:EndVertex{name:'EV1'})  return st, te, en"
	suite.Equal(expectedQueryString, queryString)
}

func (suite *EdgeQueryBuilderTestSuite) TestBuildWithVarName() {
	suite.edgeQueryBuilder.SetLabel([]string{"TestEdgeLabel"})
	suite.edgeQueryBuilder.SetStartVertexLabels([]string{"StartVertex"})
	suite.edgeQueryBuilder.SetEndVertexLabels([]string{"EndVertex"})
	suite.edgeQueryBuilder.SetQueryMode(core.Read)
	suite.edgeQueryBuilder.SetStartVertexSelector(core.KVMap{"name": "SV1"})
	suite.edgeQueryBuilder.SetEndVertexSelector(core.KVMap{"name": "EV1"})
	suite.edgeQueryBuilder.SetSelector(core.KVMap{"weight": 10})
	suite.edgeQueryBuilder.SetEdgeFetchMode(core.EdgeWithCompleteVertex)
	suite.edgeQueryBuilder.SetVariableName("r")

	queryString, err := suite.edgeQueryBuilder.Build()
	suite.NoError(err)
	expectedQueryString := "MATCH (st:StartVertex{name:'SV1'})-[r:TestEdgeLabel{weight: 10}]-(en:EndVertex{name:'EV1'})  return st, r, en"
	suite.Equal(expectedQueryString, queryString)
}

func (suite *EdgeQueryBuilderTestSuite) TestBuildWithMultipleStartVertexLabels() {
	suite.edgeQueryBuilder.SetLabel([]string{"TestEdgeLabel"})
	suite.edgeQueryBuilder.SetStartVertexLabels([]string{"StartVertex", "StartVertex1"})
	suite.edgeQueryBuilder.SetEndVertexLabels([]string{"EndVertex"})
	suite.edgeQueryBuilder.SetQueryMode(core.Read)
	suite.edgeQueryBuilder.SetStartVertexSelector(core.KVMap{"name": "SV1"})
	suite.edgeQueryBuilder.SetEndVertexSelector(core.KVMap{"name": "EV1"})
	suite.edgeQueryBuilder.SetSelector(core.KVMap{"weight": 10})
	suite.edgeQueryBuilder.SetEdgeFetchMode(core.EdgeWithCompleteVertex)
	suite.edgeQueryBuilder.SetVariableName("r")

	queryString, err := suite.edgeQueryBuilder.Build()
	suite.NoError(err)
	expectedQueryString := "MATCH (st:StartVertex:StartVertex1{name:'SV1'})-[r:TestEdgeLabel{weight: 10}]-(en:EndVertex{name:'EV1'})  return st, r, en"
	suite.Equal(expectedQueryString, queryString)
}

func (suite *EdgeQueryBuilderTestSuite) TestBuildWithMultipleEndVertexLabels() {
	suite.edgeQueryBuilder.SetLabel([]string{"TestEdgeLabel"})
	suite.edgeQueryBuilder.SetStartVertexLabels([]string{"StartVertex"})
	suite.edgeQueryBuilder.SetEndVertexLabels([]string{"EndVertex", "EndVertex1"})
	suite.edgeQueryBuilder.SetQueryMode(core.Read)
	suite.edgeQueryBuilder.SetStartVertexSelector(core.KVMap{"name": "SV1"})
	suite.edgeQueryBuilder.SetEndVertexSelector(core.KVMap{"name": "EV1"})
	suite.edgeQueryBuilder.SetSelector(core.KVMap{"weight": 10})
	suite.edgeQueryBuilder.SetEdgeFetchMode(core.EdgeWithCompleteVertex)
	suite.edgeQueryBuilder.SetVariableName("r")

	queryString, err := suite.edgeQueryBuilder.Build()
	suite.NoError(err)
	expectedQueryString := "MATCH (st:StartVertex{name:'SV1'})-[r:TestEdgeLabel{weight: 10}]-(en:EndVertex:EndVertex1{name:'EV1'})  return st, r, en"
	suite.Equal(expectedQueryString, queryString)
}

func (suite *EdgeQueryBuilderTestSuite) TestQueryModeWrite() {
	suite.edgeQueryBuilder.SetLabel([]string{"TestEdgeLabel"})
	suite.edgeQueryBuilder.SetStartVertexLabels([]string{"StartVertex"})
	suite.edgeQueryBuilder.SetEndVertexLabels([]string{"EndVertex"})
	suite.edgeQueryBuilder.SetQueryMode(core.Write)
	suite.edgeQueryBuilder.SetStartVertexSelector(core.KVMap{"name": "SV1"})
	suite.edgeQueryBuilder.SetEndVertexSelector(core.KVMap{"name": "EV1"})
	suite.edgeQueryBuilder.SetSelector(core.KVMap{"weight": 10})
	suite.edgeQueryBuilder.SetEdgeFetchMode(core.EdgeWithCompleteVertex)

	queryString, err := suite.edgeQueryBuilder.Build()
	suite.NoError(err)
	expectedQueryString := "MERGE (st:StartVertex{name:'SV1'})-[te:TestEdgeLabel{weight: 10}]-(en:EndVertex{name:'EV1'})  return st, te, en"
	suite.Equal(expectedQueryString, queryString)
}

func (suite *EdgeQueryBuilderTestSuite) TestBuildWithMultipleStartAndEndVertexLabels() {
	suite.edgeQueryBuilder.SetLabel([]string{"TestEdgeLabel"})
	suite.edgeQueryBuilder.SetStartVertexLabels([]string{"StartVertex", "StartVertex1"})
	suite.edgeQueryBuilder.SetEndVertexLabels([]string{"EndVertex", "EndVertex1"})
	suite.edgeQueryBuilder.SetQueryMode(core.Read)
	suite.edgeQueryBuilder.SetStartVertexSelector(core.KVMap{"name": "SV1"})
	suite.edgeQueryBuilder.SetEndVertexSelector(core.KVMap{"name": "EV1"})
	suite.edgeQueryBuilder.SetSelector(core.KVMap{"weight": 10})
	suite.edgeQueryBuilder.SetEdgeFetchMode(core.EdgeWithCompleteVertex)
	suite.edgeQueryBuilder.SetVariableName("r")

	queryString, err := suite.edgeQueryBuilder.Build()
	suite.NoError(err)
	expectedQueryString := "MATCH (st:StartVertex:StartVertex1{name:'SV1'})-[r:TestEdgeLabel{weight: 10}]-(en:EndVertex:EndVertex1{name:'EV1'})  return st, r, en"
	suite.Equal(expectedQueryString, queryString)
}

func (suite *EdgeQueryBuilderTestSuite) TestBuildWithFilters() {
	suite.edgeQueryBuilder.SetLabel([]string{"TestEdgeLabel"})
	suite.edgeQueryBuilder.SetStartVertexLabels([]string{"StartVertex"})
	suite.edgeQueryBuilder.SetEndVertexLabels([]string{"EndVertex"})
	suite.edgeQueryBuilder.SetQueryMode(core.Read)
	suite.edgeQueryBuilder.SetStartVertexSelector(core.KVMap{"name": "SV1"})
	suite.edgeQueryBuilder.SetEndVertexSelector(core.KVMap{"name": "EV1"})
	suite.edgeQueryBuilder.SetSelector(core.KVMap{"weight": 10})
	suite.edgeQueryBuilder.SetEdgeFetchMode(core.EdgeWithVertexIds)

	suite.edgeQueryBuilder.SetStartVertexFilters(core.KVMap{"name": "SV1"})
	suite.edgeQueryBuilder.SetEndVertexFilters(core.KVMap{"name": "EV1"})
	suite.edgeQueryBuilder.SetFilters(core.KVMap{"name": "TestEdge", "weight": 100})

	queryString, err := suite.edgeQueryBuilder.Build()
	suite.NoError(err)

	queryComponents := strings.Split(queryString, " WHERE ")
	suite.Equal(2, len(queryComponents))
	filterString := queryComponents[1]
	filterString = strings.Replace(filterString, "return te", "", -1)
	filterComponents := strings.Split(filterString, " AND ")
	expectedFilterCompnents := []string{"st.name='SV1'", "en.name='EV1'", "te.name='TestEdge'", "te.weight=100"}
	actualFilterComponentsSet := make(map[string]bool)
	for _, v := range filterComponents {
		actualFilterComponentsSet[strings.TrimLeft(strings.TrimRight(v, " "), " ")] = true
	}
	allComponentsFound := true
	for _, v := range expectedFilterCompnents {
		if _, found := actualFilterComponentsSet[v]; !found {
			allComponentsFound = false
		}
	}
	suite.True(allComponentsFound)
}

func (suite *EdgeQueryBuilderTestSuite) TestBuildMultipleEdgeLabels() {
	suite.edgeQueryBuilder.SetLabel([]string{"label1", "label2"})
	queryString, err := suite.edgeQueryBuilder.Build()
	suite.Error(err)
	suite.Equal("", queryString)
}

func TestEdgeQueryBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(EdgeQueryBuilderTestSuite))
}
