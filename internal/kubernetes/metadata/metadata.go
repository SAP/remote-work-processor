package metadata

import (
	"fmt"
	"github.com/SAP/remote-work-processor/internal/utils"
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
		instanceID = utils.GetRequiredEnv(INSTANCE_ID_ENV_VAR)
	}
	return RemoteWorkProcessorMetadata{
		operatorId:  utils.GetRequiredEnv(OPERATOR_ID_ENV_VAR),
		environment: utils.GetRequiredEnv(ENVIRONMENT_ENV_VAR),
		instanceId:  instanceID,
		version:     version,
		autopiHost:  utils.GetRequiredEnv(AUTOPI_HOST_ENV_VAR),
		autopiPort:  utils.GetRequiredEnv(AUTOPI_PORT_ENV_VAR),
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
