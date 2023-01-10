package neo

import (
	"context"
	"strconv"
	"testing"

	"github.com/prahaladd/gograph/core"
	itests "github.com/prahaladd/gograph/integrationtests"
	"github.com/prahaladd/gograph/neo"
	"github.com/prahaladd/gograph/omg"
	"github.com/stretchr/testify/suite"
)

const (
	defaultProtocol = "neo4j"
	defaultHost     = "localhost"
	defaultRealm    = ""
	defaultUsername = "neo4j"
	defaultPassword = "neo4j"
	defaultPort     = "7687"
)

type Neo4JIntegrationTestSuite struct {
	suite.Suite
	connection core.Connection
	store      omg.Store
}

func (suite *Neo4JIntegrationTestSuite) SetupTest() {
	protocol := itests.GetFromEnvWithDefault("NEO4J_PROTOCOL", defaultProtocol)
	host := itests.GetFromEnvWithDefault("NEO4J_HOST", defaultHost)
	portString := itests.GetFromEnvWithDefault("NEO4J_PORT", "")

	var port *int32
	if len(portString) > 0 {
		parsedPort, err := strconv.ParseInt(portString, 10, 32)
		suite.NoError(err)
		port = new(int32)
		*port = int32(parsedPort)
	}

	realm := itests.GetFromEnvWithDefault("NEO4J_REALM", defaultRealm)
	user := itests.GetFromEnvWithDefault("NEO4J_USER", defaultUsername)
	pwd := itests.GetFromEnvWithDefault("NEO4J_PWD", defaultPassword)
	neo4jConnectionFactory := core.GetConnectorFactory("neo4j")
	connection, err := neo4jConnectionFactory(protocol, host, realm, port, map[string]interface{}{neo.NEO4J_USER_KEY: user, neo.NEO4J_PWD_KEY: pwd}, nil)
	suite.connection = connection
	suite.NoErrorf(err, "error whe setting up Neo4j Test : %v", err)
	suite.cleanupDB()
	suite.store = omg.NewGenericStore(connection, omg.NewReflectionMapper())
}

func (suite *Neo4JIntegrationTestSuite) TestWriteAndQuery() {
	query := "CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)"
	queryResult, err := suite.connection.ExecuteQuery(context.Background(), query, core.Write, map[string]any{"message": "hello, world"})
	suite.NoErrorf(err, "error executing result : %v", err)
	suite.Equal(1, len(queryResult.Rows))

	// attempt to read the node created in the above
	query = "MATCH (a:Greeting) return a"
	queryResult, err = suite.connection.ExecuteQuery(context.Background(), query, core.Read, nil)
	suite.NoErrorf(err, "error executing query : %v", err)
	suite.Equal(1, len(queryResult.Rows))

}

func (suite *Neo4JIntegrationTestSuite) TestVertexQuery() {
	query := "create (p:Person{name:'Tintin'})-[r:LIVES_IN{since:1929}]->(c:Country{name:'Belgium'}) return p, r, c"
	queryResult, err := suite.connection.ExecuteQuery(context.Background(), query, core.Write, map[string]any{"message": "hello, world"})
	suite.NoErrorf(err, "error executing result : %v", err)
	suite.Equal(1, len(queryResult.Rows))
	row := queryResult.Rows[0]
	suite.Equal(3, len(row))
	vertices, err := suite.connection.QueryVertex(context.Background(), "Person", core.KVMap{"name": "Tintin"}, nil, nil)
	suite.NoError(err)
	suite.NotNil(vertices)
	suite.Equal(1, len(vertices))
	vertex := vertices[0]
	suite.Equal("Tintin", vertex.Properties["name"])

	edge, err := suite.connection.QueryEdge(context.Background(), []string{"Person"}, []string{"Country"}, "LIVES_IN",
		core.KVMap{"name": "Tintin"}, core.KVMap{"name": "Belgium"}, nil, nil, nil, nil, nil, core.EdgeWithCompleteVertex)
	suite.NoError(err)
	suite.Equal(1, len(edge))
	suite.Equal("LIVES_IN", edge[0].Type)
}

func (suite *Neo4JIntegrationTestSuite) TestStoreVertex() {
	vertex := core.Vertex{
		Labels:     []string{"OMGStoreVertex"},
		Properties: core.KVMap{"Name": "Tom and Jerry", "Genre": "Kids Cartoon series"},
	}
	err := suite.connection.StoreVertex(context.Background(), &vertex)
	suite.NoError(err)

	//retrieve the vertex and validate that vertex ids
	storedVertex, err := suite.connection.QueryVertex(context.Background(), "OMGStoreVertex", core.KVMap{"Name": "Tom and Jerry"}, nil, nil)
	suite.NoError(err)
	suite.Equal(1, len(storedVertex))
	suite.Equal(storedVertex[0].ID, vertex.ID)
}

func (suite *Neo4JIntegrationTestSuite) TestStoreEdge() {
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

	err := suite.connection.StoreEdge(context.Background(), &rel)
	suite.NoError(err)

	//retrieve the vertex and validate that vertex ids
	storedVertex, err := suite.connection.QueryVertex(context.Background(), "Cartoon", core.KVMap{"Name": "Tom and Jerry"}, nil, nil)
	suite.NoError(err)
	suite.Equal(1, len(storedVertex))
	suite.Equal(storedVertex[0].ID, cv.ID)

	storedVertex, err = suite.connection.QueryVertex(context.Background(), "Team", core.KVMap{"Name": "William Hanna and Joseph Barbara"}, nil, nil)
	suite.NoError(err)
	suite.Equal(1, len(storedVertex))
	suite.Equal(storedVertex[0].ID, pv.ID)

	storedEdge, err := suite.connection.QueryEdge(context.Background(), []string{"Cartoon"}, []string{"Team"}, "CREATED_BY", nil, nil, core.KVMap{"Year": 1940}, nil, nil, nil, nil, core.EdgeWithVertexIds)
	suite.NoError(err)
	suite.Equal(1, len(storedEdge))
	suite.Equal(storedEdge[0].ID, rel.ID)
	suite.Equal(cv.ID, storedEdge[0].SourceVertexID)
	suite.Equal(pv.ID, storedEdge[0].DestinationVertexID)
}

func (suite *Neo4JIntegrationTestSuite) TestStoreEdgeWithInvalidData() {

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
		err := suite.connection.StoreEdge(context.Background(), &rel)
		suite.Error(err)
	}

}

func (suite *Neo4JIntegrationTestSuite) TestStoreOmgStructAsVertex() {
	p := person{Name: "Tom", Age: 10}
	err := suite.store.PersistVertex(context.TODO(), &p)
	suite.NoError(err)
	p2, err := suite.store.ReadVertex(context.TODO(), &person{Name: "Tom", Age: 10})
	suite.NoError(err)
	suite.Equal(1, len(p2))
	suite.Equal(p, *(p2[0].(*person)))

}

func (suite *Neo4JIntegrationTestSuite) TestStoreOmgStructsAsEdge() {
	p := person{Name: "Tom", Age: 10}
	c := city{Name: "Mumbai", PinCode: 400001}
	r := livesin{Area: "Town Hall", Since: 1982}

	vc := omg.VertexRelation{SourceVertex: &p, DestinationVertex: &c, Relationship: &r}
	err := suite.store.PersistEdge(context.TODO(), &vc)

	suite.NoError(err)

	// query the edge back
	vrs, err := suite.store.ReadEdge(context.TODO(), &vc)
	suite.NoError(err)
	suite.Equal(1, len(vrs))
	suite.Equal(vc, *vrs[0])
}

func (suite *Neo4JIntegrationTestSuite) TestStoreOmgStructsAsEdgeNoSrcVertex() {
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

func (suite *Neo4JIntegrationTestSuite) TestCustomDatabase() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, neo.ContextKeyDbName, "testdb")
	res, err := suite.connection.ExecuteQuery(ctx, "match (m)-[r]-(n) return m,r,n", core.Read, nil)
	suite.Nil(res)
	suite.Error(err)
}

func (suite *Neo4JIntegrationTestSuite) TestCustomDatabaseWithEmptyName() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, neo.ContextKeyDbName, "")
	res, err := suite.connection.ExecuteQuery(ctx, "match (m)-[r]-(n) return m,r,n", core.Read, nil)
	suite.NoError(err)
	suite.NotNil(res)
}

func (suite *Neo4JIntegrationTestSuite) TearDownTest() {
	suite.cleanupDB()
}

func (suite *Neo4JIntegrationTestSuite) cleanupDB() {
	query := "MATCH (n) DETACH DELETE(n)"
	_, err := suite.connection.ExecuteQuery(context.Background(), query, core.Write, map[string]any{"message": "hello, world"})
	suite.NoErrorf(err, "error executing result : %v", err)
}

func TestNeo4JIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(Neo4JIntegrationTestSuite))
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
