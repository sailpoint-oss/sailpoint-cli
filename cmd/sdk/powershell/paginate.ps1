
$JSON = @"
{
	"indices": [
		"identities"
	],
	"query": {
		"query": "*",
		"fields": [
		"name"
		]
	},
	"sort": [
		"-displayName"
	]
	}
"@

$Search = ConvertFrom-JsonToSearch -Json $JSON

try {

    Invoke-PaginateSearch -Increment 50 -Limit 10000 -Search $Search

} catch {
    Write-Host ("Exception occurred when calling Invoke-PaginateSearch: {0}" -f $_.ErrorDetails)
    Write-Host ("Response headers: {0}" -f $_.Exception.Response.Headers)
}