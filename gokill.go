package main

import (
	"encoding/json"
	"flag"
	"fmt"

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
	b := []byte(`

[
    {
        "type": "TimeOut", 
        "name": "custom timeout",
        "options": {
            "duration": 5
        },
        "actions": [
            {
                "type": "TimeOut",
                "options": {
                    "duration": 4
                },
                "stage": 1
            },
            {
                "type": "Print",
                "options": {
                    "message": "shutdown -h now"
                },
                "stage": 1
            },
            {
                "type": "Print",
                "options": {
                    "message": "shutdown -h now"
                },
                "stage": 2 
            },
            {
                "type": "TimeOut",
                "options": {
                    "duration": 4
                },
                "stage": 5
            },
            {
                "type": "Print",
                "options": {
                    "message": "shutdown -h now"
                },
                "stage": 4
            },
            {
                "type": "Print",
                "options": {
                    "message": "shutdown -h now"
                },
                "stage": 7 
            }
        ]
    }
]
	`)

	configFile := flag.String("c", "", "path to config file")
	showDoc := flag.Bool("d", false, "show doc")
	flag.Parse()

	if *showDoc {
		fmt.Print(GetDocumentation())
		return
	}

	if *configFile == "" {
		fmt.Println("No config file given. Use --help to show usage.")
		//return
	}

	var f []internal.KillSwitchConfig
	err := json.Unmarshal(b, &f)

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
