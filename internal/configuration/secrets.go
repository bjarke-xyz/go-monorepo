package configuration

import (
	"log"
	"os"
)

func GetSwarmSecret(name string, fallback string) string {
	// Docker Swarm secrets are mounted at /run/secrets/{name}
	secretPath := "/run/secrets/" + name
	valueBytes, err := os.ReadFile(secretPath)
	if err != nil {
		log.Printf("Could not read swarm secret, using fallback: %v", err)
		return fallback
	}
	return string(valueBytes)
}
