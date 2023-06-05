package metadata

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	OPERATOR_ID_ENV_VAR = "OPERATOR_ID"
	ENVIRONMENT_ENV_VAR = "ENVIRONMENT"
	INSTANCE_ID_ENV_VAR = "INSTANCE_ID"
	IMAGE_ENV_VAR       = "IMAGE"

	IMAGE_TAG_SEPARATOR = ":"
)

var (
	once     sync.Once
	Metadata RemoteWorkProcessorMetadata
)

type RemoteWorkProcessorMetadata struct {
	operatorId  string
	environment string
	instanceId  string
	image       string
}

func InitRemoteWorkProcessorMetadata() {
	operatorId, err := getEnv(OPERATOR_ID_ENV_VAR)
	if err != nil {
		log.Fatal(err)
	}

	environment, err := getEnv(ENVIRONMENT_ENV_VAR)
	if err != nil {
		log.Fatal(err)
	}

	instanceId, err := getEnv(INSTANCE_ID_ENV_VAR)
	if err != nil {
		log.Fatal(err)
	}

	image, err := getEnv(IMAGE_ENV_VAR)
	if err != nil {
		log.Fatal(err)
	}

	once.Do(func() {
		Metadata = RemoteWorkProcessorMetadata{
			operatorId:  operatorId,
			environment: environment,
			instanceId:  instanceId,
			image:       image,
		}
	})
}

func (p RemoteWorkProcessorMetadata) Id() string {
	return fmt.Sprintf("%s:%s:%s", p.operatorId, p.environment, p.instanceId)
}

func (p RemoteWorkProcessorMetadata) BinaryVersion() string {
	return p.image[strings.LastIndex(p.image, IMAGE_TAG_SEPARATOR)+1:]
}

func getEnv(key string) (string, error) {
	h, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("failed to create remote work processor id, because %s must be set", key)
	}
	return strings.TrimSpace(h), nil
}
