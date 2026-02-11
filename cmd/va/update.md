==Long==
# Update

Perform update operations on a SailPoint VA. 

This command connects to the VA over SSH (port 22), runs an update check on the appliance, then reboots it. You must have network connectivity to the VA. It authenticates as the sailpoint user using the VA password.

====

==Example==
```bash
sail va update 10.10.10.25
sail va update 10.10.10.10 10.10.10.11 -p S@ilp0int -p S@ilp0int
```
====