# OttoMap

OttoMap is a tool that translates TribeNet turn report files into JSON data.

A future version of the tool will convert the JSON data into a map.

> WARNING OttoMap is in early development. 
> I am breaking things and changing types almost daily. 
 
## Overview
I'm planning on translating a small subset of the turn report.
See the files in the `domain` directory to get an idea of what we're looking at.

I think that will be enough data to feed the map generator.
Let me know if you think that there's something missing.

## Input Data
OttoMap expects all turn reports to be in text files in a single directory.

OttoMap loads all files that match the pattern "YEAR-MONTH.CLAN_ID.input.txt."
YEAR and MONTH are the three-digit year and two-digit month from the "Current Turn" line of the report.
CLAN_ID is the four-digit identifier for the clan (it must include the leading zero).

```bash
$ ls -1 input/*.txt

input/899-12.0138.input.txt
input/900-01.0138.input.txt
input/900-02.0138.input.txt
input/900-03.0138.input.txt
input/900-04.0138.input.txt
input/900-05.0138.input.txt
```

The files are created by opening the turn report (the `.DOCX` file),
selecting all the text, and pasting it into a plain text file.

```bash
$ ls -1 input/*.docx
input/899-12.0138.Turn-Report.docx
input/900-01.0138.Turn-Report.docx
input/900-02.0138.Turn-Report.docx
input/900-03.0138.Turn-Report.docx
input/900-04.0138.Turn-Report.docx
input/900-05.0138.Turn-Report.docx

$ file input/*

input/899-12.0138.Turn-Report.docx: Microsoft Word 2007+
input/899-12.0138.input.txt:        Unicode text, UTF-8 text
input/900-01.0138.Turn-Report.docx: Microsoft Word 2007+
input/900-01.0138.input.txt:        Unicode text, UTF-8 text
input/900-02.0138.Turn-Report.docx: Microsoft Word 2007+
input/900-02.0138.input.txt:        Unicode text, UTF-8 text
input/900-03.0138.Turn-Report.docx: Microsoft Word 2007+
input/900-03.0138.input.txt:        Unicode text, UTF-8 text
input/900-04.0138.Turn-Report.docx: Microsoft Word 2007+
input/900-04.0138.input.txt:        Unicode text, UTF-8 text
input/900-05.0138.Turn-Report.docx: Microsoft Word 2007+
input/900-05.0138.input.txt:        Unicode text, UTF-8 text
```

Spaces, line breaks, page breaks, and section breaks are important to the parser.
Please try to avoid altering them.


## Grids and the Big Map
The big map is divided into 676 grids arranged in 26 columns and 26 rows.
The grids use letters, not digits, for their coordinates on the big map.
The grid at the top left is (A, A) and the grid at the bottom right is (Z, Z).

Each grid has 30 columns and 21 rows.
The hex at the top left is 1, 1 and the hex at the bottom right is 30, 21.
Hexes are "flat" on the top and even rows are shifted down.
For example, hex (13, 10) has

1. (13, 9) to the north
2. (14, 9) to the north-east
3. (14, 10) to the south-east
4. (13, 11) to the south
5. (12, 10) to the south-west
6. (12, 9) to the north-west

In turn reports, a hex in the grid is usually displayed as "AA 1310."

You can convert grid coordinates to big map coordinates.
A coordinate like "VN 0810" is "(V8, N10)" on the big map.

You can convert also convert grid coordinates to absolute coordinates by scaling the column and row values.
For our coordinate of "VN 0810," "V" is the 22nd grid from the left and "N" is the 14th grid from the top.
This gives us a column of (22 - 1) * 30 + (8 - 1) = 637 and row of (14 - 1) * 21 + (10 - 1) = 282, or (637, 282)
(We subtract one before multiplying because absolute coordinates start at zero, not one.)
