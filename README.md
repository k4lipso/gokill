# gokill

gokill is aimed at activists, journalists and others that need to protect their data against access under all circumstances.
gokill falls under the category of anti-forensic tools, helping you to protect yourself against repression.
It is built for worst case scenarios when intruders physical gaining access to a device.
In such heated situations gokill helps you automatically executing tasks like:
- locking the screen 
- notify someone
- deleting data
- encrypting partitions 
- destroying encrypted partitions
- and many more

the tasks gokill executes could be done by hand using shellscripts, cronjobs, daemons ect.
but that means everyone needs to figure it out for themselves, and eventually make mistakes.
the idea of gokill is to provide a wide variarity of possibilities out of the box while making sure they are well tested.


gokill aims to be highly configurable and easily extendable.

'gokill' is a tool that completes some actions when a certain event occurs.
actions can vary from shuting down the machine to sending mails over erasing data.
actions can be triggert by certain conditions like specific outcomes of unix
comands or not having internet connection.

actions and triggers should be easy to extend and handled like plugins. they
also should be self documenting.
every action and trigger should be testable at anytime as a 'dry-run'.
actions can have a 'stage' defined. the lowest stage is started first,
and only when all actions on that stage are finished next stage is triggered

gokill should run as daemon. config should be read from /etc/somename/config.json

## Config Example
``` json
[ //list of triggers
    {
		"type": "UsbDisconnect",
		"name": "First Trigger",
		"options": {
			"deviceId": "ata-Samsung_SSD_860_EVO_1TB_S4AALKWJDI102",
			"waitTillConnected": true //only trigger when usb drive was actually attached before
		}
        "actions": [ //list of actions that will be executed when triggered
            {
                "name": "unixCommand",
                "options": {
                    "command": "shutdown -h now"
                },
                "stage": 2 // defines the order in which actions are triggered.
            },
            {
                "type": "sendMail",
                "options": {
                    "smtpserver": "domain.org",
                    "port": 667,
                    "recipients": [ "mail1@host.org", "mail2@host.org" ],
                    "message": "kill switch was triggered",
                    "attachments": [ "/path/atachments" ],
                    "pubkeys": "/path/to/keys.pub"
                },
                "stage": 1 //this event is triggered first, then the shutdown
            },
        ]
    },
    {
		"type": "EthernetDisconnect",
		"name": "Second Trigger",
		"options": {
			"interfaceName": "eth0",
		}
        "actions": [
            {
                "name": "unixCommand",
                "options": {
                    "command": "env DISPLAY=:0 sudo su -c i3lock someUser"
                }
            }
        ]
    }
]
```

## nix support

gokill enjoys full nix support. gokill exposes a nix flakes that outputs a gokill package, a nixosModule and more.
That means you can super easily incorporate gokill into your existing nixosConfigurations. 
Here is a small example config:

``` nix
{
  services.gokill.enable = true;
  services.gokill.triggers = [
    {
      type = "EthernetDisconnect";
      name = "MainTrigger";
      options = {
        interfaceName = "eth1";
      };
      actions = [
        {
            type = "Command";
            options = {
                command = "echo hello world";
            };
            stage = 1;
        }
      ];
    }
  ];
}
```

This will automatically configure and enable a systemd running gokill as root user in the background

## actions
- [x] shutdown
- [ ] wipe ram
- [ ] send mail
- [ ] delete data
- [ ] shred area
- [x] random command
- [ ] wordpress post
- [ ] ipfs command
- [ ] [buskill 'triggers'](https://github.com/BusKill/awesome-buskill-triggers)
    - [x] [lock-screen](https://github.com/BusKill/buskill-linux/tree/master/triggers)
    - [x] shutdown
    - [ ] luks header shredder
    - [ ] veracrypt self-destruct

## Triggers
- [ ] no internet
- [x] [pull usb stick](https://github.com/deepakjois/gousbdrivedetector/blob/master/usbdrivedetector_linux.go)
- [x] ethernet unplugged
- [ ] power adapter disconnected
- [ ] unix command
- anyOf
    - trigger wrapper containing many triggers and fires as soon as one of them
      is triggered
- allOf
- [ ] ipfs trigger
