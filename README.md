# htop-go

simple cpu and memory viewer for Linux

## Dependencies

- Go 1.20+
- Linux (/proc)

## Install
```shell
go build -o htop-go
```
## Usage

`./htop-go`

## Example 
```shell
CPU: Intel(R) Xeon(R) CPU E5-2690 v4 @ 2.60GHz
Memory: 1.30G/32.0G
Uptime: 4h29m57s
CPU0 0.0%
CPU1 0.0%
CPU2 0.0%
CPU3 0.0%
CPU4 2.0%
CPU5 0.0%
CPU6 2.0%
CPU7 0.0%
CPU8 1.0%
CPU9 0.0%
CPU10 1.0%
CPU11 0.0%
CPU12 0.0%
CPU13 1.0%
CPU14 2.0%
CPU15 0.0%
CPU16 0.0%
CPU17 0.0%
CPU18 0.0%
CPU19 1.0%
CPU20 0.0%
CPU21 0.0%
CPU22 0.0%
CPU23 5.0%
CPU24 0.0%
CPU25 0.0%
CPU26 1.0%
CPU27 0.0%
PID    NAME    CPU%    MEM%
1    (systemd)    0.0%
107    (systemd-udevd)    0.0%
11823    (polkitd)    0.0%
123    (systemd-resolve)    0.0%
1269    (cloudcode_cli)    0.0%
128    (systemd-timesyn)    0.0%
1292    (node)    0.0%
187    (cron)    0.0%
188    (dbus-daemon)    0.0%
2    (init-systemd(Ub)    0.0%
200    (systemd-logind)    0.0%
204    (wsl-pro-service)    0.0%
211    (vsftpd)    0.0%
223    (rsyslogd)    0.0%
228    (agetty)    0.0%
231    (containerd)    0.0%
234    (agetty)    0.0%
240    (unattended-upgr)    0.0%
313    (dockerd)    0.0%
60    (systemd-journal)    0.0%
633    (login)    0.0%
652    (SessionLeader)    0.0%
653    (Relay(654))    0.0%
654    (sh)    0.0%
655    (sh)    0.0%
660    (sh)    0.0%
664    (node)    0.0%
696    (systemd)    0.0%
697    ((sd-pam))    0.0%
706    (bash)    0.0%
748    (SessionLeader)    0.0%
749    (Relay(753))    0.0%
753    (node)    0.0%
760    (SessionLeader)    0.0%
76012    (bash)    0.0%
76063    (SessionLeader)    0.0%
76064    (Relay(76071))    0.0%
76071    (bash)    0.0%
761    (Relay(762))    0.0%
762    (node)    0.0%
768    (node)    0.0%
79705    (go)    0.0%
79866    (htop-go)    1.0%
799    (node)    0.0%
8    (init)    0.0%
811    (node)    9.0%
882    (bash)    0.0%
913    (gopls)    0.0%
925    (gopls)    0.0%
```