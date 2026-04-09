# COMP429-Project

## Building
Download golang from here.
https://go.dev/dl/

And a simple go run . from this directory will work.

## Contributions

### Demian Garcia
= Added peer/connection tracking for both incoming and outgoing TCP connections
- Implemented handling for connect, list, send, terminate, and exit
- Added error handling for invalid input, dupe connections, self-connections, and missing connection IDs
- Added disconnect cleanup so terminated peers are removed from the active connection list on both sides

### Bella Felipe
- Implemented command-line interface (input parsing, command handling)
- Implemented `help`, `myip`, `myport`, and `exit`
- Implemented `connect` command and outgoing connections
- Designed and implemented connection storage system
- Implemented `list` command for active connections
- Implemented `terminate` command to close connections
- Implemented `send` command for peer-to-peer messaging
- Added exit cleanup to close all active connections
- Built TCP server listener for incoming connections
