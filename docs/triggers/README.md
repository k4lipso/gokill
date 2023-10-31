# Triggers

Triggers wait for certain events and execute the actions defined for them.
There are different Triggers for different use cases.
For example ```UsbDisconnect``` is triggered when a certain Usb Drive is unplugged.
If you want your actions to be triggered when an ethernet cable is pulled use ```EthernetDisconnect``` instead.

Triggers have the following syntax:
``` json
{
  "type": "SomeTrigger",
  "name": "MyFirstTrigger",
  "options": { //each trigger defines its own options
    "firstOption": 23,
    "secondOption": "foo"
  },
  "actions": [] //list actions that should be executed here
}
```
