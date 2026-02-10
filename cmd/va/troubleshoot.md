==Long==
# Troubleshoot

Perform troubleshooting operations against a VA. 

This command connects to the VA over SSH (port 22) to run a troubleshooting script, then collects the resulting stuntlog file via SFTP. You must have network connectivity to the VA. It authenticates as the sailpoint user using the VA password.

====

==Example==
```bash
sail va troubleshoot 10.10.10.10
```
====