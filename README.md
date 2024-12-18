# Environment Variables Object Notation
is a RedSock's environment variables parsing library (Golang)

## Advancements

- Zero external packages
- Support for composite types
- Support custom functions for parsing 

## Examples

### Test that marshal and unmarshal matreshka config
```go
package main

import (
    "go.redsock.ru/evon"
)

func Test_marshalling_env(t *testing.T) {

	ai := NewEmptyConfig()

	err := ai.Unmarshal(fullConfig)
	require.NoError(t, err)

	res := env.MarshalEnvWithPrefix("MATRESHKA", &ai)

	expected := []env.Node{
		{
			Name:  "MATRESHKA_APP_INFO_NAME",
			Value: "matreshka",
		},
		{
			Name:  "MATRESHKA_APP_INFO_VERSION",
			Value: "v0.0.1",
		},
		{
			Name:  "MATRESHKA_APP_INFO_STARTUP_DURATION",
			Value: time.Second * 10,
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_POSTGRES_HOST",
			Value: "localhost",
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_POSTGRES_PORT",
			Value: uint64(5432),
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_POSTGRES_USER",
			Value: "matreshka",
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_POSTGRES_PWD",
			Value: "matreshka",
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_POSTGRES_DB_NAME",
			Value: "matreshka",
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_POSTGRES_SSL_MODE",
			Value: "disable",
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_REDIS_HOST",
			Value: "localhost",
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_REDIS_PORT",
			Value: uint16(6379),
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_REDIS_USER",
			Value: "",
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_REDIS_PWD",
			Value: "",
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_REDIS_DB",
			Value: 0,
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_TELEGRAM_API_KEY",
			Value: "some_api_key",
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_GRPC-RSCLI-EXAMPLE_CONNECTION_STRING",
			Value: "0.0.0.0:50051",
		},
		{
			Name:  "MATRESHKA_DATA_SOURCES_GRPC-RSCLI-EXAMPLE_MODULE",
			Value: "go.redsock.ru/rscli_example",
		},
		{
			Name:  "MATRESHKA_SERVERS_REST_PORT",
			Value: uint16(8080),
		},
		{
			Name:  "MATRESHKA_SERVERS_GRPC_PORT",
			Value: uint16(50051),
		},
	}

	require.ElementsMatch(t, expected, res)
}

func Test_unmarshal_env(t *testing.T) {
	const fileIn = `MATRESHKA_APP_INFO_NAME=matreshka
MATRESHKA_APP_INFO_VERSION=v0.0.1
MATRESHKA_APP_INFO_STARTUP_DURATION=10s
MATRESHKA_DATA_SOURCES_POSTGRES_HOST=localhost
MATRESHKA_DATA_SOURCES_POSTGRES_PORT=5432
MATRESHKA_DATA_SOURCES_POSTGRES_USER=matreshka
MATRESHKA_DATA_SOURCES_POSTGRES_PWD=matreshka
MATRESHKA_DATA_SOURCES_POSTGRES_DB_NAME=matreshka
MATRESHKA_DATA_SOURCES_POSTGRES_SSL_MODE=disable
MATRESHKA_DATA_SOURCES_REDIS_HOST=localhost
MATRESHKA_DATA_SOURCES_REDIS_PORT=6379
MATRESHKA_DATA_SOURCES_REDIS_USER=redis_matreshka
MATRESHKA_DATA_SOURCES_REDIS_PWD=redis_matreshka_pwd
MATRESHKA_DATA_SOURCES_REDIS_DB=2
MATRESHKA_DATA_SOURCES_TELEGRAM_API_KEY=some_api_key
MATRESHKA_DATA_SOURCES_GRPC-RSCLI-EXAMPLE_CONNECTION_STRING=0.0.0.0:50051
MATRESHKA_DATA_SOURCES_GRPC-RSCLI-EXAMPLE_MODULE=go.redsock.ru/rscli_example
MATRESHKA_SERVERS_REST_PORT=8080
MATRESHKA_SERVERS_GRPC_PORT=50051
`
	var c AppConfig
	env.UnmarshalWithPrefix("MATRESHKA", []byte(fileIn), &c)
	expected := AppConfig{
		AppInfo: AppInfo{
			Name:            "matreshka",
			Version:         "v0.0.1",
			StartupDuration: time.Second * 10,
		},
		DataSources: DataSources{
			&data_sources.Postgres{
				Name:    "postgres",
				Host:    "localhost",
				Port:    5432,
				User:    "matreshka",
				Pwd:     "matreshka",
				DbName:  "matreshka",
				SslMode: "disable",
			},
			&data_sources.Redis{
				Name: "redis",
				Host: "localhost",
				Port: 6379,
				User: "redis_matreshka",
				Pwd:  "redis_matreshka_pwd",
				Db:   2,
			},
			&data_sources.Telegram{
				Name:   "telegram",
				ApiKey: "some_api_key",
			},
			&data_sources.GRPC{
				Name:             "grpc_rscli_example",
				ConnectionString: "0.0.0.0:50051",
				Module:           "go.redsock.ru/rscli_example",
			},
		},
		Servers: Servers{
			&servers.Rest{
				Name: "rest",
				Port: 8080,
			},
			&servers.GRPC{
				Name: "grpc",
				Port: 50051,
			},
		},
		Environment: nil,
	}
	require.Equal(t, c, expected)
}

```


###  Custom marshalers and unmarshalers
```go
package main

import (
	"go.redsock.ru/evon"
)

type Resource interface {
    // GetName - returns Name defined in config file
    GetName() string
    GetType() string
}

type DataSources []Resource
func (r *DataSources) MarshalEnv(prefix string) []env.Node {
	if prefix != "" {
		prefix += "_"
	}

	out := make([]env.Node, 0, len(*r))
	for _, resource := range *r {
		resourceName := strings.Replace(resource.GetName(), "_", "-", -1)
		out = append(out, env.MarshalEnvWithPrefix(prefix+resourceName, resource)...)
	}

	return out
}
func (r *DataSources) UnmarshalEnv(rootNode *env.Node) error {
	sources := make(DataSources, 0)
	for _, dataSourceNode := range rootNode.InnerNodes {
		name := dataSourceNode.Name

		if strings.HasPrefix(dataSourceNode.Name, rootNode.Name) {
			name = name[len(rootNode.Name)+1:]
		}

		name = strings.Replace(name, "-", "_", -1)

		dst := data_sources.GetResourceByName(name)

		env.NodeToStruct(dataSourceNode.Name, dataSourceNode, dst)
		sources = append(sources, dst)
	}

	*r = sources

	return nil
}
```