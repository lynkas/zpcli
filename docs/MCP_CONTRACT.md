# MCP Contract

## Server

- name: `zpcli`
- protocol version: `2024-11-05`
- transports:
  - stdio
  - SSE over HTTP

## Tool Naming

Preferred stable tool names:

- `search_videos`
- `get_video_detail`
- `list_sites`
- `add_site`
- `remove_site`
- `validate_sites`
- `health_check`

Backward-compatible aliases currently exposed:

- `search`
- `get_detail`

## Tool Summary

### `search_videos`

Input:

```json
{
  "keyword": "keyword"
}
```

Output:

- text table of search results

### `get_video_detail`

Input:

```json
{
  "site_id": "1.1",
  "vod_id": "123",
  "episode": "1"
}
```

Output:

- full detail text when `episode` is omitted
- direct episode URL text when `episode` is provided

### `list_sites`

Input:

```json
{}
```

Output:

- text list of configured series and domains

### `add_site`

Input:

```json
{
  "domain": "example.com",
  "series_id": "1"
}
```

Output:

- success or error text

### `remove_site`

Input:

```json
{
  "id": "1.1"
}
```

Output:

- success or error text

### `validate_sites`

Input:

```json
{}
```

Output:

- validation summary text

### `health_check`

Input:

```json
{}
```

Output:

- health summary text

## Compatibility Notes

- stable names should be preferred by new clients
- text output remains the current transport format
- future work should add structured result payloads without silently breaking existing text clients
