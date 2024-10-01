$Limit = 250 # Int32 | Max number of results to return. See [V3 API Standard Collection Parameters](https://developer.sailpoint.com/docs/api/standard-collection-parameters) for more information. (optional) (default to 250)
$Offset = 0 # Int32 | Offset into the full result set. Usually specified with *limit* to paginate through the results. See [V3 API Standard Collection Parameters](https://developer.sailpoint.com/docs/api/standard-collection-parameters) for more information. (optional) (default to 0)
$Count = $true # Boolean | If *true* it will populate the *X-Total-Count* response header with the number of results that would be returned if *limit* and *offset* were ignored.  Since requesting a total count can have a performance impact, it is recommended not to send **count=true** if that value will not be used.  See [V3 API Standard Collection Parameters](https://developer.sailpoint.com/docs/api/standard-collection-parameters) for more information. (optional) (default to $false)
$Filters = 'sourceId eq "f4e73766efdf4dc6acdeed179606d694"' # String | Filter results using the standard syntax described in [V3 API Standard Collection Parameters](https://developer.sailpoint.com/docs/api/standard-collection-parameters#filtering-results)  Filtering is supported for the following fields and operators:  **id**: *eq, in*  **identityId**: *eq*  **name**: *eq, in*  **nativeIdentity**: *eq, in*  **sourceId**: *eq, in*  **uncorrelated**: *eq* (optional)

# Accounts List
try {
    
    Get-Accounts -Limit $Limit -Offset $Offset -Count $Count -Filters $Filters

} catch {
    Write-Host ("Exception occurred when calling Get-Accounts: {0}" -f $_.ErrorDetails)
    Write-Host ("Response headers: {0}" -f $_.Exception.Response.Headers)
}