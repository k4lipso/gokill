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

	var result string

	writeOptions := func(documenters []internal.Documenter) {
		for _, act := range documenters {
			result += fmt.Sprintf("\n### %v\nDescription: %v  \nValues:\n", act.GetName(), act.GetDescription())

			for _, opt := range act.GetOptions() {
				result += fmt.Sprintf("- Name: **%v**\n\t- Type: %v\n\t- Descr: %v\n\t- Default: %v\n",
					opt.Name, opt.Type, opt.Description, opt.Default)
				result += "\n\n"
			}
		}
	}

	result = "# Available Triggers:\n\n"
	writeOptions(triggers.GetDocumenters())
	result += "\n\n# Available Actions:\n\n"
	writeOptions(actions)

	return result
}

func main() {
	configFilePath := flag.String("c", "", "path to config file")
	showDoc := flag.Bool("d", false, "show doc")
	testRun := flag.Bool("t", false, "test run")

	flag.Parse()

	if *showDoc {
		fmt.Print(GetDocumentation())
		return
	}

	if *configFilePath == "" {
		fmt.Println("No config file given. Use --help to show usage.")
		return
	}

	actions.TestRun = *testRun

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