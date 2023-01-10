package omg

import (
	"context"
	"errors"
	"reflect"

	"github.com/prahaladd/gograph/core"
)

// Store represents a contract to be implemented by a persistence layer
// which persists or reads OMG based types to the underlying graph database
type Store interface {

	// PersistVertex persists a struct implementing the GraphObject interface to
	// the underlying graph database.
	//
	// Returns errors encountered during persistence.
	PersistVertex(context.Context, GraphObject) error

	// ReadVertex reads a vertex from the graph database using the selectors specified
	// within the example vertex. The example vertex need not be a fully formed
	// entity, the query generated to read the vertex would consider all non empty
	// fields from the struct to generate the vertex selectors
	ReadVertex(context.Context, GraphObject) ([]GraphObject, error)

	// PersistEdge persists a vertex relation to the underlying graph database
	// A vertex relation is a concise mechanism to declare a relationship.
	//
	// A specified vertex relation must adhere to the following rules:
	//
	// - source vertex cannot be nil
	//
	// - If both source and destination vertex are specified, then edge cannot be nil
	//
	// The use case where-in only the source vertex is specified, but the relationship and the
	// destination vertex is nil is equivalent to  creating a single isolated vertex
	// from the graph database
	PersistEdge(context.Context, *VertexRelation) error

	// ReadEdge reads an edge along with the associated vertices using the vertex and edge selectors
	// specified within the example vertex relation object.
	//
	// Returns a list of all vertex relations satisfying the example on sucess, an error other wise.
	//
	// A vertex relation is a concise mechanism to declare a relationship.
	//
	// The returned result depends upon the contents of the passed in example relation:
	//
	// - source vertex =  nil, relationship != nil, destination != nil returns all source vertices that have the specified relation
	// to the specified destination vertex
	//
	// - source vertex != nil, relationship != nil and destination != nil returns all destination vertices that have the specified
	// relation to the source vertex
	//
	// - source vertex != nil, relationship = nil and destination != nil is an Invalid input
	//
	// - source vertex = nil, relationship = nil and destination = nil is an Invalid input
	//
	// The use case where-in only the source vertex is specified, but the relationship and the
	// destination vertex is nil is equivalent to  querying for a single isolated vertex
	// from the graph database
	ReadEdge(context.Context, *VertexRelation) ([]*VertexRelation, error)
}

type GenericStore struct {
	connection core.Connection
	mapper     Mapper
}

// PersistVertex persists a struct implementing the GraphObject interface to
// the underlying graph database.
//
// Returns errors encountered during persistence.
func (gs *GenericStore) PersistVertex(ctx context.Context, vertex GraphObject) error {
	if vertex.GetType() != Vertex {
		return errors.New("specified value must be of graph object type vertex")
	}
	v, err := gs.mapper.ToVertex(vertex, []string{vertex.GetLabel()})
	if err != nil {
		return err
	}
	return gs.connection.StoreVertex(ctx, v)
}

// ReadVertex reads a vertex from the graph database using the selectors specified
// within the example vertex. The example vertex need not be a fully formed
// entity, the query generated to read the vertex would consider all non empty
// fields from the struct to generate the vertex selectors
func (gs *GenericStore) ReadVertex(ctx context.Context, exampleVertex GraphObject) ([]GraphObject, error) {
	if exampleVertex.GetType() != Vertex {
		return nil, errors.New("specified value must be of graph object type vertex")
	}
	v, err := gs.mapper.ToVertex(exampleVertex, []string{exampleVertex.GetLabel()})
	if err != nil {
		return nil, err
	}

	resultVertices, err := gs.connection.QueryVertex(ctx, v.GetLabel()[0], v.GetProperties(), nil, nil)
	if err != nil {
		return nil, err
	}

	toRet := make([]GraphObject, 0)
	for _, rv := range resultVertices {
		graphObj := reflect.New(reflect.TypeOf(exampleVertex).Elem())
		gs.mapper.FromVertex(rv, graphObj.Interface())
		toRet = append(toRet, graphObj.Interface().(GraphObject))
	}

	return toRet, nil
}

// PersistEdge persists a vertex relation to the underlying graph database
// A vertex relation is a concise mechanism to declare a relationship.
//
// A specified vertex relation must adhere to the following rules:
//
// - source vertex cannot be nil
//
// - If both source and destination vertex are specified, then edge cannot be nil
//
// The use case where-in only the source vertex is specified, but the relationship and the
// destination vertex is nil is equivalent to  creating a single isolated vertex
// from the graph database
func (gs *GenericStore) PersistEdge(ctx context.Context, edge *VertexRelation) error {
	if edge.SourceVertex == nil {
		return errors.New("source vertex cannot be nil")
	}
	if edge.SourceVertex == nil && edge.DestinationVertex == nil {
		return errors.New("source vertex and destination vertex cannot be nil")
	}
	if edge.SourceVertex != nil && edge.DestinationVertex != nil && edge.Relationship == nil {
		return errors.New("edge cannot be nil when source and destination vertices are specified")
	}
	// validate that the types are correct
	if edge.SourceVertex.GetType() != Vertex || edge.DestinationVertex.GetType() != Vertex {
		return errors.New("the type of source or destination vertex must be Vertex")
	}
	if edge.Relationship.GetType() != Edge {
		return errors.New("the type of relationship must be Edge")
	}
	srcVertex, err := gs.mapper.ToVertex(edge.SourceVertex, []string{edge.SourceVertex.GetLabel()})
	if err != nil {
		return err
	}
	destVertex, err := gs.mapper.ToVertex(edge.DestinationVertex, []string{edge.DestinationVertex.GetLabel()})
	if err != nil {
		return err
	}
	relType := edge.Relationship.GetLabel()

	rel, err := gs.mapper.ToEdge(edge.Relationship, &relType)

	if err != nil {
		return err
	}

	rel.SourceVertex = srcVertex
	rel.DestinationVertex = destVertex

	return gs.connection.StoreEdge(ctx, rel)
}

// ReadEdge reads an edge along with the associated vertices using the vertex and edge selectors
// specified within the example vertex relation object.
//
// Returns a list of all vertex relations satisfying the example on sucess, an error other wise.
//
// A vertex relation is a concise mechanism to declare a relationship.

func (gs *GenericStore) ReadEdge(ctx context.Context, exampleEdge *VertexRelation) ([]*VertexRelation, error) {
	emptyExample := VertexRelation{}
	if exampleEdge == &emptyExample {
		return nil, errors.New("no source vertex, destination vertex and relation specified")
	}
	if exampleEdge.SourceVertex == nil || exampleEdge.DestinationVertex == nil || exampleEdge.Relationship == nil {
		return nil, errors.New("vertex relation must contain the example source vertex, destination vertex and relation to query")
	}
	var srcVertex, destVertex *core.Vertex
	var rel *core.Edge
	var err error

	if exampleEdge.SourceVertex != nil {
		srcVertex, err = gs.mapper.ToVertex(exampleEdge.SourceVertex, []string{exampleEdge.SourceVertex.GetLabel()})
		if err != nil {
			return nil, err
		}
	}
	if exampleEdge.DestinationVertex != nil {
		destVertex, err = gs.mapper.ToVertex(exampleEdge.DestinationVertex, []string{exampleEdge.DestinationVertex.GetLabel()})
		if err != nil {
			return nil, err
		}
	}
	if exampleEdge.Relationship != nil {
		relLabel := exampleEdge.Relationship.GetLabel()
		rel, err = gs.mapper.ToEdge(exampleEdge.Relationship, &relLabel)
		if err != nil {
			return nil, err
		}
		relLabel = rel.Type
	}
	rel.SourceVertex = srcVertex
	rel.DestinationVertex = destVertex
	edges, err := gs.connection.QueryEdge(ctx, srcVertex.Labels, destVertex.Labels, rel.Type, srcVertex.Properties, destVertex.Properties, rel.Properties, nil, nil, nil, nil, core.EdgeWithCompleteVertex)
	if err != nil {
		return nil, err
	}

	vrs := make([]*VertexRelation, 0)
	for _, edge := range edges {
		vr := VertexRelation{}
		srcVertexObj := reflect.New(reflect.TypeOf(exampleEdge.SourceVertex).Elem())
		gs.mapper.FromVertex(edge.SourceVertex, srcVertexObj.Interface())
		destVertexObj := reflect.New(reflect.TypeOf(exampleEdge.DestinationVertex).Elem())
		gs.mapper.FromVertex(edge.DestinationVertex, destVertexObj.Interface())
		relObj := reflect.New(reflect.TypeOf(exampleEdge.Relationship).Elem())
		gs.mapper.FromEdge(edge, relObj.Interface())
		vr.SourceVertex = srcVertexObj.Interface().(GraphObject)
		vr.DestinationVertex = destVertexObj.Interface().(GraphObject)
		vr.Relationship = relObj.Interface().(GraphObject)
		vrs = append(vrs, &vr)
	}
	return vrs, nil
}

func NewGenericStore(connection core.Connection, mapper Mapper) Store {
	return &GenericStore{connection: connection, mapper: mapper}
}
