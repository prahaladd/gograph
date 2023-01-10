package omg

type GraphObjectType int8

const (
	Vertex GraphObjectType = iota
	Edge
)

// GraphObject is a contract to be implemented by all user defined struct types
// that would be persisted to the graph database using the OMG layer.
type GraphObject interface {

	// GetLabel returns the label associated with the graph object
	GetLabel() string

	// GraphType() returns the type associated with the Graph object
	GetType() GraphObjectType
}

// GraphComponent represents a sub-graph between two vertices and the associated relationship modeled as an edge between the vertices
//
// When the connectivity contains a reference only to the source node and no references to relationship or destination node, the connectity
// represents an isolated vertex.
type VertexRelation struct {
	// SourceVertex represents the source node of a relation
	SourceVertex GraphObject
	// Relationship represents an edge between the source and destination vertices
	Relationship GraphObject
	// DestinationVertex represents the destination node of a relation
	DestinationVertex GraphObject
}
