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
