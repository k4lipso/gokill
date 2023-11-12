# gokill

gokill is a [software dead man's switch](https://en.wikipedia.org/wiki/Dead_man%27s_switch#Software) that empowers users to configure various events. If these events occur, they trigger predefined actions.
The tool is designed for activists, journalists, and individuals who require robust protection for their data, ensuring it remains inaccessible under any circumstances. It belongs to the category of anti-forensic tools, providing a means to safeguard against potential repression. It is specifically crafted for worst-case scenarios, such as when intruders gain physical access to a device. In these intense situations, gokill can automatically perform tasks to enhance your security. Those could be:
- locking the screen 
- sending chat messages
- deleting data
- encrypting partitions 
- destroying encrypted partitions
- ect

#### documentation
A full list of Triggers and Actions with all their configuration options can be found here: 

## usage
If you use NixOS gokill can easily be integrated into your system configuration - scroll down for more info on that.  

For all other linux distributions gokill currently needs to be built and setup manually. This is supposed to change.
Iam currently working/researching on publishing gokill as [ppa](https://help.launchpad.net/Packaging/PPA) and as snap.
If you have other recommendations let me know.  


``` bash
# Clone the gokill repository
git clone https://github.com/k4lipso/gokill
cd gokill

# Build gokill - requires libolm
go build github.com/k4lipso/gokill

# Create a config.json and run gokill
./gokill -c config.json
# Running gokill manually is annoying, it is acutally meant to run as systemd unit.
```

## Config Example

gokill is configured using a json file. it consists of a list of triggers, where each of the triggers as a list of 
actions that will be executed once triggered.

``` json
[ //list of triggers
    {
		"type": "UsbDisconnect", //triggers when the given device is disconnected
		"name": "First Trigger",
		"options": {
			"deviceId": "ata-Samsung_SSD_860_EVO_1TB_S4AALKWJDI102",
			"waitTillConnected": true //only trigger when usb drive was actually attached before
		},
        "actions": [ //list of actions that will be executed when triggered
            {
                "name": "unixCommand",
                "options": {
                    "command": "shutdown -h now"
                },
                "stage": 2 // defines the order in which actions are triggered.
            },
            {
		        "type": "SendTelegram",
		        "options": {
		        	"token": "3345823487:FFGdEFxc1pA18d02Akslw-lkwjdA92KAH2",
		        	"chatId": -832325872,
		        	"message": "attention, intruders got my device!",
		        	"testMessage": "this is just a test, no worries"
		        },
                "stage": 1 //this event is triggered first, then the shutdown
            }
        ]
    },
    {
		"type": "EthernetDisconnect",
		"name": "Second Trigger",
		"options": {
			"interfaceName": "eth0",
		},
        "actions": [
            {
                "name": "unixCommand",
                "options": {
                    "command": "env DISPLAY=:0 sudo su -c i3lock someUser" //example of locking someUser's screen as root
                }
            }
        ]
    }
]
```

## nix support
gokill enjoys full nix support. gokill exposes a nix flakes that outputs a gokill package, a nixosModule and more.
That means you can super easily incorporate gokill into your existing nixosConfigurations. 

### NixOS Module
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

### Build Documentation locally

``` bash
nix run github:k4lipso/gokill#docs
```

### Run integrations tests

``` bash
nix flake check github:k4lipso/gokill
```

## todos

- export snap
- export ppa

### actions
- [x] shutdown
- [ ] wipe ram
- [ ] ~~send mail~~
- send chat message
    - [x] telegram
    - [x] matrix
- [ ] delete data
- [ ] shred area
- [x] run command
- [ ] wordpress post
- [ ] ipfs command
- [buskill 'triggers'](https://github.com/BusKill/awesome-buskill-triggers)
    - [x] [lock-screen](https://github.com/BusKill/buskill-linux/tree/master/triggers)
    - [x] shutdown
    - [ ] luks header shredder
    - [ ] veracrypt self-destruct

### triggers
- [ ] no internet
- [x] [pull usb stick](https://github.com/deepakjois/gousbdrivedetector/blob/master/usbdrivedetector_linux.go)
- [x] ethernet unplugged
- receive specific chat message
    - [x] telegram
    - [ ] matrix
- [ ] power adapter disconnected
- [ ] unix command
- anyOf
    - trigger wrapper containing many triggers and fires as soon as one of them
      is triggered
- allOf
- [ ] ipfs trigger
