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

Client guidance:

- new MCP clients should call the stable names listed above
- prefer documenting and testing against stable names only

## Tool Summary

### `search_videos`

Purpose:

- Search videos across configured sites using a keyword.

Required input:

- `keyword` (`string`): one or more words to search for

Optional input:

- none

Behavior:

- searches the configured sites using the provided keyword
- returns text output in the current implementation

Input:

```json
{
  "keyword": "keyword"
}
```

Output:

- text table of search results

Example output:

```text
[1.1] Movie Title
  vod_id: 123
```

Example call:

```json
{
  "keyword": "movie"
}
```

### `get_video_detail`

Purpose:

- Fetch full video metadata for one configured site entry.
- Optionally resolve a single episode URL instead of returning the full detail body.

Required input:

- `site_id` (`string`): configured site ID, such as `1.1`
- `vod_id` (`string`): video ID on that site

Optional input:

- `episode` (`string`): when provided, return the matching episode URL if found

Behavior:

- when `episode` is omitted, returns the full detail text
- when `episode` is present, returns only the matching episode URL text

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

Example output without `episode`:

```text
Title: Movie Title
Type: Movie
Episodes:
  1 => https://...
```

Example output with `episode`:

```text
https://example.com/play/episode-1
```

Example call:

```json
{
  "site_id": "1.1",
  "vod_id": "123",
  "episode": "1"
}
```

### `list_sites`

Purpose:

- Return all configured series and domains currently stored in the local config.

Required input:

- none

Optional input:

- none

Behavior:

- lists all series IDs, domain IDs, URLs, and failure counts
- useful before calling `get_video_detail`, `add_site`, or `remove_site`

Input:

```json
{}
```

Output:

- text list of configured series and domains

Example output:

```text
Series 1:
  [1.1] URL: example.com [Failures: 0]
  [1.2] URL: backup.example.com [Failures: 1]
```

Example call:

```json
{}
```

### `add_site`

Purpose:

- Add a new site domain to the local config.
- Can either create a new series or append a domain to an existing series.

Required input:

- `domain` (`string`): bare host or full endpoint URL

Optional input:

- `series_id` (`string`): target series ID; if omitted, a new series is created

Behavior:

- when `series_id` is omitted, creates a new series with this domain
- when `series_id` is provided, adds the domain to the existing series

Input:

```json
{
  "domain": "example.com",
  "series_id": "1"
}
```

Output:

- success or error text

Example output:

```text
Added domain backup.example.com to series 1.
```

Example calls:

Create a new series:

```json
{
  "domain": "example.com"
}
```

Add to an existing series:

```json
{
  "domain": "backup.example.com",
  "series_id": "1"
}
```

### `remove_site`

Purpose:

- Remove either an entire series or one domain inside a series.

Required input:

- `id` (`string`): either a series ID like `1` or a domain ID like `1.1`

Optional input:

- none

Behavior:

- `1` removes the entire series
- `1.1` removes only one domain from a series
- if the removed domain is the last domain in its series, the empty series is removed

Input:

```json
{
  "id": "1.1"
}
```

Output:

- success or error text

Example output:

```text
Removed domain 1.1.
```

Example calls:

Remove a whole series:

```json
{
  "id": "1"
}
```

Remove a single domain:

```json
{
  "id": "1.1"
}
```

### `validate_sites`

Purpose:

- Validate the current stored configuration and report structural or data issues.

Required input:

- none

Optional input:

- none

Behavior:

- returns a validation summary in text form
- useful before search or detail operations if config quality is uncertain

Input:

```json
{}
```

Output:

- validation summary text

Example output:

```text
Configuration is valid.
```

Possible issue output:

```text
Found 1 issue(s):
  [warning] site 1.2: missing URL
```

Example call:

```json
{}
```

### `health_check`

Purpose:

- Return an operational summary of the current config.

Required input:

- none

Optional input:

- none

Behavior:

- reports config path, version, series count, domain count, warnings, and errors

Input:

```json
{}
```

Output:

- health summary text

Example output:

```text
Config:   /Users/example/Library/Application Support/zpcli/sites.json
Version:  1
Series:   2
Domains:  3
Errors:   0
Warnings: 1
```

Example call:

```json
{}
```

## Compatibility Notes

- stable names should be preferred by new clients
- text output remains the current transport format
- future work should add structured result payloads without silently breaking existing text clients
- output examples in this document are illustrative and may vary slightly with data shape and rendering rules
