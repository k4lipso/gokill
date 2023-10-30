# Available Triggers:


### UsbDisconnect
Description: Triggers when given usb drive is disconnected  
Values:
- **waitTillConnected**
	- Type: bool
	- Descr: Only trigger when device was connected before
	- Default: true
- **deviceId**
	- Type: string
	- Descr: Name of device under /dev/disk/by-id/
	- Default: ""


