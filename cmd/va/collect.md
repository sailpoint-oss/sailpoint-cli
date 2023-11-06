==Long==
# Collect

Collect files from a remote virtual appliance.

Files are collected over SFTP. Passwords are provided via the --password (-p) flag or they will be prompted for at runtime. Server addresses can be DNS names or IP addresses, and they're provided as arguments, separated by spaces.

Log Files:
```bash
/home/sailpoint/log/ccg.log  
/home/sailpoint/log/charon.log   
```

Config Files:
```bash
/home/sailpoint/proxy.yaml  
/etc/systemd/network/static.network  
/etc/resolv.conf  
```
====

==Example==
```bash
sail va collect 10.10.10.25 10.10.10.26 -p S@ilp0int -p S@ilp0int
sail va collect 10.10.10.25 --config
sail va collect 10.10.10.26 --log
sail va collect 10.10.10.25 --log --output log_files
sail va collect 10.10.10.25 --output all_files
```
====