package neo

import (
	"bytes"
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
		v := core.Vertex{}
		v.Properties = make(core.KVMap)
		node := row["v"].(neo4j.Node)
		v.Labels = append(v.Labels, node.Labels...)
		v.ID = core.NewId(node.ElementId)
		for key, val := range node.Props {
			v.Properties[key] = val
		}
		vertices = append(vertices, &v)
	}
	return vertices, nil
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

func (neo *Neo4jConnection) buildSelector(filters map[string]interface{}) string {
	if filters == nil || len(filters) == 0 {
		return ""
	}
	buffer := bytes.Buffer{}
	buffer.WriteString("{")
	firstFilterProcessed := false
	for k, v := range filters {
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

func init() {
	core.RegisterConnectorFactory("neo4j", NewConnection)
}
