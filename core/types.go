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
	GetId() *Identifier

	// GetLabel returns the set of labels associated with a Graph element
	GetLabel() []string

	// GetProperties returns the set of properties associated with the graph element
	GetProperties() KVMap
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

type EdgeFetchMode int8

const (
	EdgeWithVertexIds = iota
	EdgeWithCompleteVertex
)

// QueryOptions is used to control the aspects of building and routing the final cypher query to be sent to the graph db.
type QueryOptions struct {
	options KVMap
}

// Connection interface represents the contracts to be satisfied by individual connection implementations
type Connection interface {

	// QueryVertex returns a vertex from the graph for the specified label
	//
	// selectors are required to "select" a particular node within the graph. If selectors are not specified, then all nodes in the graph
	// with the specified label woould be selected.
	//
	// filters are used to filter out the results from the set of selected nodes
	QueryVertex(ctx context.Context, label string, selectors, filters, queryParams KVMap) ([]*Vertex, error)

	// QueryEdge returns a set of edges for the specified label
	//
	// selctors are required to select a particular relationship within the graph. If selectors are not specified, then all edges in the graph
	// with the specified labels would be selected
	//
	// filters are used to filter out the results from the set of selected edges
	//
	// The level of detail about the start and end nodes of an edge  can be controled by the fetch mode. Currently, the library
	// supports returning edges where-in the ids of the start and end vertices of the relations are available.
	QueryEdge(ctx context.Context, startVertexLabel, endVertexLabel []string, label string, startVertexSelectors, endVertexSelectors, selectors KVMap, startVertexFilters, endVertexFilters, filters KVMap, queryParams KVMap, fetchMode EdgeFetchMode) ([]*Edge, error)

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

	// StoreVertex stores a vertex to the underlying graph database.
	//
	// Upon successful storage, the passed in vertex object's ID field would be set to the ID returned by the database.
	// Returns an error if there is a failure when persisting the vertex
	StoreVertex(ctx context.Context, vertex *Vertex) error

	// StoreEdge stores a connected component to the graph database. It can be used to create a new relation
	// between two vertices or update the properties for an existing relation
	//
	// Upon successful storage, the ID field of the participating vertex and edge object are populated
	// with the DB specific identifier.
	// Returns an error if there is a failure when persisting the edge
	StoreEdge(ctx context.Context, edge *Edge) error
}

// GetConnection returns a connection to a specified graph type.
func GetConnection(graphDBType string, protocol, host, realm string, port *int32, auth, options map[string]interface{}) (Connection, error) {
	factory := GetConnectorFactory(graphDBType)
	if factory == nil {
		return nil, fmt.Errorf("cannot find connection factory for graph type: %s", graphDBType)
	}
	return factory(protocol, host, realm, port, auth, options)
}
