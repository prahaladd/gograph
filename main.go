package main

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/prahaladd/gograph/core"
	"github.com/prahaladd/gograph/neo"
)

func main() {
	fmt.Println("Hello world. Ready to Go Neo!!")
	queryNeo4J(context.Background())
}

func queryNeo4J(ctx context.Context) {

	auth := make(map[string]interface{})
	auth[neo.NEO4J_USER_KEY] = "neo4j"
	auth[neo.NEO4J_PWD_KEY] = "test"
	// neo4jConnection, err := neo.NewConnection("neo4j", "localhost", "", nil, auth, nil)
	neo4jConnection, err := core.GetConnection("neo4j", "neo4j", "localhost", "", nil, auth, nil)
	
	if err != nil {
		fmt.Println("Error connecting to neo4j ", err)
		return
	}

	query := "CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)"
	queryResult, err := neo4jConnection.ExecuteQuery(context.Background(), query, core.Write, map[string]any{"message": "hello, world"})
	if err != nil {
		fmt.Println("Error connecting to neo4j ", err)
		return
	}
	for _, row := range queryResult.Rows {
		for k, v := range row {
			fmt.Printf("%s == %v\n", k, v)
		}
	}
	// attempt to read the node created in the above
	query = "MATCH (a:Greeting) return a"
	queryResult, err = neo4jConnection.ExecuteQuery(context.Background(), query, core.Read, nil)
	if err != nil {
		fmt.Println("Error connecting to neo4j ", err)
		return
	}
	for _, row := range queryResult.Rows {
		for k, v := range row {
			fmt.Printf("%s == %v\n", k, v)
		}
	}

	vertices, err := neo4jConnection.QueryVertex(context.Background(), "Person", map[string]interface{}{"name": "Tom"}, nil)
	if err != nil {
		fmt.Println("Error connecting to neo4j ", err)
		return
	}
	for _, v := range vertices {
		fmt.Println("Found vertex : " + fmt.Sprintf("%v", *v))
	}
}
func helloWorld(ctx context.Context, uri, username, password string) (string, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))

	if err != nil {
		return "", err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	greeting, err := session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)",
			map[string]any{"message": "hello, world"})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return "", err
	}

	return greeting.(string), nil
}
