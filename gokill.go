package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"unknown.com/gokill/triggers"
)

func main() {
	configFile := flag.String("c", "", "path to config file")
	flag.Parse()

	if *configFile == "" {
		fmt.Println("No config file given. Use --help to show usage.")
		//return
	}

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

	var f []triggers.KillSwitchConfig
	err := json.Unmarshal(b, &f)

	if err != nil {
		fmt.Println(err)
		return
	}

	trigger, err := triggers.NewTrigger(f[0])

	if err != nil {
		fmt.Println(err)
		return
	}

	trigger.Listen()

	//stagedActions := actions.StagedActions{make(chan bool), 0, []actions.Stage{}}

	//stageOne := actions.Stage{[]actions.Action{
	//	actions.Printer{"first action\n", stagedActions.ActionChan},
	//	actions.Printer{"second actiloo\n", stagedActions.ActionChan},
	//	actions.TimeOut{stagedActions.ActionChan},
	//}}

	//stageTwo := actions.Stage{[]actions.Action{
	//	actions.Printer{"third action\n", stagedActions.ActionChan},
	//	actions.TimeOut{stagedActions.ActionChan},
	//}}

	//stageThree := actions.Stage{[]actions.Action{
	//	actions.Printer{"four action\n", stagedActions.ActionChan},
	//	actions.Printer{"five action\n", stagedActions.ActionChan},
	//	actions.Printer{"six action\n", stagedActions.ActionChan},
	//}}

	//stagedActions.Stages = []actions.Stage{stageOne, stageTwo, stageThree}

	//timeOut := triggers.NewTimeOut(2*time.Second, stagedActions)
	//timeOut.Listen()
}
