==Long==
# Set

Set a VA cluster's log configuration. You can update a cluster's root logging level, the duration of its logging, and the connector logging class.

This example command sets the "TRACE" root logging level, a duration of 30 minutes, and a connector logging class of "sailpoint.connector.ADLDAPConnector=TRACE". 

Refer to your respective [connector guide](https://documentation.sailpoint.com/connectors/identitynow/landingpages/help/landingpages/identitynow_connectivity_landing.html) to see which connector logging classes are available. 
====

==Example==
```bash
sail cluster log set 2c91808580f6cc1a01811af8cf5f18cb -r TRACE -d 30 -c sailpoint.connector.ADLDAPConnector=TRACE 
```
====