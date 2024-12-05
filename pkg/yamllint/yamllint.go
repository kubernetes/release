package yamllint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

// validateYAML reads and parses the YAML file at the given path.
func validateYAML(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf("Error reading %s:\n%s", filePath, err.Error())
	}

	var content interface{}
	if err := yaml.UnmarshalStrict(data, &content); err != nil {
		return fmt.Sprintf("Error in %s:\n%s", filePath, err.Error())
	}

	return "Valid"
}

func FindYAMLFiles(dirPath string) ([]string, error) {
	var yamlFiles []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})

	return yamlFiles, err
}

// validateYAMLFilesInDirectory finds and validates all YAML files in the given directory.
func validateYAMLFilesInDirectory(dirPath string) {
	files, err := FindYAMLFiles(dirPath)
	if err != nil {
		fmt.Printf("Error finding YAML files in %s:\n%s", dirPath, err.Error())
		return
	}

	if len(files) == 0 {
		fmt.Println("No YAML files found.")
		return
	}
	invalidFiles := []string{}
	for _, file := range files {
		result := validateYAML(file)
		if result != "Valid" {
			invalidFiles = append(invalidFiles, result)
		}
	}

	if len(invalidFiles) > 0 {
		fmt.Println("YAML Validation Failed:\n")
		for _, errorMsg := range invalidFiles {
			fmt.Println(errorMsg)
			fmt.Println("\n" + strings.Repeat("-", 40) + "\n")
		}
	} else {
		fmt.Println("All YAML files are valid.")
	}
}

// func main() {
// 	if len(os.Args) < 2 {
// 		fmt.Println("No directory specified.")
// 		os.Exit(1)
// 	}

// 	dirPath := os.Args[1]
// 	validateYAMLFilesInDirectory(dirPath)
// }
