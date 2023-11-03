# histgram

Render histograms in the terminal.


```
$ histgram -h
Usage:
  histogram [OPTIONS]

Application Options:
  -b, --bins=  Number of bins in the histogram (default: 10)
  -c, --chart  Draw a bar chart

Help Options:
  -h, --help   Show this help message
```

`histogram` render histograms from the given list of numbers.


```
$ head normaldistribution.txt
-2.17009037776025
1.811910672953939
-1.6953977102723776
-1.5981931399228735
-1.5141018876192867
1.4395314709384586
-1.3722038089987159
2.575829326545648
-1.9599639845509635
-1.3105791121681314

$ cat normaldistribution.txt | histgram
Total count = 300
Min/Avg/Max = -2.58 / 0.00 / 2.58

[   -2.58 -    -2.06]	       6
[   -2.06 -    -1.55]	      12
[   -1.55 -    -1.03]	      27
[   -1.03 -    -0.52]	      45
[   -0.52 -    -0.00]	      60
[    0.00 -     0.52]	      60
[    0.52 -     1.03]	      45
[    1.03 -     1.55]	      27
[    1.55 -     2.06]	      12
[    2.06 -     2.58]	       6
```

`histogram` can also draw a bar chart of the histogram..

```
$ cat normaldistribution.txt | histgram -c
Total count = 300
Min/Avg/Max = -2.58 / 0.00 / 2.58

[   -2.58 -    -2.06]	       6  ||||
[   -2.06 -    -1.55]	      12  ||||||||
[   -1.55 -    -1.03]	      27  ||||||||||||||||||
[   -1.03 -    -0.52]	      45  ||||||||||||||||||||||||||||||
[   -0.52 -    -0.00]	      60  ||||||||||||||||||||||||||||||||||||||||
[    0.00 -     0.52]	      60  ||||||||||||||||||||||||||||||||||||||||
[    0.52 -     1.03]	      45  ||||||||||||||||||||||||||||||
[    1.03 -     1.55]	      27  ||||||||||||||||||
[    1.55 -     2.06]	      12  ||||||||
[    2.06 -     2.58]	       6  ||||
```
