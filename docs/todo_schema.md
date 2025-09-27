# TODO Storage Schema

The TODO tooling (builtins and TUI board) now share `internal/todo.Service` for reading and writing records. Items live under a namespaced memstore key derived from the workspace path (`todo:project:<sha1-prefix>`). Each namespace stores one `meta:schema_version` entry so future migrations can detect and upgrade data in-place.

## Current Schema (`v1`)

- **Items** are stored under keys `item:<id>` with the JSON payload matching `todo.Item`.
- **IDs** are nanosecond timestamps generated at creation time.
- **Tags** are trimmed, de-duplicated case-insensitively, and sorted to keep list filtering stable.
- **Status / Priority** values are stored verbatim, allowing the CLI and TUI to evolve independently while remaining compatible.

When the service initializes, it will create the schema version entry if it does not exist and will error if an unexpected version is encountered. Future migrations can bump `todo.SchemaVersion` and provide upgrade tooling without touching individual consumers.
