
# OttoMap Command Line Interface

OttoMap is a tool that translates TribeNet turn report files into JSON data and generates maps. This document provides instructions for running OttoMap from the command line.

## Important Assumptions

These instructions assume that you're running `ottomap` (on macOS and Linux) or `ottomap.exe` (on Windows) from the root directory of the project. We'll use the default directories (`data/input` and `data/output`) to keep the examples short.

> **NOTE**:
> - On Windows, use `ottomap.exe` to run commands.
> - On macOS and Linux, use `./ottomap` to run commands.

## Storing Turn Report Files

OttoMap expects turn report files to be stored in the `data/input` directory within the project. This directory includes `.gitignore` files to prevent accidentally uploading sensitive data to version control systems.

## Available Commands

OttoMap provides the following commands:

### `version`

The `version` command displays the current version of OttoMap.

```bash
$ ottomap version
```

Output example:
```
0.2.0
```

### `index reports`

The `index reports` command creates an index for the turn report files to process. The index is saved in the file `data/input/config.json` and contains all the metadata about the reports.

```bash
$ ottomap index reports
```

Output example:
```
config: 0899-12.0991: added    input/899-12.0991.report.txt
config: 0900-01.0991: added    input/900-01.0991.report.txt
config: saved   data/config.json
index:  created data/config.json
```

### `map`

The `map` command generates a map based on the indexed turn report files.

```bash
$ ottomap map --turn 901-12 --clan 0991 --show-grid-coords
```

Output example:
```
map: config: file data/config.json
config: loaded data/config.json
map: config: path   data/config.json
map: config: output data/output
map: config: clan   "0991"
map: config: turn year   901
map: config: turn month   12
map: created  data/output/0900-01.0991.wxx
map: created  data/output/0991.wxx
```

You can specify additional options for the `map` command:

- `--output` or `-o`: Specify the output directory for the generated map files.
- `--config` or `-c`: Specify the path to a configuration file for customizing the map generation.

## Running OttoMap

To run OttoMap, follow these steps:

1. **Open a terminal or command prompt**:
    - On Windows, press `Win + R`, type `cmd`, and press `Enter`.
    - On macOS, press `Cmd + Space`, type `Terminal`, and press `Enter`.
    - On Linux, press `Ctrl + Alt + T`.

2. **Navigate to the OttoMap project directory**:
    - Use the `cd` command followed by the path to your project directory. For example:
      ```bash
      cd path/to/ottomap
      ```

3. **Place your turn report files in the `data/input` directory**.

4. **Run the desired command with any necessary options**:
    - To index the report files, run:
      ```bash
      $ ottomap index reports
      ```
    - To generate a map using the default settings, run:
      ```bash
      $ ottomap map --turn 901-12 --clan 0991
      ```

This will index the turn report files and generate a map in the default output directory.

Note: Detailed information about the configuration options and map generation settings can be found in the project's documentation.