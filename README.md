# go-common


go-common is a collection of useful utilities for Go project. Include:
- Configuration file loader
- HTTP server
- gRPC server
- Logging
- Monitoring
- Tracing
- MySQL
- Redis
- Kafka

All utilities is written as [Fx](https://github.com/uber-go/fx) module .But, Some of them can be use without `Fx`

## Installation
```shell
go get go-common
```

## Quickstart

### Configuration file loader
By default, configuration file are stored in `CONFIG_PATH` env variable. If `CONFIG_PATH` is empty, it'll read  `config/local.yaml`. Configuration is read and stored in `viper`

Usage:
```go
import (
    "go-common/modulefx/config"
    "go.uber.org/fx"
    "github.com/spf13/viper"
)

func main() {
    //use as Fx module
    ...
    app := fx.New(
    	config.Module,
        ...
    )
    app.Run()

    //or without Fx
    config.LoadConfiguration()
}

//example get config
func getServiceName() string {
    return viper.Get("service.name")
}
```

### HTTP server
We use **[Gin](https://github.com/gin-gonic/gin)** for http server. This Http server included health check (`/info` and `/health`) and prometheus metrics (`/metrics`) by default.

Configuration:
| Key  | Type  | Explain  |  Example |
|---|---|---|---|
|  http.port | int  | server's port  | 8080  |


Usage:
```go
import (
    "github.com/gin-gonic/gin"
    "github.com/nmtri1912/go-common/modulefx/httpserver"
    "go.uber.org/fx"
)

func main() {
    ...
    //Fx module
    app := fx.New(
        fx.Provide(newGinHandler),
    	httpserver.Module,
        ...
    )
    app.Run()
}

func newGinHandler() *gin.Engine {
    r := gin.Default()
    ...
    //write controller
    return r
}
```

We also provide *simple http server*, which contains only health check (`/info` and `/health`) and prometheus metrics (`/metrics`). This is for service use gRPC as the primary protocol. 

Usage:
```go
import (
    "github.com/gin-gonic/gin"
    httpserver "go-common/modulefx/simpleserver"
    "go.uber.org/fx"
)

func main() {
    ...
    //Fx module
    app := fx.New(
    	httpserver.Module,
        ...
    )
    app.Run()
}
```

### gRPC server
gRPC server included health check service and recover, tracing, logging, authentication interceptor.
Configuration:
| Key  | Type  | Explain  |  Example |
|---|---|---|---|
|  grpc.port | int  | server's port  | 9090  |

Usage:
```go
import (
    "go-common/modulefx/grpcserver"
    "go.uber.org/fx"
)

func NewGrpcService(gameInfoService gameinfoservice.GameInfoService,
	dbConfig dbconfig.DBConfig, cache cache.Cache) *grpcserver.GrpcService {
	service := &GameInfoGrpcServerImpl{}
	clients := viper.GetStringMapString("api-client-key.client-key-map")
	methodClients := viper.GetStringMapStringSlice("api-client-key.api-clients-map")

	return &grpcserver.GrpcService{
		ServiceDesc:          &gameinfo_grpc.ToroCrushGameInfo_ServiceDesc,
		ServiceImpl:          service,
		Clients:              clients,
		AllowedMethodClients: methodClients,
	}
}


func main() {
    ...
    //Fx module
    app := fx.New(
        fx.Provide(NewGrpcService),
    	grpcserver.Module,
        ...
    )
    app.Run()
}
```


### Logging
We use **Zap** for logging and wrap it to log trace_id and span_id (if present).

Usage:
```go
import (
    "context"
    "go-common/modulefx/logger"
    log	"go-common/pkg/logger"
	
    "go.uber.org/zap"
    "go.uber.org/fx"
)

func main() {
    ...
    //Fx module
    app := fx.New(
    	logger.Module,
        ...
    )
    app.Run()
}

//example
func httpHandler(ctx context.Context) {
    ...
    //log without trace_id and span_id field
    log.L().Info("Hello",zap.String("name","ce"))

    //log with trace_id and span_id field
    log.Ctx(ctx).Info("Hello",zap.String("name","ce"))
    ...
}
```

### Monitor
Export Promethus metrics

Usage:
```go
import (
    "go-common/modulefx/monitor"
    "go.uber.org/fx"
)

func main() {
    ...
    //Fx module
    app := fx.New(
    	monitor.Module,
        ...
    )
    app.Run()
    
    //without Fx
    monitor.InitMonitorMetrics()
}

//example
const DOMAIN = "example"
const GRPC_TYPE = "grpc"

func GetUserResults() {
    var err error = nil
    defer func(start time.Time) {
        go monitor.RecordMetrics(DOMAIN,GRPC_TYPE, "GetUserResults", start, err)
    }(time.Now())
    ...
    //your code here
} 
```

### Distributed Tracing
We use **[OpenTelemetry](https://opentelemetry.io/docs/instrumentation/go/)** for distrubted tracing. Which is a Zipkin client.
Configuration:
| Key  | Type  | Explain  |  Example |
|---|---|---|---|
|  zipkin.url | string  | zipkin server  | http://zipkin.vn:9411/api/v2/spans  |
|  zipkin.rate | float  | example rate  | 0.1  |
|  debug.tracing | boolean  | true: log all tracing events. Default is false | true |
|  service.name | string  | service's name for tracing data | ce-game-info |


Usage:
```go
import (
    "go-common/modulefx/tracing"
    "go.uber.org/fx"
)

func main() {
    ...
    //Fx module
    app := fx.New(
    	tracing.Module,
        ...
    )
    app.Run()
    ...
}
```

### MySQL
We use **[Sqlx](https://github.com/jmoiron/sqlx)** for MySQL client.

Configuration:
| Key  | Type  | Explain  |  Example |
|---|---|---|---|
|  mysql.username | string  | mysql's user name  | dev  |
|  mysql.password | string  | mysql's password  | youcantseeme  |
|  mysql.url | string  | server ip:port | localhost:3306 |
|  mysql.schema | string  | schema's name | ce_game_info |


Usage:
```go
import (
    "go-common/modulefx/mysql"
    "github.com/jmoiron/sqlx"
    "go.uber.org/fx"
)

func main() {
    ...
    //Fx module
    app := fx.New(
    	mysql.Module,
        ...
    )
    app.Run()
    ...
}

//example
func getUserResultsFromDB(db *sqlx.DB) {
    ...
    rows, err := r.db.QueryxContext(ctx, selectStmt, userId)
    ...
}
```


### Redis
We use **[go-redis](https://github.com/go-redis/redis)** for redis client.

Configuration:
| Key  | Type  | Explain  |  Example |
|---|---|---|---|
|  redis.addresses | string  | redis server addresses  | localhost:6357  |
|  redis.monitor-hook | boolean  | true: enable monitoring hook. Default is false | true |


Usage:
```go
import (
    "go-common/modulefx/redis"
    redis_client "go-common/pkg/redis"

    "go.uber.org/fx"
)

func main() {
    ...
    //Fx module
    app := fx.New(
    	redis.Module,
        ...
    )
    app.Run()
    ...
}

//example
func getUserResultsFromCache(rdb *redis_client.Cache) {
    ...
    rdb.Get(ctx,"KEY")
    ...
}
```


### Kafka
We use **[Sarama](https://github.com/Shopify/sarama)** for Kafka client.


Producer:
```go
import (
    kafkaProducer "go-common/pkg/kafka/producer"

    "go.uber.org/fx"
)

func main() {
    ...
    //Fx module
    app := fx.New(
    	fx.Provide(NewCeKafkaTemplate),
        ...
    )
    app.Run()
    ...
}

func NewCeKafkaTemplate(lifecycle fx.Lifecycle) kafkaProducer.KafkaProducer {
    ...
    return kafkaProducer.NewKafkaTemplate(lifecycle, brokers)
}

//example
func sendMessage(producer kafkaProducer.KafkaProducer) {
    ...
    producer.ProduceCtx(ctx, topic, key, message)
    ...
}
```

Consumer:
```go
import (
    kafkaConsumer "go-common/pkg/kafka/consumer"
    log	"go-common/pkg/logger"

    "github.com/Shopify/sarama"
    "github.com/spf13/viper"
    "go.uber.org/fx"
)

func main() {
    ...
    //Fx module
    app := fx.New(
    	fx.Invoke(StartLogConsumer),
        ...
    )
    app.Run()
    ...
}

func StartLogConsumer(lifecycle fx.Lifecycle) {
    ...
    consumer := kafkaConsumer.NewKafkaConsumer(brokers, groupId, []string{topic}, consumerHandler, numWorker)
    lifecycle.Append(fx.Hook{
    	OnStart: func(ctx context.Context) error {
    		go consumer.Start()
    		return nil
    	},
    	OnStop: func(ctx context.Context) error {
    		consumer.Close()
    		return nil
    	}},
    )
    ...
}

func consumerHandler(message *sarama.ConsumerMessage) {
    ctx, span := kafkaConsumer.ExtractSpanAndTracingContext(message)
    defer span.End()
    log.Ctx(ctx).Info("received log", zap.String("message", string(message.Value))   
}
```