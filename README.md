# OttoMap

OttoMap is a tool that translates TribeNet turn report files into JSON data.

A future version of the tool will convert the JSON data into a map.

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
