package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Path struct {
	Input  string
	Output string
}

type Project struct {
	ProjectName string
	Output      string
	Path        []Path
}

var PathData = []Path{
	{
		Input:  "./templates/config/config.txt",
		Output: "/config/config.go",
	},
	{
		Input:  "./templates/config/fiber.txt",
		Output: "/config/fiber.go",
	},
	{
		Input:  "./templates/constants/response_code.txt",
		Output: "/constants/response_code.go",
	},
	{
		Input:  "./templates/utils/logger/logger.txt",
		Output: "/utils/logger/logger.go",
	},
	{
		Input:  "./templates/utils/metric/prometheus.txt",
		Output: "/utils/metric/prometheus.go",
	},
	{
		Input:  "./templates/utils/responses/responses.txt",
		Output: "/utils/responses/responses.go",
	},
	{
		Input:  "./templates/entities/sample.txt",
		Output: "/entities/sample.go",
	},
	{
		Input:  "./templates/models/response.txt",
		Output: "/models/response.go",
	},
	{
		Input:  "./templates/models/sample_request.txt",
		Output: "/models/sample_request.go",
	},
	{
		Input:  "./templates/models/sample_response.txt",
		Output: "/models/sample_response.go",
	},
	{
		Input:  "./templates/mapper/sample_entity_to_sample_model.txt",
		Output: "/mapper/sample_entity_to_sample_model.go",
	},
	{
		Input:  "./templates/middleware/logger.txt",
		Output: "/middleware/logger.go",
	},
	{
		Input:  "./templates/database/connect.txt",
		Output: "/database/connect.go",
	},
	{
		Input:  "./templates/exceptions/exception.txt",
		Output: "/exceptions/exception.go",
	},
	{
		Input:  "./templates/repositories/sample_repository.txt",
		Output: "/repositories/sample_repository.go",
	},
	{
		Input:  "./templates/repositories/impl/sample_repository_impl.txt",
		Output: "/repositories/impl/sample_repository_impl.go",
	},
	{
		Input:  "./templates/services/sample_service.txt",
		Output: "/services/sample_service.go",
	},
	{
		Input:  "./templates/services/impl/sample_service_impl.txt",
		Output: "/services/impl/sample_service_impl.go",
	},
	{
		Input:  "./templates/controllers/sample_controller.txt",
		Output: "/controllers/sample_controller.go",
	},
	{
		Input:  "./templates/controllers/impl/sample_controller_impl.txt",
		Output: "/controllers/impl/sample_controller_impl.go",
	},
	{
		Input:  "./templates/routers/api_router.txt",
		Output: "/routers/api_router.go",
	},
	{
		Input:  "./templates/main.txt",
		Output: "/main.go",
	},
	{
		Input:  "./templates/.env.txt",
		Output: "/.env",
	},
}

var commands = map[string][]string{
	"init": {"mod", "init"},
	"tidy": {"mod", "tidy"},
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter project name: ")
	projectName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("An error occured while reading input. Please try again", err)
		return
	}

	projectName = strings.TrimSuffix(projectName, "\n")
	projectName = strings.ReplaceAll(projectName, " ", "-")
	fmt.Println(projectName)

	fmt.Print("Enter output path: ")
	output, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("An error occured while reading input. Please try again", err)
		return
	}

	output = strings.TrimSuffix(output, "\n")
	fmt.Println(output)

	project := Project{}
	project.ProjectName = projectName
	project.Output = output
	project.Path = PathData

	for _, path := range project.Path {
		temp := template.Must(template.ParseFiles(path.Input))

		var buf bytes.Buffer
		err := temp.Execute(&buf, project.ProjectName)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("generating file in", fmt.Sprintf("%s%s", output, path.Output))
		err = WriteFile(buf.Bytes(), fmt.Sprintf("%s%s", output, path.Output))
		if err != nil {
			log.Fatal(err)
		}
	}

	InitProject(project.Output, project.ProjectName)
	fmt.Println("done")
}

func InitProject(path, name string) {
	for key, command := range commands {
		fmt.Println("running go mod", key)
		if key == "init" {
			command = append(command, name)
		}
		cmd := exec.Command("go", command...)
		cmd.Dir = path
		_, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func WriteFile(b []byte, output string) error {
	p, err := format.Source(b)
	if err != nil {
		return err
	}

	dirName := filepath.Dir(output)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
	}

	return os.WriteFile(output, p, 0644)
}
