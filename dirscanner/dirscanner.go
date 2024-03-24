package dirscanner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type DirEntry struct {
    Name string `json:"name"`
}

// ScanAndCleanGitHubDirs is the only function main.go will call. It encapsulates all operations.
func ScanAndCleanGitHubDirs() {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        fmt.Println("Error getting home directory:", err)
        return
    }

    startDir := filepath.Join(homeDir, "GitHub")
    jsonFileName := "dirs.json"
    substringsToRemove := []string{"node_modules", ".git", "dist", "build"}

    // Combine directory scanning and JSON generation in one step
    dirs, err := findDirectoriesContainingFiles(startDir)
    if err != nil {
        fmt.Println("Error finding directories:", err)
        return
    }

    // Remove unwanted directories from the list
    cleanedDirs := cleanDirs(dirs, substringsToRemove)

    // Write the cleaned list to a JSON file
    if err := writeDirsJSON(cleanedDirs, jsonFileName); err != nil {
        fmt.Println("Error writing dirs.json:", err)
        return
    }

    fmt.Println("Directories have been written to dirs.json successfully.")
}

func findDirectoriesContainingFiles(startDir string) ([]DirEntry, error) {
    var dirs []DirEntry
    err := filepath.Walk(startDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() && (containsFile(path, "package.json") || containsFile(path, "go.mod")) {
            relPath, err := filepath.Rel(startDir, path)
            if err != nil {
                return err
            }
            dirs = append(dirs, DirEntry{Name: filepath.ToSlash(relPath)})
        }
        return nil
    })
    return dirs, err
}

func cleanDirs(dirs []DirEntry, substringsToRemove []string) []DirEntry {
    var cleanedDirs []DirEntry
    for _, dir := range dirs {
        include := true
        for _, substr := range substringsToRemove {
            if strings.Contains(dir.Name, substr) {
                include = false
                break
            }
        }
        if include {
            cleanedDirs = append(cleanedDirs, dir)
        }
    }
    return cleanedDirs
}

func writeDirsJSON(dirs []DirEntry, jsonFileName string) error {
    jsonData, err := json.MarshalIndent(dirs, "", "    ")
    if err != nil {
        return fmt.Errorf("error marshaling directories: %w", err)
    }
    return ioutil.WriteFile(jsonFileName, jsonData, 0644)
}

func containsFile(dirPath, fileName string) bool {
    filePath := filepath.Join(dirPath, fileName)
    _, err := os.Stat(filePath)
    return !os.IsNotExist(err)
}
