# GoAnalysis

GoAnalysis does simple analysis of your source code.

For example:

```
$ goanalysis std

functions with N return arguments:
-----------------------------
returns | number of functions
      0 | 8693
      1 | 7743
      2 | 1975
      3 | 205
      4 | 25
      5 | 4
      6 | 1
      7 | 1
-----------------------------

total number of functions: 18647
total number of functions with multiple return parameters: 2211
number of functions with 2 return arguments, where the second argument is an error: 1562

percentage of functions where multiple return parameters are really what we want: 3.480453
```