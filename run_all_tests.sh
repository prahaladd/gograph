echo -e "Executing gograph integration tests\n"
echo -e "Cleaning up any previously running test containers...\n"
docker container kill memgraphdb
docker container kill neo4jdb
docker container kill agensgraph
docker container rm memgraphdb
docker container rm neo4jdb
docker container rm agensgraph
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
#initialize an Agensgraph container and execute integration tests
echo -e "Starting up Agensgraph instance...\n"
docker kill agensgraph
docker container rm agensgraph
docker run -d \
    --name agensgraph \
    -e POSTGRES_PASSWORD=agensgraph \
    -e PGDATA=/var/lib/postgresql/data/pgdata \
    -v /home/prahaladd/graphdb:/var/lib/postgresql/data \
    -p 5432:5432 \
    bitnine/agensgraph:v2.1.3
#Agensgraph requires a set of configuration parameters for the tests to run
export AGENS_USER=postgres
export AGENS_PWD=agensgraph
export AGENS_DB=graphdb
export AGENS_HOST=localhost
#sleep for a few seconds to allow the postgres database to initialize
sleep 10
#clean test cache and execute tests
go clean -testcache
go test ./...
docker container kill memgraphdb
docker container kill neo4jdb
docker container kill agensgraph
docker container rm memgraphdb
docker container rm neo4jdb
docker container rm agensgraph