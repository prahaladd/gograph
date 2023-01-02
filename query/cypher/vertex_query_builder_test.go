package cypher

import (
	"strings"
	"testing"

	"github.com/prahaladd/gograph/core"
	"github.com/stretchr/testify/suite"
)

type VertexQueryBuilderTestSuite struct {
	suite.Suite
	queryBuilder *VertexQueryBuilder
}

func (suite *VertexQueryBuilderTestSuite) SetupTest() {
	suite.queryBuilder = NewVertexQueryBuilder()
}

func (suite *VertexQueryBuilderTestSuite) TestBuildQuerySingleLabelNoVariable() {
	suite.queryBuilder.SetLabel([]string{"Label1"})
	suite.queryBuilder.SetQueryMode(core.Read)
	suite.queryBuilder.SetSelector(map[string]interface{}{"name": "TestName"})
	suite.queryBuilder.SetFilters(map[string]interface{}{"age": 10})

	query, err := suite.queryBuilder.Build()
	suite.NoError(err)
	suite.True(strings.Contains(query, "MATCH"))
	mutateClausePresent := strings.Contains(query, "MERGE") || strings.Contains(query, "CREATE")
	suite.False(mutateClausePresent)
	suite.True(strings.Contains(query, "la:Label1{name:'TestName'}"))
	expectedQuery := "MATCH (la:Label1{name:'TestName'})  WHERE la.age=10 return la"
	suite.Equal(expectedQuery, query)
}

func (suite *VertexQueryBuilderTestSuite) TestBuildQuerySingleLabelVariable() {
	suite.queryBuilder.SetLabel([]string{"Label1"})
	suite.queryBuilder.SetQueryMode(core.Read)
	suite.queryBuilder.SetSelector(map[string]interface{}{"name": "TestName"})
	suite.queryBuilder.SetFilters(map[string]interface{}{"age": 10})
	suite.queryBuilder.SetVarName("var")

	query, err := suite.queryBuilder.Build()
	suite.NoError(err)
	suite.True(strings.Contains(query, "MATCH"))
	mutateClausePresent := strings.Contains(query, "MERGE") || strings.Contains(query, "CREATE")
	suite.False(mutateClausePresent)
	suite.True(strings.Contains(query, "var:Label1{name:'TestName'}"))
	expectedQuery := "MATCH (var:Label1{name:'TestName'})  WHERE var.age=10 return var"
	suite.Equal(expectedQuery, query)
}

func (suite *VertexQueryBuilderTestSuite) TestBuildQueryMultiLabelVariable() {
	suite.queryBuilder.SetLabel([]string{"Label1", "Label2"})
	suite.queryBuilder.SetQueryMode(core.Read)
	suite.queryBuilder.SetSelector(map[string]interface{}{"name": "TestName"})
	suite.queryBuilder.SetFilters(map[string]interface{}{"age": 10})

	query, err := suite.queryBuilder.Build()
	suite.NoError(err)
	suite.True(strings.Contains(query, "MATCH"))
	mutateClausePresent := strings.Contains(query, "MERGE") || strings.Contains(query, "CREATE")
	suite.False(mutateClausePresent)
	suite.True(strings.Contains(query, "la:Label1:Label2{name:'TestName'}"))
	expectedQuery := "MATCH (la:Label1:Label2{name:'TestName'})  WHERE la.age=10 return la"
	suite.Equal(expectedQuery, query)
}

func (suite *VertexQueryBuilderTestSuite) TestQueryModeWrite() {
	suite.queryBuilder.SetLabel([]string{"Label1"})
	suite.queryBuilder.SetQueryMode(core.Write)
	suite.queryBuilder.SetSelector(map[string]interface{}{"name": "TestName"})
	suite.queryBuilder.SetFilters(map[string]interface{}{"age": 10})
	suite.queryBuilder.SetVarName("var")

	query, err := suite.queryBuilder.Build()
	suite.NoError(err)
	suite.True(strings.Contains(query, "MERGE"))
	mutateClausePresent := strings.Contains(query, "MERGE") || strings.Contains(query, "CREATE")
	suite.True(mutateClausePresent)
	suite.True(strings.Contains(query, "var:Label1{name:'TestName'}"))
	expectedQuery := "MERGE (var:Label1{name:'TestName'})  WHERE var.age=10 return var"
	suite.Equal(expectedQuery, query)
}

func (suite *VertexQueryBuilderTestSuite) TestFiltersMultiple() {
	suite.queryBuilder.SetLabel([]string{"Label1"})
	suite.queryBuilder.SetQueryMode(core.Write)
	suite.queryBuilder.SetSelector(map[string]interface{}{"name": "TestName"})
	suite.queryBuilder.SetFilters(map[string]interface{}{"age": 10, "name": "TestVertex1"})
	suite.queryBuilder.SetVarName("var")
	

	query, err := suite.queryBuilder.Build()
	suite.NoError(err)
	suite.True(strings.Contains(query, "MERGE"))
	mutateClausePresent := strings.Contains(query, "MERGE") || strings.Contains(query, "CREATE")
	suite.True(mutateClausePresent)
	suite.True(strings.Contains(query, "var:Label1{name:'TestName'}"))
	
	queryComponents := strings.Split(query, " WHERE ")
	suite.Equal(2, len(queryComponents))
	filterString := queryComponents[1]
	filterString = strings.Replace(filterString, "return var", "", -1)
	filterComponents := strings.Split(filterString, " AND ")
	expectedFilterCompnents := []string{"var.age=10", "var.name='TestVertex1'"}
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

func TestVertexQueryBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(VertexQueryBuilderTestSuite))
}
