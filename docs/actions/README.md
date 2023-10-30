# Actions

Actions are executed when their parent Trigger got triggered. 
They then perform some certain task depending on the specific action.
Those can vary from shutding down the machine, removing a file or running a bash command.
**Some Actions may cause permanent damage to the system. This is intended but should be used with caution.**  

Actions can have a ```Stage``` assigned to define in which order they should run.
The lowest stage is executed first and only when finished the next stage is executed.
Actions on the same Stage run concurrently.

Actions have the following syntax:
``` json
{
  "type": "SomeAction",
  "options": { //each action defines its own options
    "firstOption": "someValue",
    "Stage": 2 //this (positive) number defines the order of multiple actions
  }
}
```

To get a list of all actions and their options from the commandline run ``` gokill -d ```
