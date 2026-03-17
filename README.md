# atlas.notes

![banner-image](banner-image.png)

`atlas.notes` is a beautiful, terminal-based markdown note manager for the Atlas Suite. It allows you to quickly create, edit, read, and delete notes from your terminal with high-fidelity rendering.

## Features
- **Markdown Support:** All notes are stored as `.md` files.
- **Beautiful Rendering:** Uses `glamour` for high-fidelity markdown rendering in the terminal.
- **CRUD Operations:** Create, List, Read, Update, and Delete notes.
- **Safe Deletion:** Confirmation required for deleting notes.
- **System Integration:** Opens your preferred `$EDITOR`.

## Installation
Requires [gobake](https://github.com/fezcode/gobake).

```bash
cd atlas.notes
gobake build
./build/atlas.notes-windows-amd64.exe -v
```

## Usage
```bash
atlas.notes ls             # List all notes
atlas.notes new my-note    # Create a new note
atlas.notes read my-note   # Read and render a note
atlas.notes edit my-note   # Edit a note
atlas.notes rm my-note     # Delete a note
```

## Storage
Notes are stored in `~/.atlas/atlas.notes.data/`.

## License
MIT
