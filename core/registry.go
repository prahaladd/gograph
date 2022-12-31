package core

type ConnectorFactory func(protocol, host, realm string, port *int32, auth, options map[string]interface{}) (Connection, error)

var graphConnectorRegistry map[string]ConnectorFactory = make(map[string]ConnectorFactory)

func RegisterConnectorFactory(graphType string, factoryFunc ConnectorFactory) {
	graphConnectorRegistry[graphType] = factoryFunc
}

func GetConnectorFactory(graphType string) ConnectorFactory {
	if factory, ok := graphConnectorRegistry[graphType]; ok {
		return factory
	}
	return nil
}
