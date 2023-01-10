package core

// Edge represents an edge within the graph
type Edge struct {
	ID                  *Identifier
	Type                string
	SourceVertexID      *Identifier
	SourceVertex        *Vertex
	DestinationVertexID *Identifier
	DestinationVertex   *Vertex
	Properties          KVMap
}

// GetId returns the identifier of the graph element as present in the underlying Graph DBMS
func (e *Edge) GetId() *Identifier {
	return e.ID
}

// GetLabel returns the set of labels associated with a Graph element
func (e *Edge) GetLabel() []string {
	return []string{e.Type}
}

// GetProperties returns the set of properties associated with the graph element
func (e *Edge) GetProperties() KVMap {
	return e.Properties
}
