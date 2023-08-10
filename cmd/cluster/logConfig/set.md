==Long==
# Set

Set a managed clusters log configuration

A list of Connectors (can be found here)[https://community.sailpoint.com/t5/IdentityNow-Articles/Enabling-Connector-Logging-in-IdentityNow/ta-p/188107]
====

==Example==
```bash
sail cluster log set 2c91808580f6cc1a01811af8cf5f18cb -r TRACE -d 30 -c sailpoint.connector.ADLDAPConnector=TRACE 
```
====