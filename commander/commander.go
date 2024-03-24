package commander

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/manifoldco/promptui"
)

type Directory struct {
    Name string `json:"name"`
}

type Package struct {
    Scripts map[string]string `json:"scripts"`
}

func RunCommander() {
    initialDir, err := os.Getwd()
    if err != nil {
        fmt.Println(err)
        return
    }

    for {
        if err := os.Chdir(initialDir); err != nil {
            fmt.Println(err)
            return
        }

        file, err := os.Open("dirs.json")
        if err != nil {
            fmt.Println(err)
            return
        }
        defer file.Close()

        var directories []Directory
        if err := json.NewDecoder(file).Decode(&directories); err != nil {
            fmt.Println(err)
            return
        }

        items := make([]string, len(directories))
        for i, dir := range directories {
            items[i] = dir.Name
        }

        prompt := promptui.Select{
            Label: "Select Project",
            Items: items,
        }

        _, result, err := prompt.Run()

        if err != nil {
            fmt.Printf("Prompt failed %v\n", err)
            return
        }

        fmt.Printf("You selected %v\n", result)

        home, err := os.UserHomeDir()
        if err != nil {
            fmt.Println(err)
            return
        }

        dirPath := filepath.Join(home, "GitHub", result)

        if err := os.Chdir(dirPath); err != nil {
            fmt.Println(err)
            return
        }

        fmt.Println("This is your current directory:")
        lsCmd := exec.Command("ls")
        lsOutput, err := lsCmd.CombinedOutput()
        if err != nil {
            fmt.Println(err)
            return
        }

        fmt.Println(string(lsOutput))

        packageFile, err := os.Open(filepath.Join(dirPath, "package.json"))
        if err != nil {
            fmt.Println(err)
            return
        }
        defer packageFile.Close()

        var pkg Package
        if err := json.NewDecoder(packageFile).Decode(&pkg); err != nil {
            fmt.Println(err)
            return
        }

        startScripts := make([]string, 0)
        testScripts := make([]string, 0)
        buildScripts := make([]string, 0)

        for name := range pkg.Scripts {
            if strings.HasPrefix(name, "start") {
                startScripts = append(startScripts, name)
            } else if strings.HasPrefix(name, "test") {
                testScripts = append(testScripts, name)
            } else if strings.HasPrefix(name, "build") {
                buildScripts = append(buildScripts, name)
            }
        }

        scriptNames := append(append(startScripts, testScripts...), buildScripts...)

        scriptPrompt := promptui.Select{
            Label: "Select Script",
            Items: scriptNames,
        }

        _, script, err := scriptPrompt.Run()

        if err != nil {
            fmt.Printf("Prompt failed %v\n", err)
            return
        }

        fmt.Printf("You selected %v\n", script)

        cmd := exec.Command("npm", "run", script)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Start(); err != nil {
            fmt.Println(err)
            return
        }

        stop := make(chan bool)
        done := make(chan error)

        go func() {
            reader := bufio.NewReader(os.Stdin)
            for {
                text, _ := reader.ReadString('\n')
                if strings.TrimSpace(text) == "stop" {
                    if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
                        fmt.Println("failed to kill process: ", err)
                    }
                    stop <- true
                    return
                }
            }
        }()

        go func() {
            done <- cmd.Wait()
        }()

        select {
        case <-stop:
            fmt.Println("Script stopped. Returning to script selection.")
        case err := <-done:
            if err != nil {
                fmt.Println(err)
                return
            }
        }
    }
}