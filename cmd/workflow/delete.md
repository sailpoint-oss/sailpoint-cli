==Long==
# Delete
Delete a workflow in IdentityNow. You can delete multiple workflows at once, and you can delete a set of workflows specified in a file. 

## API References:
 - https://developer.sailpoint.com/idn/api/beta/delete-workflow
====

==Example==

## Arguments:
```bash
sail workflow delete id1
sail workflow delete id1 id2 ...
sail workflow del $(cat list_of_workflowIDs.txt) 
```
====