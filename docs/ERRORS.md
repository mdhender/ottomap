# Parsing Errors

You may have to update the text file copies of your report files.

If you don't understand what needs to be fixed, please ask for help on the TribeNet Discord's `#mapping` channel.

## General Notes

When the parser encounters a line that it doesn't recognize, it will print the report id, the input, and then an error message.

```text
parse: report 0900-01.0138: unit 0138e1: parsing error
parse: input: "0138e1 Status: PRAIRIE,,River S 0138e1"
parse: error: status:1:24 (23): no match found, expected: [ \t] or [0-9]
```

The report id should help you locate the file that needs to be fixed.
(Please update the `.txt` copy of the file; the original `.docx` is not used by this application.)

If the unit id is available, it will also be displayed to help you find the section of the report that needs to be fixed.

The line shows the input from that report file.

The error message shows the section being parsed, the line number, the column number, and the parser's best guess at
what the problem is.

Note that the line number is always 1 because of the way the application looks at the input.

The column number shows you where the error happened.
(It's usually pretty close, anyway.)
Use that to help figure out what to fix.

After you've made your update (again, please don't update your original `.docx` report file),
just restart the application.

> NOTE:
I'm trying to get all the error messages to be consistent.
If you notice one that's wonky, please report it.

## No movement results found
If you run `ottomap map` and it ends with a line like `map: no movement results found`,
the likely cause is a copy+paste error with the report file.

Check that the first line of the report file starts with `Tribe 0nnn` where `0nnn` is your clan number.

If it does, it might be that your text editor is saving
[BOM](https://en.wikipedia.org/wiki/Byte_order_mark)
bytes to the file.
Please try running with the `--skip-bom` flag.

If that doesn't work, please report the error on the `#mapping-tools` channel of the Discord server.

## Backslashes
The report uses backslashes ("\") as movement step separators.
When we report an error, you'll see two backslashes.
That's because backslashes are special to the `printf` statement, so it doubles them on output.

If you see:

```text
"Scout 1:Scout SE-RH, River SE S SW\\NE-PR, River S\\ not enough M.P's to move to SE into PRAIRIE, nothing of interest found"
```

The line in the report is actually:

```text
Scout 1:Scout SE-RH, River SE S SW\NE-PR, River S\ not enough M.P's to move to SE into PRAIRIE, nothing of interest found
```

## Scout lines

Sometimes a backslash should actually be a comma.
If you have an error like this:

```text
parse: section: scout 1: "Scout 1:Scout SE-RH, River SE S SW\\NE-PR, River S\\SE-PR, River SE S SW\\ 1540\\NE-PR, River S\\ not enough M.P's to move to SE into PRAIRIE, nothing of interest found"
parse: report 0900-01.0138: unit 0138e1: parsing error
parse: input: "1540"
parse: error: scout:1:1 (0): no match found, expected: "Can't Move on", "N", "NE", "NW", "No Ford on River to", "Not enough M.P's to move to", "S", "SE" or "SW"
```

The fix is to replace the backslash with a comma:

```text
Scout 1:Scout SE-RH, River SE S SW\NE-PR, River S\SE-PR, River SE S SW, 1540\NE-PR, River S\ not enough M.P's to move to SE into PRAIRIE, nothing of interest found
```

Sometimes there are extra characters in the input.
This is due to the GMs making a typo when updating your turn report.
They do a lot of work to make it presentable and sometimes make an honest mistake.

```text
parse: section: scout 1: "Scout 1:Scout SW-PR,  \\SW-PR,  \\-,  1138e1\\SW-CH,  \\SW-PR,  O SW,  NW\\ Not enough M.P's to move to S into PRAIRIE,  Nothing of interest found"
parse: report 0900-03.0138: unit 2138e1: parsing error
parse: input: "-,  1138e1"
parse: error: scout:1:1 (0): no match found, expected: "N", "NE", "NW", "S", "SE", "SW", [Cc] or [Nn]
```

The fix is to remove those characters:

```text
Scout 1:Scout SW-PR,  \SW-PR,    1138e1\SW-CH,  \SW-PR,  O SW,  NW\ Not enough M.P's to move to S into PRAIRIE,  Nothing of interest found
```

> You may want to confer with the GM to find out what the line should actually have been.

You may see a line start with `Scout ,` instead of just `Scout `:

```text
parse: section: scout "Scout 1:Scout , Can't Move on Ocean to N of HEX,  Patrolled and found 2138e1"
parse: section: scout ", Can't Move on Ocean to N of HEX,  Patrolled and found 2138e1"
parse: section: scout 1: "Scout 1:Scout , Can't Move on Ocean to N of HEX,  Patrolled and found 2138e1"
parse: report 0900-04.0138: unit 1138e1: parsing error
parse: input: ", Can't Move on Ocean to N of HEX,  Patrolled and found 2138e1"
parse: error: scout:1:1 (0): no match found, expected: "N", "NE", "NW", "S", "SE", "SW", [Cc] or [Nn]
```

In that case, just remove the comma:

```text
Scout 1:Scout Can't Move on Ocean to N of HEX,  Patrolled and found 2138e1
```

## Status lines

Sometimes there are extra commas in the status line.
If you have an error like this:

```text
parse: report 0900-01.0138: unit 0138e1: parsing error
parse: input: "0138e1 Status: PRAIRIE,,River S 0138e1"
parse: error: status:1:24 (23): no match found, expected: [ \t] or [0-9]
```

Please remove the extra comma:

```text
0138e1 Status: PRAIRIE,River S 0138e1
```

Sometimes there is a missing comma that should follow River, Ford, or Pass directions.

```text
parse: report 0900-01.0138: unit 0138e1: parsing error
parse: input: "0138e1 Status: PRAIRIE,,River S 0138e1"
parse: error: status:1:33 (32): no match found, expected: ",", "N", "NE", "NW", "S", "SE", "SW", [ \t] or EOF
```

Please insert the comma after the list of directions:

```text
0138e1 Status: PRAIRIE,River S, 0138e1
```
