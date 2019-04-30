# Telnet-Chat

Telnet-Chat server is a simple chat server, accessible with a telnet client

# Building & running

To build, simply run the provided script: `./build.sh`

To run, edit the config values in `etc/config.json` including

`bind_addr` String. Address to bind the TCP listener to
`bind_port` String. Port to bind the TCP listener to
`log_file` String. Path to a file to log messages
`rooms` String array. List of room names to create upon server initialization


and run the server by executing `./telnet-chat`

# Usage

Connect to the server by using a telnet client

```bash
 telnet 0.0.0.0 11000
Trying 0.0.0.0...
Connected to 0.0.0.0.
Escape character is '^]'.
Desired username: gopher
Select a room to join by its number
Available rooms:
	0: General chat
	1: Announcements
	2: Golang
0
Welcome to "General chat"!
Type /help to display the help message
>> hello
```

# Help

At any time you may get help by typing `/help` in a room

Available commands are:

```
/help: print this help message
/name: change your username
/leave: leave the current room and choose another
/quit: disconnect from the server
```
