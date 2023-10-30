# gokill
'gokill' is a daemon that completes some actions when a certain event occurs.
actions can vary from shuting down the machine to sending mails over erasing data.
actions can be triggert by certain conditions like specific outcomes of unix
comands or not having internet connection.

actions and triggers should be easy to extend and handled like plugins. they
also should be self documenting.
every action and trigger should be testable at anytime as a 'dry-run'.
actions can have a 'stage' defined. the lowest stage is started first,
and only when all actions on that stage are finished next stage is triggered

the killswitch will run as daemon. config should be read from
/etc/somename/config.json

many devices can be connected to each other over ipfs. that makes it possible
to send triggers to each other. for example device A can trigger an event on
device B. no matter where they are, no zentralized service necessary.

it should be evaluated if and how smartphones could be included to that.

## actions
- [x]shutdown
- [ ] wipe ram
- [ ]send mail
- [ ]delete data
- [ ]shred area
- [x]random command
- [ ]wordpress post
- [ ]ipfs command
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

## Config

##### Example
``` json
[ //list of triggers
    {
        "type": "command", //actual trigger
        "name": "custom name",
        "options": {
            "command": "true",
            "interval": "1m"
        },
        "actions": [
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
    }
]
```
