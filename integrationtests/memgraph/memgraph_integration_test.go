package memgraph

import (
	"context"
	"strconv"
	"testing"

	"github.com/prahaladd/gograph/core"
	itests "github.com/prahaladd/gograph/integrationtests"
	_ "github.com/prahaladd/gograph/memgraph"
	"github.com/prahaladd/gograph/neo"
	"github.com/stretchr/testify/suite"
)

const (
	defaultProtocol = "bolt"
	defaultHost     = "localhost"
	defaultRealm    = ""
	defaultUsername = "neo4j"
	defaultPassword = "neo4j"
	defaultPort     = "7687"
)

type MemgraphIntegrationTestSuite struct {
	suite.Suite
	connection core.Connection
}

func (suite *MemgraphIntegrationTestSuite) SetupTest() {
	protocol := itests.GetFromEnvWithDefault("MG_PROTOCOL", defaultProtocol)
	host := itests.GetFromEnvWithDefault("MG_HOST", defaultHost)
	portString := itests.GetFromEnvWithDefault("MG_PORT", "")

	var port *int32
	if len(portString) > 0 {
		parsedPort, err := strconv.ParseInt(portString, 10, 32)
		suite.NoError(err)
		port = new(int32)
		*port = int32(parsedPort)
	}

	realm := itests.GetFromEnvWithDefault("MG_REALM", defaultRealm)
	user := itests.GetFromEnvWithDefault("MG_USER", defaultUsername)
	pwd := itests.GetFromEnvWithDefault("MG_PWD", defaultPassword)
	memGraphConnectionFactory := core.GetConnectorFactory("memgraph")
	connection, err := memGraphConnectionFactory(protocol, host, realm, port, map[string]interface{}{neo.NEO4J_USER_KEY: user, neo.NEO4J_PWD_KEY: pwd}, nil)
	suite.connection = connection
	suite.NoErrorf(err, "error whe setting up Neo4j Test : %v", err)
	suite.cleanupDB()
}

func (suite *MemgraphIntegrationTestSuite) TestWriteAndQuery() {
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

func (suite *MemgraphIntegrationTestSuite) TestVertexQuery() {
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

func (suite *MemgraphIntegrationTestSuite) TearDownTest() {
	suite.cleanupDB()
}

func (suite *MemgraphIntegrationTestSuite) cleanupDB() {
	query := "MATCH (n) DETACH DELETE(n)"
	_, err := suite.connection.ExecuteQuery(context.Background(), query, core.Write, map[string]any{"message": "hello, world"})
	suite.NoErrorf(err, "error executing result : %v", err)
}

func TestMemgraphIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(MemgraphIntegrationTestSuite))
}
