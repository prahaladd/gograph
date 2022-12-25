package core

import (
	"context"
	"fmt"
)

// Identifier defines an in interface to be implemented by all comparable types serving as Graph node identifiers.
type Identifier struct {
	value any
}

func (id *Identifier) Value() any {
	return id.value
}

func (id *Identifier) String() string {
	return fmt.Sprintf("%v", id.value)
}

func NewId(value any) *Identifier {
	return &Identifier{value: value}
}

// Properties is a type alias for a map of key value pairs representing vertex or edge properties
type KVMap map[string]interface{}

// GraphElement defines the interface to be implemented by all elements (Vertices and Edges) of a graph.
type GraphElement interface {
	// GetId returns the identifier of the graph element as present in the underlying Graph DBMS
	GetId() Identifier

	// GetLabel returns the set of labels associated with a Graph element
	GetLabel() []string

	// GetProperties returns the set of properties associated with the graph element
	GetProperties() KVMap
}

// Vertex represents a vertex within the graph
type Vertex struct {
	ID         Identifier
	Label      string
	Properties KVMap
}

// Edge represents an edge within the graph
type Edge struct {
	ID                  Identifier
	Label               string
	SourceVertexID      Identifier
	DestinationVertexID Identifier
	Properties          KVMap
}

// Row is a type alias for a simple KV map and represents a single row of record obtained from the graph db
type Row map[string]interface{}

// QueryResult represents the result of a query execution and is made up of 0 or more rows.
type QueryResult struct {
	Rows []Row
}

type QueryMode int8

const (
	Read QueryMode = iota
	Write
)

// Connection interface represents the contracts to be satisfied by individual connection implementations
type Connection interface {

	// QueryVertex returns a vertex from the graph for the specified label
	//
	// selectors are required to "select" a particular node within the graph. If selectors are not specified, then all nodes in the graph
	// with the specified label woould be selected.
	//
	// filters are used filter out the results from the set of selected nodes
	QueryVertex(ctx context.Context, label string, selectors, filters map[string]interface{}) ([]*Vertex, error)

	// ExecuteReadQuery executes a query and transforms the native result set obtained from the DB to a QueryResult using the specified transform function
	//
	// The specified query must be a valid Cypher or Gremlin query.
	//
	// The mode parameter specifies the mode of the query - read or write. This parameter is required in
	// order to handle API invocations for certain drivers such as Neo4J where-in the query mode would
	// determine the access mode of the session being used to execute the query
	//
	// The queryParams parameter can be used to inject dynamic data into the query. This is an optional argument and can be nil
	//
	// The context can contain additional query and session configuration parameters required for execution
	ExecuteQuery(ctx context.Context, query string, mode QueryMode, queryParams map[string]interface{}) (*QueryResult, error)

	// Close closes the connection to a database.
	//
	// Not all implementations of the below method ould actually close a connection. For e.g. if the database is being
	//
	// updated using an HTTP(s) interface then there is no requirement to close the connection explicitly.
	Close(ctx context.Context) error
}

// GetConnection returns a connection to a specified graph type.
func GetConnection(graphDBType string, protocol, host, realm string, port *int32, auth, options map[string]interface{}) (Connection, error) {
	factory := GetConnectorFactory(graphDBType)
	if factory == nil {
		return nil, fmt.Errorf("cannot find connection factory for graph type: %s", graphDBType)
	}
	return factory(protocol, host, realm, port, auth, options)
}
