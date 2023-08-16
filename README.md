
# Switch Library Manager Web
Easily manage your switch game backups

#### Features:
- Cross platform, works on Windows / Mac / Linux
- Web interface
- Scan your local switch backup library (NSP/NSZ/XCI)
- Read titleId/version by decrypting NSP/XCI/NSZ (requires prod.keys)
- If no prod.keys present, fallback to read titleId/version by parsing file name  (example: `Super Mario Odyssey [0100000000010000][v0].nsp`).
- Lists missing update files (for games and DLC)
- Lists missing DLCs
- ~~Automatically organize games per folder~~ Not yet done
- ~~Rename files based on metadata read from NSP~~ Not yet done
- ~~Delete old update files (in case you have multiple update files for the same game, only the latest will remain)~~ Not yet done
- ~~Delete empty folders~~ Not yet done
- Zero dependencies, all crypto operations implemented in Go. 

## Keys (optional)
Having a prod.keys file will allow you to ensure the files you have a correctly classified.
The app will look for the "prod.keys" file in the data folder.
You can also specify a custom location in the settings page.

Note: Only the header_key, and the key_area_key_application_XX keys are required.

## Naming template
The following template elements are supported:
- {TITLE_NAME} - game name
- {TITLE_ID} - title id
- {VERSION} - version id (only applicable to files)
- {VERSION_TXT} - version number (like 1.0.0) (only applicable to files)
- {REGION} - region
- {TYPE} - impacts DLCs/updates, will appear as ["UPD","DLC"]
- {DLC_NAME} - DLC name (only applicable to DLCs)

## Reporting issues
Please set debug mode to 'true', and attach the docker log to allow for quicker resolution.

## Usage

Please refer to the Docker documentation on how to run a Docker image.

##### Volumes inside the container
- `/usr/local/share/switch-library-manager-web`
- `/mnt/roms`
 
##### Example
```
$ docker run --rm \
	-v /home/johndoe/switch-library-manager-web:/usr/local/share/switch-library-manager-web:Z \
	-v /home/johndoe/Backups/Switch:/mnt/roms:Z \
	-p 3000:3000 \
	ghcr.io/dtrunk90/switch-library-manager-web
```

## Building
- Install and setup [Gulp](https://gulpjs.com)
- Install and setup [Go](https://go.dev)
- Clone the repo: `git clone https://github.com/dtrunk90/switch-library-manager-web.git`
- Execute `make`
- Binary will be available under build

#### Thanks
This program is based on [giwty's switch-library-manager](https://github.com/giwty/switch-library-manager)

This program relies on [blawar's titledb](https://github.com/blawar/titledb), to get the latest titles and versions.

