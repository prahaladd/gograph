package core

// Vertex represents a vertex within the graph
type Vertex struct {
	ID         *Identifier
	Labels     []string
	Properties KVMap
}

// GetId returns the identifier of the graph element as present in the underlying Graph DBMS
func (v *Vertex) GetId() *Identifier {
	return v.ID
}

// GetLabel returns the set of labels associated with a Graph element
func (v *Vertex) GetLabel() []string {
	return v.Labels
}

// GetProperties returns the set of properties associated with the graph element
func (v *Vertex) GetProperties() KVMap {
	return v.Properties
}
