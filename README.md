# histogram

Render histograms in the terminal.


```
$ histogram -h
Usage:
  histogram [OPTIONS]

Application Options:
  -b, --bins=  Number of bins in the histogram (default: 10)

Help Options:
  -h, --help   Show this help message
```

`histogram` render histograms from the given list of numbers.


```
$ head normaldistribution.txt
0.4
1.98
-0.01
1.46
0.5
0.72
-0.66
-0.29
1.75
-0.96

$ cat normaldistribution.txt | histogram
Total count = 1000
Min/Avg/Max = -3.25 / -0.01 / 3.23

 [ -3.25,  -2.60 ]      2  
 [ -2.60,  -1.95 ]     19  |||
 [ -1.95,  -1.31 ]     77  ||||||||||||
 [ -1.31,  -0.66 ]    166  |||||||||||||||||||||||||||
 [ -0.66,  -0.01 ]    236  ||||||||||||||||||||||||||||||||||||||
 [ -0.01,   0.64 ]    243  ||||||||||||||||||||||||||||||||||||||||
 [  0.64,   1.29 ]    160  ||||||||||||||||||||||||||
 [  1.29,   1.93 ]     75  ||||||||||||
 [  1.93,   2.58 ]     18  ||
 [  2.58,   3.23 ]      4  
```
