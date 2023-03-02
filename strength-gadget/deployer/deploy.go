package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type SourceHashes struct {
	Hashes sync.Map `json:"hashes"`
}

type ApiConfig struct {
	Directory          string   `json:"directory"`
	AllowedHttpMethods []string `json:"allowedHttpMethods"`
	Hash               string   `json:"hash"`
}

type UiConfig struct {
	Directory string `json:"directory"`
}

type ConfigType struct {
	Type string `json:"type"`
}

func main() {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	ssmClient := ssm.New(sess)

	noHashesExistYetError := errors.New("ParameterNotFound: ")
	sourceHashesPath := "source_hashes"
	result, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(sourceHashesPath),
		WithDecryption: aws.Bool(true),
	})
	if err != nil && err.Error() != noHashesExistYetError.Error() {
		log.Fatalf("an unexpected error has occurred when attempting to fetch aws params: %v", err)
	}

	var existingSourceHashes SourceHashes
	if noHashesExistYetError == nil {
		err = json.Unmarshal([]byte(*result.Parameter.Value), &existingSourceHashes)
		if err != nil {
			log.Fatalf("an unexpected error has occurred when attempting to unmarshall source hashes: %v", err)
		}
	}

	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("an unexpected error has occurred when attempting to retrieve the current working directory: %v", err)
	}
	var paths []string
	// Note, this filepath.Walk is meant to be, ran against a freshly cloned repo (e.g., during CICD). You may notice performance issues when running locally due to the existence of heavy directories like node_modules.
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("an unexpected error has occurred when attempting to transverse the directory tree: %v", err)
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "deployment_config.json" {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("an unexpected error has occurred when attempting to identify directories containing deployment configs: %v", err)
	}
	if len(paths) == 0 {
		log.Fatalf("must have at least one file in the mono-repo named: \"deployment_config.json\" for something to deploy")
	}

	// Ensure all directories containing a deployment_config.json file have a unique name
	err = confirmUniqueNameOfDeploymentDirectories(paths)
	if err != nil {
		log.Fatalf("error, found a dupicate deployment directory name: %v", err)
	}

	var apiConfigs []ApiConfig
	//var uiConfigs []UiConfig
	newSourceHashes := SourceHashes{
		Hashes: sync.Map{},
	}
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(paths))
	for _, path := range paths {
		go func(p string) {
			var directoryHash string
			directoryHash, err = hashDirectory(filepath.Dir(p))
			if err != nil {
				log.Fatalf("error has happend when attempting to generate a hash for the directory %s. Error: %v", p, err)
			}
			log.Printf("directory: %s, hash: %s", p, directoryHash)

			var apiConfig *ApiConfig
			apiConfig, err = generateApiConfig(p, directoryHash)
			if err != nil {
				log.Fatalf("error has happend when attempting to generate a Terraform lambda configuration file: %v", err)
			}
			if areDirectoryContentsDirty(p, directoryHash, &existingSourceHashes) {
				var isApiConfigType *bool
				isApiConfigType, err = checkIfApiConfigType(p)
				if err != nil {
					log.Fatalf("error has happend when attempting to check if the config type is of API: %v", err)
				}

				if *isApiConfigType {

					err = generateApiArtifact(p)
					if err != nil {
						log.Fatalf("error has happend when attempting to generate an API artifact: %v", err)
					}

				} else {
					//	todo do UI deployment staging stuff
				}

				if err != nil {
					log.Fatalf("error has happend when attempting to append an Api Config the the list: %v", err)
				}
			}
			apiConfigs = append(apiConfigs, *apiConfig)
			newSourceHashes.Hashes.Store(getLowestDirectory(p), directoryHash)
			waitGroup.Done()
		}(path)
	}
	waitGroup.Wait()

	err = writeApiConfigsToFile(apiConfigs)
	if err != nil {
		log.Fatalf("error has occurred when writing API configs to a file: %v", err)
	}

	err = updateHashes(&newSourceHashes, ssmClient, sourceHashesPath)
	if err != nil {
		log.Fatalf("error has occurred when updating hashes: %v", err)
	}
}

func confirmUniqueNameOfDeploymentDirectories(paths []string) error {
	seen := make(map[string]bool)
	for _, path := range paths {
		dir := getLowestDirectory(path)
		if seen[dir] {
			return errors.New("found duplicate directory: " + dir)
		}
		seen[dir] = true
	}
	return nil
}

func updateHashes(hashes *SourceHashes, ssmClient *ssm.SSM, sourceHashesPath string) error {

	// Convert the sync.Map to a regular map
	convertedMap := make(map[string]string)
	hashes.Hashes.Range(func(key, value interface{}) bool {
		convertedMap[key.(string)] = value.(string)
		return true
	})

	data, err := json.Marshal(convertedMap)
	if err != nil {
		return fmt.Errorf("error happend when attempting to marshal source hashes: %v", err)
	}

	result := string(data)
	parameterTypeString := ssm.ParameterTypeString
	overwrite := true
	input := &ssm.PutParameterInput{
		Name:      &sourceHashesPath,
		Type:      &parameterTypeString,
		Value:     &result,
		Overwrite: &overwrite,
	}
	_, err = ssmClient.PutParameter(input)
	if err != nil {
		return fmt.Errorf("error happend when attempting to record hashes to aws param store: %v", err)
	}

	return nil
}

func generateApiArtifact(path string) error {
	outputDir := fmt.Sprintf("/tmp/lambdas/%s", getLowestDirectory(path))
	cmd := fmt.Sprintf("go mod tidy && GOOS=linux GOARCH=amd64 go build -o %s/app main.go", outputDir)
	c := exec.Command("sh", "-c", cmd)
	c.Dir = filepath.Dir(path)
	output, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error, when building artifact and placing in tmp dir. Output: %s. Error: %v", string(output), err)
	}

	c = exec.Command("sh", "-c", "zip lambda-handler.zip app")
	c.Dir = outputDir
	output, err = c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error, when zipping artifact. Output: %s. Error: %v", string(output), err)
	}
	return nil
}

func writeApiConfigsToFile(result []ApiConfig) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("error has happend when attempting to marshall the configurations into a json string: %v", err)
	}

	file, err := os.Create("/tmp/lambda_configs.json")
	if err != nil {
		return fmt.Errorf("error has happend when attempting to create the file: %v", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error has happened when attempting to write to the file: %v", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("error has happened when attempting to close the file: %v", err)
	}
	return nil
}

func generateApiConfig(path string, hash string) (*ApiConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("error has happend when attempting to open the config file: %v", err)
	}
	var apiConfig ApiConfig
	base := getLowestDirectory(path)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&apiConfig)
	if err != nil {
		return nil, fmt.Errorf("error has occurred when decoding the api %s deployment_config.json file: %v", base, err)
	}
	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("an error has happend when attempting to close the file: %v", err)
	}

	return &ApiConfig{
		Directory:          base,
		AllowedHttpMethods: apiConfig.AllowedHttpMethods,
		Hash:               hash,
	}, nil
}

func getLowestDirectory(path string) string {
	dir := filepath.Dir(path)
	return filepath.Base(dir)
}

func checkIfApiConfigType(p string) (*bool, error) {
	file, err := os.Open(p)
	if err != nil {
		log.Fatalf("error has happend when attempting to open the config file: %v", err)
	}
	var configType ConfigType
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configType)
	if err != nil {
		return nil, fmt.Errorf("error has occurred when attempting to decode the config: %v", err)
	}
	result := configType.Type == "API"
	return &result, nil
}

func areDirectoryContentsDirty(p string, directoryHash string, existingSourceHashes *SourceHashes) bool {
	value, ok := existingSourceHashes.Hashes.Load(filepath.Base(p))
	if !ok {
		return true
	}
	return directoryHash != value
}

func hashDirectory(dir string) (string, error) {
	h := sha256.New()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error occurred when transversing file directory for the purpose of hashing its contents: %v", err)
		}

		if info.IsDir() {
			return nil
		}

		fileHash, err := hashFile(path)
		if err != nil {
			return fmt.Errorf("error occurred when attempting to hash a file: %v", err)
		}

		h.Write(fileHash)

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error has happend when attempting to hash all files in the directory: %v", err)
	}
	result := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return result, nil
}

func hashFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error occurred when attempting to open the file: %v", err)
	}

	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("error occurred when attempting to perform the hash operation: %v", err)
	}
	err = f.Close()
	if err != nil {
		return nil, fmt.Errorf("error occurred when attempting to close the file: %v", err)
	}

	return h.Sum(nil), nil
}