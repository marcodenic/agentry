# Plugin Registry

This directory hosts a minimal registry used by `agentry plugin fetch`.
Each entry in `index.json` lists the download `url` of a plugin archive and
its SHA256 checksum.

To add a new plugin, append an object to the `plugins` array:

```json
{
  "name": "myplugin",
  "url": "https://example.com/myplugin.zip",
  "sha256": "<sha256>"
}
```

Please keep the list sorted alphabetically and open a pull request.

