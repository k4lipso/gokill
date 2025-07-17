# gokill

gokill is a [software dead man's switch](https://en.wikipedia.org/wiki/Dead_man%27s_switch#Software). It belongs to the category of anti-forensic tools, providing a means to safeguard against potential repression. It is specifically crafted for worst-case scenarios, such as when intruders gain physical access to a device. In these intense situations, gokill can automatically perform tasks to enhance your security. Those could be:
- deleting data
- sending chat messages
- encrypting partitions 
- destroying encrypted partitions
- locking the screen 
- ect

#### documentation
A full list of Triggers and Actions with all their configuration options can be found here: https://k4lipso.github.io/gokill/

## usage
If you use NixOS gokill can easily be integrated into your system configuration - scroll down for more info on that.  

For all other linux distributions gokill currently needs to be built and setup manually.

``` bash
# Clone the gokill repository
git clone https://github.com/k4lipso/gokill
cd gokill

# Build gokill - requires libolm
go build cmd/gokill/gokill.go

# Create a config.json and run gokill
./gokill -c config.json
# Running gokill manually is annoying, it is acutally meant to run as systemd unit.
```

## Config Example

gokill is configured using a json file. it consists of a list of triggers, where each of the triggers as a list of 
actions that will be executed once triggered. The example configures gokill to send a message on Telegram and shutdown
the device as soon as a specific USB drive gets disconnected. In addition to that it locks the screen if an ethernet
cable is disconnected.

``` nix
[
    {
		"type": "UsbDisconnect",
		"name": "First Trigger",
		"options": {
			"deviceId": "ata-Samsung_SSD_860_EVO_1TB_S4AALKWJDI102",
			"waitTillConnected": true
		},
        "actions": [
            {
                "name": "unixCommand",
                "options": {
                    "command": "shutdown -h now"
                },
                "stage": 2
            },
            {
		        "type": "SendTelegram",
		        "options": {
		        	"token": "3345823487:FFGdEFxc1pA18d02Akslw-lkwjdA92KAH2",
		        	"chatId": -832325872,
		        	"message": "attention, intruders got my device!",
		        	"testMessage": "this is just a test, no worries"
		        },
                "stage": 1
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
                    "command": "env DISPLAY=:0 sudo su -c i3lock someUser"
                }
            }
        ]
    }
]
```

## nix support
gokill exposes a nix flakes that outputs a gokill package, a nixosModule and more.
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
