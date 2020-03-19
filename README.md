# benchmirror

Little script to update the fastest ubuntu mirror list by whish. By default the one from your region:
http://mirrors.ubuntu.com/mirrors.txt but you can put your own list and benchmark it.
They will all be checked in parrallel as the check procedure has it's own channel.

Really effective by using it with https://github.com/ilikenwf/apt-fast .



```
./benchmirror 
Ubuntu mirror checker
Usage of ./benchmirror:
  -b	start benchmark
  -f string
    	path to file containing the mirror urls
  -l int
    	limit latency, discard mirrors slower then (time in ms)  (default 5000)
  -t int
    	timeout setting for the http calls (time in sec) (default 5)
  -v	enable verbose output
```

Example: `./benchmirror -b -l 100 -t 3`
