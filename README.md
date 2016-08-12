# M3 Install

The installer utility for M3. This application allows users to install
modpacks defined according to the M3 spec. Seamless updates are also
supported, as `m3-install` will properly clean up any obsolete mods or
Forge files it encounters during the installation

In effect, running `m3-install` from your minecraft directory will set
the environment to match the specification found in `modpack.json`.

## Command options

* `-f {file}` - Specification to import (default: "modpack.json")
* `-dir {dir}` - Set the working directory (defualt: ".")
* `-moddir {dir}` - Set the mod directory (default: "mods")
* `-n {N}` - Max concurrent downloads (default: 3)
* `-server` - Run installer in server mode (default: false)
* `-client` - Run installer in client mode (default: false)
* `-v` - Use verbose output (default: false)
* `-vv` - Use very verbose output (default: false)

## Specification format
```json
{
    "Forge": {
        "Version": "<required_forge_server_version>"
    },
    "Ignore": [
        "<filename_to_ignore.jar>",
        ...
    ],
    "Mods": [
        {
            "Name": "<required_mod_name>",
            "Version": "<optional_mod_version>",
            "Checksum": "<optional_file_sha356_checksum}",
            "Url": "<required_file_download_url>"
        },
        ... 
        {
            "Name": "<required_mod_name>",
            "Version": "<optional_mod_version>",
            "Checksum": "<optional_file_sha356_checksum>",
            "Curse": "<required_curseforge_file_id>"
        },
        ...
    ]
}
```
