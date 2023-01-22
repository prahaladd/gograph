package agensgraph

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/prahaladd/gograph/agensgraph"
	"github.com/prahaladd/gograph/core"
	"github.com/prahaladd/gograph/integrationtests"
	"github.com/prahaladd/gograph/omg"
	"github.com/stretchr/testify/suite"
)

type AgensGraphIntegrationTestSuite struct {
	suite.Suite
	connection       core.Connection
	store            omg.Store
	context          context.Context
	vlabelsToCleanUp []string
	elabelsToCleanUp []string
}

func (suite *AgensGraphIntegrationTestSuite) SetupTest() {
	host := integrationtests.GetFromEnvWithDefault("AGENS_HOST", "localhost")
	portFromEnv := integrationtests.GetFromEnvWithDefault("AGENS_PORT", "5432")
	portParsed, err := strconv.ParseInt(portFromEnv, 10, 32)
	suite.NoError(err)
	port := new(int32)
	*port = int32(portParsed)
	db := integrationtests.GetFromEnvWithDefault("AGENS_DB", "")
	suite.NotEqual("", db)
	protocol := integrationtests.GetFromEnvWithDefault("AGENS_PROTOCOL", "")
	userName := integrationtests.GetFromEnvWithDefault("AGENS_USER", "")
	suite.NotEqual("", userName)
	pwd := integrationtests.GetFromEnvWithDefault("AGENS_PWD", "")
	suite.NotEqual("", pwd)
	agensConnectionFactory := core.GetConnectorFactory("agens")
	conn, err := agensConnectionFactory(protocol, host, "", port, core.KVMap{agensgraph.AGENS_PASSWD_KEY: pwd, agensgraph.AGENS_USER_KEY: userName}, core.KVMap{agensgraph.AGENS_DBNAME_KEY: db})
	suite.NoError(err)
	suite.connection = conn
	suite.context = context.Background()
	suite.context = context.WithValue(suite.context, agensgraph.ContextKeyGraphName, "agens")
	suite.context = context.WithValue(suite.context, agensgraph.ContextKeyWriteModeCreate, true)
	suite.vlabelsToCleanUp = make([]string, 0)
	suite.elabelsToCleanUp = make([]string, 0)
	suite.cleanupDB()
	suite.store = omg.NewGenericStore(suite.connection, omg.NewReflectionMapper())

}

func (suite *AgensGraphIntegrationTestSuite) TestWriteAndQuery() {
	suite.vlabelsToCleanUp = append(suite.vlabelsToCleanUp, "Greeting")
	query := "CREATE (a:Greeting) SET a.message = 'hello, world' RETURN a.message + ' from node ', id(a)"
	queryResult, err := suite.connection.ExecuteQuery(suite.context, query, core.Write, nil)
	suite.NoErrorf(err, "error executing result : %v", err)
	suite.Equal(1, len(queryResult.Rows))

	// attempt to read the node created in the above
	query = "MATCH (a:Greeting) return a"
	queryResult, err = suite.connection.ExecuteQuery(suite.context, query, core.Read, nil)
	suite.NoErrorf(err, "error executing query : %v", err)
	suite.Equal(1, len(queryResult.Rows))

}

func (suite *AgensGraphIntegrationTestSuite) TestVertexQuery() {
	suite.vlabelsToCleanUp = append(suite.vlabelsToCleanUp, "Person", "Country")
	suite.elabelsToCleanUp = append(suite.elabelsToCleanUp, "LIVES_IN")
	query := "create (p:Person{name:'Tintin'})-[r:LIVES_IN{since:1929}]->(c:Country{name:'Belgium'}) return p, r, c"
	queryResult, err := suite.connection.ExecuteQuery(suite.context, query, core.Write, map[string]any{"message": "hello, world"})
	suite.NoErrorf(err, "error executing result : %v", err)
	suite.Equal(1, len(queryResult.Rows))
	row := queryResult.Rows[0]
	suite.Equal(3, len(row))
	vertices, err := suite.connection.QueryVertex(suite.context, "Person", core.KVMap{"name": "Tintin"}, nil, nil)
	suite.NoError(err)
	suite.NotNil(vertices)
	suite.Equal(1, len(vertices))
	vertex := vertices[0]
	suite.Equal("Tintin", vertex.Properties["name"])

	edge, err := suite.connection.QueryEdge(suite.context, []string{"Person"}, []string{"Country"}, "LIVES_IN",
		core.KVMap{"name": "Tintin"}, core.KVMap{"name": "Belgium"}, nil, nil, nil, nil, nil, core.EdgeWithCompleteVertex)
	suite.NoError(err)
	suite.Equal(1, len(edge))
	suite.Equal(strings.ToLower("LIVES_IN"), edge[0].Type)
}

func (suite *AgensGraphIntegrationTestSuite) TestStoreVertex() {
	suite.vlabelsToCleanUp = append(suite.vlabelsToCleanUp, "OMGStoreVertex")
	vertex := core.Vertex{
		Labels:     []string{"OMGStoreVertex"},
		Properties: core.KVMap{"Name": "Tom and Jerry", "Genre": "Kids Cartoon series"},
	}
	err := suite.connection.StoreVertex(suite.context, &vertex)
	suite.NoError(err)

	//retrieve the vertex and validate that vertex ids
	storedVertex, err := suite.connection.QueryVertex(suite.context, "OMGStoreVertex", core.KVMap{"Name": "Tom and Jerry"}, nil, nil)
	suite.NoError(err)
	suite.Equal(1, len(storedVertex))
	suite.Equal(storedVertex[0].ID, vertex.ID)
}

func (suite *AgensGraphIntegrationTestSuite) TestStoreEdge() {
	suite.vlabelsToCleanUp = append(suite.vlabelsToCleanUp, "Cartoon", "Team")
	suite.elabelsToCleanUp = append(suite.elabelsToCleanUp, "CREATED_BY")

	cv := core.Vertex{
		Labels:     []string{"Cartoon"},
		Properties: core.KVMap{"Name": "Tom and Jerry", "Genre": "Kids Cartoon series"},
	}

	pv := core.Vertex{
		Labels:     []string{"Team"},
		Properties: core.KVMap{"Name": "William Hanna and Joseph Barbara"},
	}

	rel := core.Edge{
		Type:              "CREATED_BY",
		Properties:        core.KVMap{"Year": 1940},
		SourceVertex:      &cv,
		DestinationVertex: &pv,
	}

	err := suite.connection.StoreEdge(suite.context, &rel)
	suite.NoError(err)

	//retrieve the vertex and validate that vertex ids
	storedVertex, err := suite.connection.QueryVertex(suite.context, "Cartoon", core.KVMap{"Name": "Tom and Jerry"}, nil, nil)
	suite.NoError(err)
	suite.Equal(1, len(storedVertex))
	suite.Equal(storedVertex[0].ID, cv.ID)

	storedVertex, err = suite.connection.QueryVertex(suite.context, "Team", core.KVMap{"Name": "William Hanna and Joseph Barbara"}, nil, nil)
	suite.NoError(err)
	suite.Equal(1, len(storedVertex))
	suite.Equal(storedVertex[0].ID, pv.ID)

	storedEdge, err := suite.connection.QueryEdge(suite.context, []string{"Cartoon"}, []string{"Team"}, "CREATED_BY", nil, nil, core.KVMap{"Year": 1940}, nil, nil, nil, nil, core.EdgeWithVertexIds)
	suite.NoError(err)
	suite.Equal(1, len(storedEdge))
	suite.Equal(storedEdge[0].ID, rel.ID)
	suite.Equal(cv.ID, storedEdge[0].SourceVertexID)
	suite.Equal(pv.ID, storedEdge[0].DestinationVertexID)
}

func (suite *AgensGraphIntegrationTestSuite) TestStoreEdgeWithInvalidData() {

	pv := core.Vertex{
		Labels:     []string{"Team"},
		Properties: core.KVMap{"Name": "William Hanna and Joseph Barbara"},
	}

	{
		rel := core.Edge{
			Type:              "CREATED_BY",
			Properties:        core.KVMap{"Year": 1940},
			DestinationVertex: &pv,
		}
		err := suite.connection.StoreEdge(suite.context, &rel)
		suite.Error(err)
	}

}

func (suite *AgensGraphIntegrationTestSuite) TestStoreOmgStructAsVertex() {
	suite.vlabelsToCleanUp = append(suite.vlabelsToCleanUp, "person")
	p := person{Name: "Tom", Age: 10}
	err := suite.store.PersistVertex(suite.context, &p)
	suite.NoError(err)
	p2, err := suite.store.ReadVertex(suite.context, &person{Name: "Tom", Age: 10})
	suite.NoError(err)
	suite.Equal(1, len(p2))
	suite.Equal(p, *(p2[0].(*person)))

}

func (suite *AgensGraphIntegrationTestSuite) TestStoreOmgStructsAsEdge() {
	suite.vlabelsToCleanUp = append(suite.vlabelsToCleanUp, "person", "city")
	suite.elabelsToCleanUp = append(suite.elabelsToCleanUp, "lives_in")

	p := person{Name: "Tom", Age: 10}
	c := city{Name: "Mumbai", PinCode: 400001}
	r := livesin{Area: "Town Hall", Since: 1982}

	vc := omg.VertexRelation{SourceVertex: &p, DestinationVertex: &c, Relationship: &r}
	err := suite.store.PersistEdge(suite.context, &vc)

	suite.NoError(err)

	// query the edge back
	vrs, err := suite.store.ReadEdge(suite.context, &vc)
	suite.NoError(err)
	suite.Equal(1, len(vrs))
	suite.Equal(vc, *vrs[0])
}

func (suite *AgensGraphIntegrationTestSuite) TestStoreOmgStructsAsEdgeNoSrcVertex() {
	p := person{Name: "Tom", Age: 10}
	c := city{Name: "Mumbai", PinCode: 400001}
	r := livesin{Area: "Town Hall", Since: 1982}

	{
		vc := omg.VertexRelation{SourceVertex: &p, DestinationVertex: &c, Relationship: &r}
		// query the edge back with no source vertex -- should result in an error
		vc.SourceVertex = nil
		_, err := suite.store.ReadEdge(context.TODO(), &vc)
		suite.Error(err)
	}

	{
		vc := omg.VertexRelation{SourceVertex: &p, DestinationVertex: &c, Relationship: &r}
		// query the edge back with no destination vertex -- should result in an error
		vc.DestinationVertex = nil
		_, err := suite.store.ReadEdge(context.TODO(), &vc)
		suite.Error(err)
	}

	{
		vc := omg.VertexRelation{SourceVertex: &p, DestinationVertex: &c, Relationship: &r}
		// query the edge back with no relation -- should result in an error
		vc.Relationship = nil
		_, err := suite.store.ReadEdge(context.TODO(), &vc)
		suite.Error(err)
	}

	{
		vc := omg.VertexRelation{}
		// query the edge back with empty vertex relation
		_, err := suite.store.ReadEdge(context.TODO(), &vc)
		suite.Error(err)
	}

}

func (suite *AgensGraphIntegrationTestSuite) TestGraphNameNotSpecified() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, agensgraph.ContextKeyGraphName, "")
	res, err := suite.connection.ExecuteQuery(ctx, "match (m)-[r]-(n) return m,r,n", core.Read, nil)
	suite.Nil(res)
	suite.Error(err)
}

type person struct {
	Name string
	Age  int64
}

func (p *person) GetLabel() string {
	return "person"
}

func (p *person) GetType() omg.GraphObjectType {
	return omg.Vertex
}

type city struct {
	Name    string
	PinCode int64
}

// GetLabel returns the label associated with the graph object
func (c *city) GetLabel() string {
	return "City"
}

// GraphType() returns the type associated with the Graph object
func (c *city) GetType() omg.GraphObjectType {
	return omg.Vertex
}

type livesin struct {
	Since int64
	Area  string
}

// GetLabel returns the label associated with the graph object
func (lv *livesin) GetLabel() string {
	return "LIVES_IN"
}

// GraphType() returns the type associated with the Graph object
func (lv *livesin) GetType() omg.GraphObjectType {
	return omg.Edge
}

func (suite *AgensGraphIntegrationTestSuite) TearDownTest() {
	suite.cleanupDB()
}

func (suite *AgensGraphIntegrationTestSuite) cleanupDB() {
	query := "MATCH (n) DETACH DELETE(n)"
	ctx := context.TODO()
	ctx = context.WithValue(ctx, agensgraph.ContextKeyGraphName, "agens")
	_, err := suite.connection.ExecuteQuery(ctx, query, core.Write, map[string]any{"message": "hello, world"})
	suite.NoErrorf(err, "error executing result : %v", err)
	// delete all vlabels and elabels to set the stage for the next run
	buff := strings.Builder{}
	for _, vlabel := range suite.vlabelsToCleanUp {
		if buff.Len() > 0 {
			buff.WriteString("\n")
		}
		buff.WriteString(fmt.Sprintf("DROP VLABEL %s;", vlabel))
	}
	for _, elabel := range suite.elabelsToCleanUp {
		if buff.Len() > 0 {
			buff.WriteString("\n")
		}
		buff.WriteString(fmt.Sprintf("DROP ELABEL %s;", elabel))
	}
	if buff.Len() > 0 {
		_, err = suite.connection.ExecuteQuery(ctx, buff.String(), core.Write, nil)
		suite.NoError(err)
	}

}

func TestAgensGraphIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AgensGraphIntegrationTestSuite))
}
