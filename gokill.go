package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"unknown.com/gokill/actions"
	"unknown.com/gokill/internal"
	"unknown.com/gokill/triggers"
)

func GetDocumentation() string {
	actions := actions.GetDocumenters()

	result := "Available Actions:\n\n"
	lineBreak := "----------------------------"

	writeOptions := func(documenters []internal.Documenter) {
		for _, act := range documenters {
			result += lineBreak
			result += fmt.Sprintf("\nName: %v\nDescription: %v\nValues:\n", act.GetName(), act.GetDescription())

			for _, opt := range act.GetOptions() {
				result += fmt.Sprintf("\tName: %v\n\tType: %v\n\tDescr: %v\n\tDefault: %v\n",
					opt.Name, opt.Type, opt.Description, opt.Default)
				result += lineBreak + "\n\n"
			}
		}
	}

	writeOptions(actions)
	result += "\n\nAvailable Triggers:\n\n"
	writeOptions(triggers.GetDocumenters())

	return result
}

func main() {
	configFilePath := flag.String("c", "", "path to config file")
	showDoc := flag.Bool("d", false, "show doc")
	flag.Parse()

	if *showDoc {
		fmt.Print(GetDocumentation())
		return
	}

	if *configFilePath == "" {
		fmt.Println("No config file given. Use --help to show usage.")
		return
	}

	configFile, err := os.ReadFile(*configFilePath)

	if err != nil {
		fmt.Println("Error loading config file: ", err)
		return
	}

	var f []internal.KillSwitchConfig
	err = json.Unmarshal(configFile, &f)

	if err != nil {
		fmt.Println(err)
		return
	}

	var triggerList []triggers.Trigger
	for _, cfg := range f {
		trigger, err := triggers.NewTrigger(cfg)

		if err != nil {
			fmt.Println(err)
			return
		}

		trigger.Listen() //TODO: not block here
		triggerList = append(triggerList, trigger)
	}

}
