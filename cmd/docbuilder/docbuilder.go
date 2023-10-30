package main

import (
	"fmt"
	"strings"
	"os"
	"flag"

	"unknown.com/gokill/actions"
	"unknown.com/gokill/triggers"
	"unknown.com/gokill/internal"
)

func getMarkdown(documenter internal.Documenter) string {
	var result string
	result += fmt.Sprintf("# %v\n%v\n\n", documenter.GetName(), documenter.GetDescription())

	result += fmt.Sprintf("*Example:*\n``` json\n%v\n```\n## Options:\n", documenter.GetExample())
	
	for _, opt := range documenter.GetOptions() {
		sanitizedDefault := "\"\""

		if len(opt.Default) > 0 {
			sanitizedDefault = opt.Default
		}

		result += fmt.Sprintf("### %v\n%v  \n\n*Type:* %v  \n\n*Default:* ```%v```  \n",
			opt.Name, opt.Description, opt.Type, sanitizedDefault)
	}

	return result
}

func writeToFile(path string, documenter internal.Documenter) error {
	fileName := fmt.Sprintf("%s/%s.md", path, documenter.GetName())

	f, err := os.Create(fileName)

	if err != nil {
		fmt.Println(err)
		return err
	}

	defer f.Close()

	_, err = f.WriteString(getMarkdown(documenter))

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func writeDocumentersToFiles(destination string) {
	writeFolder := func(typeName string, documenters []internal.Documenter) {
		path := fmt.Sprintf("%s/%s", destination, typeName)
		_ = os.Mkdir(path, os.ModePerm)
		for _, documenter := range documenters {
			writeToFile(path, documenter)
		}
	}

	actions := actions.GetDocumenters()
	writeFolder("actions", actions)

	triggers := triggers.GetDocumenters()
	writeFolder("triggers", triggers)
}

func printDocumentersSummary() {
	result := fmt.Sprintf("- [Triggers](triggers/README.md)\n")
	for _, trigger := range triggers.GetDocumenters() {
		result += fmt.Sprintf("\t- [%s](triggers/%s.md)\n", trigger.GetName(), trigger.GetName())
	}

	result += fmt.Sprintf("- [Actions](actions/README.md)\n")
	for _, action := range actions.GetDocumenters() {
		result += fmt.Sprintf("\t- [%s](actions/%s.md)\n", action.GetName(), action.GetName())
	}

	fmt.Print(result)
}


func main() {
	outputPath := flag.String("output", "", "path where docs/ shoud be created")

	flag.Parse()

	if *outputPath == "" {
		printDocumentersSummary()
		return
	}

	if len(*outputPath) > 1 {
		*outputPath = strings.TrimSuffix(*outputPath, "/")
	}

	writeDocumentersToFiles(*outputPath)
}
