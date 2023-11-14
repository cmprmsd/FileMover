# File Mover
![Mover](mover.png)
This Go tool was written to move files/folders from A to B. 

The tool watches the source folders for new files/folders and moves them to their corresponding destinations. 

Folder pairs are specified in a `folder_pairs.conf` configuration file. 

To handle large files or folders with many files correctly the tool will wait for 10 seconds of idle time before moving any file.

## Usage
1. Create a `folder_pairs.conf` file with the desired folder pairs.
2. Set the DEBUG_LEVEL environment variable (optional).
3. Run the tool:

```sh
git clone github.com/cmprmsd/FileMover
go run FileMover.go
```


## Configuration File: folder_pairs.conf
To set up folder pairs for synchronization, create a configuration file named `folder_pairs.conf`. Each line in the file should specify a source and destination pair separated by a `colon` (`:`). Empty lines and lines starting with # or // are ignored as comments.

Example `folder_pairs.conf` content:

```bash
# Name 1 (optional comments with # or //)
/source1:/destination1
# Name 2
/source2:/destination2
```

## Example Use Case

I use this this tool on a central [Syncthing](https://syncthing.net/) instance to send files between my devices even if any is offline.

You'll have one library per device with folders for each additional device you want to manage. Each managed device will also have a separate library for receiving data.

E.g.
```bash
# Tablet
/SourceTablet/ToPhone:/ReceivePhone/
/SourceTablet/ToPC:/ReceivePC/

# PC
/SourcePC/ToPhone:/ReceivePhone/
/SourcePC/ToTablet:/ReceiveTablet/

# Phone
/SourcePhone/ToPC:/ReceivePC/
/SourcePhone/ToTablet:/ReceiveTablet/
```


## Debug Levels

The tool supports three debug levels:

1. **Debug Level 0:** No logs are displayed (default).
2. **Debug Level 1:** Minimal logs, including errors and important events.
3. **Debug Level 2:** Detailed logs, including all events.

You can set the desired debug level by defining the `DEBUG_LEVEL` environment variable. For example:

```sh
export DEBUG_LEVEL=2
# or 
DEBUG_LEVEL=2 go run FileMover.go
```


The tool will start watching the specified source directories and synchronize them with their corresponding destinations when new files/folders are added.


## Dependencies
This tool uses the `github.com/rjeczalik/notify` package for file and directory change notifications.

## License
This project is licensed under the GPL - see the LICENSE file for details.



