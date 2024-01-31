package metadata

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	OPERATOR_ID_ENV_VAR = "OPERATOR_ID"
	ENVIRONMENT_ENV_VAR = "ENVIRONMENT"
	INSTANCE_ID_ENV_VAR = "INSTANCE_ID"
	AUTOPI_HOST_ENV_VAR = "AUTOPI_HOSTNAME"
	AUTOPI_PORT_ENV_VAR = "AUTOPI_PORT"
)

type RemoteWorkProcessorMetadata struct {
	operatorId  string
	environment string
	instanceId  string
	version     string
	autopiHost  string
	autopiPort  string
}

func LoadMetadata(instanceID, version string) RemoteWorkProcessorMetadata {
	if len(instanceID) == 0 {
		instanceID = getRequiredEnv(INSTANCE_ID_ENV_VAR)
	}
	return RemoteWorkProcessorMetadata{
		operatorId:  getRequiredEnv(OPERATOR_ID_ENV_VAR),
		environment: getRequiredEnv(ENVIRONMENT_ENV_VAR),
		instanceId:  instanceID,
		version:     version,
		autopiHost:  getRequiredEnv(AUTOPI_HOST_ENV_VAR),
		autopiPort:  getRequiredEnv(AUTOPI_PORT_ENV_VAR),
	}
}

func (m RemoteWorkProcessorMetadata) SessionID() string {
	return fmt.Sprintf("%s:%s:%s", m.operatorId, m.environment, m.instanceId)
}

func (m RemoteWorkProcessorMetadata) BinaryVersion() string {
	return m.version
}

func (m RemoteWorkProcessorMetadata) AutoPiHost() string {
	return m.autopiHost
}

func (m RemoteWorkProcessorMetadata) AutoPiPort() string {
	return m.autopiPort
}

func getRequiredEnv(key string) string {
	value, present := os.LookupEnv(key)
	if !present {
		log.Fatalf("failed to load remote work processor metadata: %q is missing", key)
	}
	return strings.TrimSpace(value)
}
