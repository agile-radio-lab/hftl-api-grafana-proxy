# REST API Grafana Proxy

This application serves as a converter from REST API response to a [Grafana simple json
datasource](https://github.com/grafana/simple-json-datasource) specific data format.

## Table data example

```
[{
    "type": "table",
    "columns": [
        {"text": "OWNER","type": "string"},
        {"text": "TABLE_NAME", "type": "string"}
    ],
    "rows": [
        ["LEAN", "AFS_DEVICE_ITEM"],
        ["LEAN", "AFS_MODULE"],
        ["LEAN", "AFS_NODE"],
        ["LEANHIS","AFS_DEVICE"],
        ["LEANHIS", "AFS_DEVICE_COMPANY"],
        ["LEANHIS", "AFS_DEVICE_DATA_ENTRY"]
    ]
}]
```
