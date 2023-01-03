package neo

import (
	"context"
	"strconv"
	"testing"

	"github.com/prahaladd/gograph/core"
	itests "github.com/prahaladd/gograph/integrationtests"
	"github.com/prahaladd/gograph/neo"
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
	connection, err := neo.NewConnection(protocol, host, realm, port, map[string]interface{}{neo.NEO4J_USER_KEY: user, neo.NEO4J_PWD_KEY: pwd}, nil)
	suite.connection = connection
	suite.NoErrorf(err, "error whe setting up Neo4j Test : %v", err)
	suite.cleanupDB()
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
