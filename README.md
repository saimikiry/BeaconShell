# Building and usage
---
To build you will need:
- go (tested on 1.22.2)

Use the following command.
```
go build -o BeaconShell.exe .\Source\main.go .\Source\utils.go .\Source\inject.go
```

To start, use the following command.
```
./BeaconShell.exe
```

# BeaconShell toolkit
---
Use "/BS help" to 
```
Targets status:
        - /BS targets                   Show current targets list;
        - /BS add <ip:port>             Add target <ip:port> to set;
        - /BS add list <file>           Add targets from <file> to set;
        - /BS group <group> <host_id>   Add target <host_id> to <group>;
        - /BS remove <host_id>          Remove target from set;
        - /BS off                       Stop sending commands to the all hosts;
        - /BS off <host_id>             Stop sending commands to the host;
        - /BS off group <group>         Stop sending commands to the <group>;
        - /BS on                        Resume sending commands to the all hosts;
        - /BS on <host_id>              Resume sending commands to the host;
        - /BS on group <group>          Resume sending commands to the <group>.

Shell injecting:
        - /BS inject <file_path> <shell_type> <OS> <Arch> <ip> <port>           Inject bind shell to code and compile it. (BETA)

Configuration:
        - /BS config                    Get current configuration;
        - /BS timeout <time>            Set targets response timeout to <time> millisecond(s);
        - /BS buffer <size>             Set buffer size to <size> byte(s).

Other:
        - /BS help                      Show this list;
        - /BS scenario <scenario>       Start scenario from file <scenario>;
        - /BS stop                      Finish session.
```
