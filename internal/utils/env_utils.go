package utils

import (
	"log"
	"os"
	"strings"
)

func GetRequiredEnv(key string) string {
	value, present := os.LookupEnv(key)
	if !present {
		log.Fatalln("failed to load remote work processor metadata: missing", key)
	}
	return strings.TrimSpace(value)
}
