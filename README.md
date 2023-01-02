# gograph

**gograph** is to graph databases what the [Go's sql](https://pkg.go.dev/database/sql) package is for SQL(and SQL-like) databases.

The project aims to provide a mechanism to interact with any graph database using a unified and minimalistic API layer for the core operations on a graph database.

Advanced functionality APIs can be built upon the core API layer. The small API surface of the core layer allows implementors to provide support for newer graph databases easily.

## Requirements
Go version 1.18 and above

## Source code layout
| Package Name   | Description   |
|---|---|
|  core | Core `struct` and type definitions. Provides the `Connection` interface to be implemented for providing core graph operations related to query execution|
| query/cypher | Utility `struct`s for building cypher queries
| neo | [Neo4J](https://neo4j.com/) specific implementation of the `Connection` interface
| integrationtests | Integration tests to validate core operations on target graph database instances

## Usage

Include the project within a Go project using the `go get` tool

`go get github.com/prahaladd/gograph`

Refer to the `integrationtests` package for examples on the API consumption.

## Contributing

You can contribute to `gograph` in many ways, all of which are equally welcome :)

* Contributing connection support for newer graph databases
* Contribute higher order graph functionality API based on the core API.
* File issues/enhancement requests for existing functionalities
* Spread awareness and excitement about the project by blogging about it
* Fork and Star the project and build something awesome :)

