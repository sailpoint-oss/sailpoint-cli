==Long==
# Troubleshoot

Perform troubleshooting operations against a VA. 

This command connects to the VA over SSH (port 22) to run the SailPoint STUNT troubleshooting script, then collects the resulting log archive via SFTP. The script is bundled with the CLI and verified against a pinned checksum before it runs; it is uploaded to `/tmp` on the VA, executed, and removed. The VA does not need internet access to fetch the script. You must have network connectivity to the VA. It authenticates as the sailpoint user using the VA password.

====

==Example==
```bash
sail va troubleshoot 10.10.10.10
```
====