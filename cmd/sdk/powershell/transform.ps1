# Create transform
$JSON = @"
{
    "name": "New Transform",
    "type": "lookup",
    "attributes" : {
        "table" : {
            "USA": "Americas",
            "FRA": "EMEA",
            "AUS": "APAC",
            "default": "Unknown Region"
        }
    }
}
"@

$Transform = ConvertFrom-JsonToTransform -Json $JSON

 try {
    New-Transform -Transform $Transform
} catch {
    Write-Host ("Exception occurred when calling New-Transform: {0}" -f $_.ErrorDetails)
    Write-Host ("Response headers: {0}" -f $_.Exception.Response.Headers)
}