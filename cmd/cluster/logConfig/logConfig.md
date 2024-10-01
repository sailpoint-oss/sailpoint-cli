==Long==

# Log Config

Get or set a VA cluster's log configuration.

## API Reference:
 - https://developer.sailpoint.com/docs/api/beta/managed-clusters
====

==Example==
```bash
sail cluster log get 2c91808580f6cc1a01811af8cf5f18cb
sail cluster log set 2c91808580f6cc1a01811af8cf5f18cb -r TRACE -d 30 -c sailpoint.connector.ADLDAPConnector=TRACE
```
====