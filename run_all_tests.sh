echo -e "Executing gograph integration tests\n"
echo -e "Starting up Neo4J instance...\n"
docker pull neo4j:latest
docker run --name neo4jdb -d --publish=7474:7474 --publish=7687:7687 --volume=$HOME/neo4j/data:/data --env=NEO4J_AUTH=none  neo4j
echo -e "Starting up Memgraph instance...\n"
docker run --name memgraphdb -d -p 7697:7687 -p 7444:7444 -p 3000:3000 memgraph/memgraph
export NEO4J_USER=neo4j
export NEO4J_PWD=neo4j
# the Memgraph database exposes the bolt protocol port on 7697 as per the arguments
# in the above docker run command
export MG_PORT=7697
go clean -testcache
go test ./...
docker container kill memgraphdb
docker container kill neo4jdb
docker container rm memgraphdb
docker container rm neo4jdb