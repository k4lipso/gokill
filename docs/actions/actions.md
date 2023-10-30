# Available Actions:

# Print
Description: When triggered prints the configured message to stdout  
Values:
- **message**
	- Type: string
	- Descr: Message that should be printed
	- Default: ""

### Timeout
Description: When triggered waits given duration before continuing with next stage  
Values:
- **duration**
	- Type: int
	- Descr: duration in seconds
	- Default: 0

# Command
Description: When triggered executes given command  
Values:
- **command**
	- Type: string
	- Descr: command to execute
	- Default: 
- **args**
	- Type: string[]
	- Descr: args
	- Default: 

### Shutdown
Description: When triggered shuts down the machine  
Values:

