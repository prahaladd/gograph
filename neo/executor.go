package neo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/prahaladd/gograph/core"
	"github.com/prahaladd/gograph/query/cypher"
)

const (
	NEO4J_USER_KEY       = "username"
	NEO4J_PWD_KEY        = "password"
	NEO4J_AUTH_TOKEN_KEY = "auth-token"
	defaultTimeout       = 5 * time.Second
)

type neo4jContextKey string

const (
	ContextKeyDbName = neo4jContextKey("dbname")
)

type Neo4jConnection struct {
	driver neo4j.DriverWithContext
}

func (neo *Neo4jConnection) QueryVertex(ctx context.Context, label string, selectors, filters, queryParams core.KVMap) ([]*core.Vertex, error) {

	vqb := cypher.NewVertexQueryBuilder()
	vqb.SetQueryMode(core.Read)
	vqb.SetLabel([]string{label})
	vqb.SetSelector(selectors)
	vqb.SetFilters(filters)
	vqb.SetVarName("v")
	query, err := vqb.Build()
	if err != nil {
		return nil, err
	}
	qr, err := neo.ExecuteQuery(ctx, query, core.Read, queryParams)
	if err != nil {
		return nil, err
	}
	vertices := make([]*core.Vertex, 0)
	for _, row := range qr.Rows {
		v := neo.nodeToVertex(row["v"].(neo4j.Node))
		vertices = append(vertices, v)
	}
	return vertices, nil
}

func (neo *Neo4jConnection) nodeToVertex(node neo4j.Node) *core.Vertex {
	v := core.Vertex{}
	v.Properties = make(core.KVMap)
	v.Labels = append(v.Labels, node.Labels...)
	v.ID = core.NewId(node.ElementId)
	for key, val := range node.Props {
		v.Properties[key] = val
	}
	return &v
}

func (neo *Neo4jConnection) QueryEdge(ctx context.Context, startVertexLabel, endVertexLabel []string, label string, startVertexSelectors, endVertexSelectors, selectors core.KVMap, startVertexFilters, endVertexFilters, filters, queryParams core.KVMap, fetchMode core.EdgeFetchMode) ([]*core.Edge, error) {

	edgeQueryBuilder := cypher.NewEdgeQueryBuilder()
	edgeQueryBuilder.SetEdgeFetchMode(fetchMode)
	edgeQueryBuilder.SetStartVertexLabels(startVertexLabel)
	edgeQueryBuilder.SetEndVertexLabels(endVertexLabel)
	edgeQueryBuilder.SetLabel([]string{label})
	edgeQueryBuilder.SetStartVertexSelector(startVertexSelectors)
	edgeQueryBuilder.SetEndVertexSelector(endVertexSelectors)
	edgeQueryBuilder.SetSelector(selectors)
	edgeQueryBuilder.SetStartVertexFilters(startVertexFilters)
	edgeQueryBuilder.SetEndVertexFilters(endVertexFilters)
	edgeQueryBuilder.SetFilters(filters)
	edgeQueryBuilder.SetVariableName("r")
	if fetchMode == core.EdgeWithCompleteVertex {
		edgeQueryBuilder.SetStartVertexVariableName("sv")
		edgeQueryBuilder.SetEndVertexVariableName("ev")
	}

	query, err := edgeQueryBuilder.Build()

	if err != nil {
		return nil, err
	}
	qr, err := neo.ExecuteQuery(ctx, query, core.Read, filters)
	if err != nil {
		return nil, err
	}
	edges := make([]*core.Edge, 0)
	for _, row := range qr.Rows {
		e := core.Edge{}
		e.Properties = make(core.KVMap)
		relationship := row["r"].(neo4j.Relationship)
		e.Type = relationship.Type
		e.ID = core.NewId(relationship.ElementId)
		for key, val := range relationship.Props {
			e.Properties[key] = val
		}
		e.SourceVertexID = core.NewId(relationship.StartElementId)
		e.DestinationVertexID = core.NewId(relationship.EndElementId)
		if fetchMode == core.EdgeWithCompleteVertex {
			e.SourceVertex = neo.nodeToVertex(row["sv"].(neo4j.Node))
			e.DestinationVertex = neo.nodeToVertex(row["ev"].(neo4j.Node))
		}
		edges = append(edges, &e)
	}
	return edges, nil
}

func (neo *Neo4jConnection) ExecuteQuery(ctx context.Context, query string, mode core.QueryMode, queryParams map[string]interface{}) (*core.QueryResult, error) {
	var sessionConfig neo4j.SessionConfig

	var queryExecuteFn func(context.Context, neo4j.ManagedTransactionWork, ...func(*neo4j.TransactionConfig)) (any, error)
	if mode == core.Read {
		sessionConfig = neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead}
	} else {
		sessionConfig = neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	}
	// the below would work only for enterprise editions of Neo4j. Community editions work only with the default
	// neo4j database and any attempt to work with a different database would result in an error.
	graphDbName, ok := ctx.Value(ContextKeyDbName).(string)
	if ok {
		if graphDbName != "" {
			sessionConfig.DatabaseName = graphDbName
		}
	}
	session := neo.driver.NewSession(ctx, sessionConfig)
	defer session.Close(ctx)
	if mode == core.Read {
		queryExecuteFn = session.ExecuteRead
	} else {
		queryExecuteFn = session.ExecuteWrite
	}
	result, err := queryExecuteFn(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		response, err := tx.Run(ctx, query, queryParams)
		if err != nil {
			return nil, err
		}
		queryResult := core.QueryResult{}
		for response.Next(ctx) {
			m := make(core.Row)
			values := response.Record().Values
			keys := response.Record().Keys
			for i := 0; i < len(keys); i++ {
				m[keys[i]] = values[i]
			}
			queryResult.Rows = append(queryResult.Rows, m)
		}
		return queryResult, nil
	}, neo4j.WithTxTimeout(defaultTimeout))

	if err != nil {
		return nil, err
	}
	qr := result.(core.QueryResult)
	return &qr, err
}

func (neo *Neo4jConnection) Close(ctx context.Context) error {
	return neo.driver.Close(ctx)
}

func (neo *Neo4jConnection) StoreVertex(ctx context.Context, vertex *core.Vertex) error {
	vqb := cypher.NewVertexQueryBuilder()
	vqb.SetQueryMode(core.Write)
	vqb.SetLabel(vertex.Labels)
	vqb.SetSelector(vertex.Properties)
	vqb.SetVarName("sv")

	query, err := vqb.Build()
	if err != nil {
		return err
	}

	qr, err := neo.ExecuteQuery(ctx, query, core.Write, nil)
	if err != nil {
		return err
	}
	if len(qr.Rows) == 0 {
		return errors.New("unexpected error. failed to store vertex")
	}
	// support only a single vertex store at a time. hence consider only the first returned row
	row := qr.Rows[0]
	node := row["sv"].(neo4j.Node)
	vertex.ID = core.NewId(node.ElementId)

	return nil
}

func (neo *Neo4jConnection) StoreEdge(ctx context.Context, edge *core.Edge) error {

	if edge.SourceVertex == nil {
		return errors.New("source node must be specified for vertex connectivity")
	}

	eqb := cypher.NewEdgeQueryBuilder()
	eqb.SetQueryMode(core.Write)
	eqb.SetStartVertexSelector(edge.SourceVertex.Properties)
	eqb.SetStartVertexVariableName("sv")
	eqb.SetStartVertexLabels(edge.SourceVertex.Labels)

	if edge.DestinationVertex != nil {
		eqb.SetEndVertexSelector(edge.DestinationVertex.Properties)
		eqb.SetEndVertexVariableName("ev")
		eqb.SetEndVertexLabels(edge.DestinationVertex.Labels)
	}

	eqb.SetEdgeFetchMode(core.EdgeWithCompleteVertex)
	eqb.SetLabel([]string{edge.Type})
	eqb.SetVariableName("rel")
	eqb.SetSelector(edge.Properties)

	query, err := eqb.Build()

	if err != nil {
		return err
	}

	qr, err := neo.ExecuteQuery(ctx, query, core.Write, nil)

	if err != nil {
		return err
	}

	if len(qr.Rows) == 0 {
		return errors.New("unexpected error. failed to store vertex connectivity")
	}

	row := qr.Rows[0]

	edge.SourceVertex.ID = core.NewId(row["sv"].(neo4j.Node).ElementId)
	edge.ID = core.NewId(row["rel"].(neo4j.Relationship).ElementId)
	edge.DestinationVertex.ID = core.NewId(row["ev"].(neo4j.Node).ElementId)

	return nil
}

// NewConnection constructs a Neo4j driver connected to a Neo4j instance using the specified auth and config options
//
// # For Neo4j connection auth options must contain either of the following
//
// - NEO4J_USER_KEY key and NEO4J_PWD_KEY key with corresponding user name and password values
// - NEO4J_AUTH_TOKEN_KEY key with an auth token object
//
// # Absence of any of the required keys would result in an error
//
// Additional options can be configured using the options by providing a KV pair as the options paramater.
func NewConnection(protocol, host, realm string, port *int32, auth, options map[string]interface{}) (core.Connection, error) {
	if !validateAuthData(auth) {
		return nil, errors.New("specify a valid NEO4J_USER_KEY and NEO4J_PWD_KEY or a NEO4J_AUTH_TOKEN_KEY")
	}
	var token neo4j.AuthToken

	if _, ok := auth[NEO4J_AUTH_TOKEN_KEY]; !ok {
		token = neo4j.BasicAuth(auth[NEO4J_USER_KEY].(string), auth[NEO4J_PWD_KEY].(string), realm)
	} else {
		token = auth[NEO4J_AUTH_TOKEN_KEY].(neo4j.AuthToken)
	}
	var target string
	if port != nil {
		target = fmt.Sprintf("%s://%s:%d", protocol, host, *port)
	} else {
		target = fmt.Sprintf("%s://%s", protocol, host)
	}

	driver, err := neo4j.NewDriverWithContext(target, token)
	if err != nil {
		return nil, err
	}
	return &Neo4jConnection{driver: driver}, nil

}

func validateAuthData(auth map[string]interface{}) bool {
	if _, ok := auth[NEO4J_AUTH_TOKEN_KEY]; ok {
		return true
	}
	_, userNameFound := auth[NEO4J_USER_KEY]
	_, pwdFound := auth[NEO4J_PWD_KEY]
	return userNameFound && pwdFound
}

func init() {
	core.RegisterConnectorFactory("neo4j", NewConnection)
}
