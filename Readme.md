docker network create -d bridge my-net

docker run -d \
    -p 27017:27017 \
    --name my-mongodb \
    --restart always \
    --network my-net \
    -e MONGO_INITDB_ROOT_USERNAME='admin' -e MONGO_INITDB_ROOT_PASSWORD='password' -v /usr/local/mongodata:/data/mongodb \
    mongo:4.4.10 

mongosh mongodb://admin:password@127.0.0.1:27017/admin
use admin
db.auth('admin', 'password')
db.createUser({ user: "demo",pwd: "demo", roles:[{role: "dbOwner", db:"demo"}]})

mongosh mongodb://demo:demo@127.0.0.1:27017/demo

curl https://raw.githubusercontent.com/redis/redis/6.2/redis.conf -o /data/redis/etc/redis.conf

docker run -d \
    -p 6379:6379 \
    --name my-redis \
    --restart always \
    --network my-net \
    -v /data/redis/etc:/usr/local/etc/redis \
    -v /data/redis/data:/data \
    -v /etc/localtime:/etc/localtime:ro \
    redis:6.0 \
    redis-server --appendonly yes 

redis-cli -h 127.0.0.1 -p 6379


docker build --network=my-net -t demo-app:v1 .
yes|docker image prune

docker run -d \
    -p 8000:8000 \
    --name demo-app \
    --restart always \
    --network my-net \
    demo-app:v1



mongoimport --uri=mongodb://demo:demo@127.0.0.1:27017/demo?authSource=admin -c user --file /data/workspace/demo-app/common/userData.json --jsonArray --drop
mongoimport --uri=mongodb://demo:demo@127.0.0.1:27017/demo?authSource=admin -c team --file /data/workspace/demo-app/common/teamData.json --jsonArray --drop 
mongoimport --uri=mongodb://demo:demo@127.0.0.1:27017/demo?authSource=admin -c role --file /data/workspace/demo-app/common/roleData.json --jsonArray --drop 

docker run -d --name mongo-express --network my-net -e ME_CONFIG_MONGODB_ADMINUSERNAME='admin' -e ME_CONFIG_MONGODB_ADMINPASSWORD='password' -p 8081:8081 mongo-express:0.54
docker run -d --name redisinsight --network my-net redis -p 8001:8001 redislabs/redisinsight