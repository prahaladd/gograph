# gograph

**gograph** is to graph databases what the [Go's sql](https://pkg.go.dev/database/sql) package is for SQL(and SQL-like) databases.

The project aims to provide a mechanism to interact with any graph database using a unified and minimalistic API layer for the core operations on a graph database.

Advanced functionality APIs can be built upon the core API layer. The small API surface of the core layer allows implementors to provide support for newer graph databases easily.

## Software Requirements
* Go version 1.18 and above for code development
* Docker runtime to execute integration tests
* Connectivity to an instance of a graph database for integration tests.


## Source code layout
| Package Name   | Description   |
|---|---|
|  core | Core `struct` and type definitions. Provides the `Connection` interface to be implemented for providing core graph operations related to query execution|
| query/cypher | Utility `struct`s for building cypher queries
| omg | Object Mapped Graph layer that facilitates storage and retrieval of user defined structs as vertices and edges within the graph database
| neo | [Neo4J](https://neo4j.com/) specific implementation of the `Connection` interface
| integrationtests | Integration tests to validate core operations on target graph database instances

## Usage

Include the project within a Go project using the `go get` tool

`go get github.com/prahaladd/gograph`

The layered nature of `gograph` allows developers to choose the mechanism of interaction with the graph database.

- Directly consume the `Connection` API and express their implementation in terms of `core.Vertex` and `core.Edge` objects. `core.Vertex` and `core.Edge` objects are converted to domain specific `struct` instances via custom mapping implementations
- Consume the `omg.Store` API to map  custom user-defined `struct` instances to vertices and edges and vice versa. In the current release, `struct` instances with primitive fields have been validate.

### Connection API usage

#### Initializing a connection

Connections to graph databases can be obtained as below

```go
// obtain a reference to the Neo4J connection factory method
neo4jConnectionFactory := core.GetConnectorFactory("neo4j")

// obtain a connection to the Neo4J server
connection, err := neo4jConnectionFactory(protocol, host, realm, port, map[string]interface{}{neo.NEO4J_USER_KEY: user, neo.NEO4J_PWD_KEY: pwd}, nil)
```
Note that the process of obtaining a connection factory only involves specifying the 
database type. Hence, obtaining a connection factory for another database like [Memgraph](https://memgraph.com/) is just about the same

```go
memGraphConnectionFactory := core.GetConnectorFactory("memgraph")

// memgraph requires the same parameters for connection as Neo4J with the exception that the protocol is bolt instead of neo4j(s)
connection, err := memGraphConnectionFactory(protocol, host, realm, port, map[string]interface{}{neo.NEO4J_USER_KEY: user, neo.NEO4J_PWD_KEY: pwd}, nil)
```
It is evident now that the the only place where any database specific information is required is for establishing the initial connection.
Once a `connection` instance to the target database has been obtained, the rest of the interactions with the graph database would occur through the methods of the `Connection` API and  do not require any database specific knowledge.

#### Querying and persisting

The below snippet shows the persistence of a vertex and querying it back to the underlying database

```go
	query := "CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)"
	queryResult, err := connection.ExecuteQuery(context.Background(), query, core.Write, map[string]any{"message": "hello, world"})
	
	// attempt to read the node created in the above
	query = "MATCH (a:Greeting) return a"
	queryResult, err = suite.connection.ExecuteQuery(context.Background(), query, core.Read, nil)
```
Edge objects can be persisted and queried similarly


Refer to [Neo4J Integration Test Suite](integrationtests/neo/neo4j_integration_test.go) for a complete set of examples of working with Neo4J.

Similarly refer to [Memgraph Integration Test Suite](integrationtests/memgraph/memgraph_integration_test.go) for udnerstanding how to interact with Memgraph. 

Note how the tests share a similar structure in terms of the API consumption patterns.

Refer to the [Integration Tests](integrationtests) package for examples on the API consumption.

### Object Modeled Graph usage

The term **Object Modeled Graph (OMG)** is inspired by and is an alternative terminology to [Neo4j Object Graph Mapper](https://neo4j.com/developer/neo4j-ogm/). 
In the context of `gograph` **Object Modeled Graph** refers to a graph database agnostic mapping layer between user-defined structs to underlying graph objects and vice-versa

#### Defining OMG compatible structs

All OMG compatible structs implement the `omg.GraphObject` interface. Here is a sample implementation (sourced again from the integration tests)

```go
type person struct {
	Name string
	Age  int64
}

func (p *person) GetLabel() string {
	return "person"
}

func (p *person) GetType() omg.GraphObjectType {
	return omg.Vertex
}

type city struct {
	Name    string
	PinCode int64
}

// GetLabel returns the label associated with the graph object
func (c *city) GetLabel() string {
	return "City"
}

// GraphType() returns the type associated with the Graph object
func (c *city) GetType() omg.GraphObjectType {
	return omg.Vertex
}

type livesin struct {
	Since int64
	Area  string
}

// GetLabel returns the label associated with the graph object
func (lv *livesin) GetLabel() string {
	return "LIVES_IN"
}

// GraphType() returns the type associated with the Graph object
func (lv *livesin) GetType() omg.GraphObjectType {
	return omg.Edge
}
```
Given the above user defined `struct` instances, the below snippet shows how to persist and query these `struct` instances to the underlying graph database

```go
func (suite *Neo4JIntegrationTestSuite) TestStoreOmgStructAsVertex() {
	p := person{Name: "Tom", Age: 10}
	err := suite.store.PersistVertex(context.TODO(), &p)
	suite.NoError(err)
	p2, err := suite.store.ReadVertex(context.TODO(), &person{Name: "Tom", Age: 10})
	suite.NoError(err)
	suite.Equal(1, len(p2))
	suite.Equal(p, *(p2[0].(*person)))

}
```
Storing and retrieving an edge is on similar lines

```go
func (suite *Neo4JIntegrationTestSuite) TestStoreOmgStructsAsEdge() {
	p := person{Name: "Tom", Age: 10}
	c := city{Name: "Mumbai", PinCode: 400001}
	r := livesin{Area: "Town Hall", Since: 1982}

	vc := omg.VertexRelation{SourceVertex: &p, DestinationVertex: &c, Relationship: &r}
	err := suite.store.PersistEdge(context.TODO(), &vc)

	suite.NoError(err)

	// query the edge back
	vrs, err := suite.store.ReadEdge(context.TODO(), &vc)
	suite.NoError(err)
	suite.Equal(1, len(vrs))
	suite.Equal(vc, *vrs[0])
}
```

## Running integration tests

Running integration tests has the following pre-requisites
- Docker runtime installed on the developer desktop
- Docker container running the specific graph database service.

Most of the standard well-known graph databases are available as docker container images. Using docker facilitates a uniform test fixture executing environment.

To run integrattion tests run the following command from the project root

```bash
chmod +x run_all_tests.sh
./run_all_tests.sh
```

## Functionality Coverage
The `gograph` module has been developed and tested on the community editions of the graph databases (currently Neo4J and Memgraph). There are no APIs around additional functionality that is exposed by enterprise editions of these databases.

However for Neo4J, the ability to specify the target database is supported within the Neo4J connector shipped with this module.


## Contributing

You can contribute to `gograph` in many ways, all of which are equally welcome :)

* Contributing connection support for newer graph databases
* Contribute higher order graph functionality API based on the core API.
* File issues/enhancement requests for existing functionalities.
* Implement fixes/enhancements to the framework
* Improve Godoc and test coverage
* Spread awareness and excitement about the project by blogging about it
* Fork and Star the project and build something awesome :)

