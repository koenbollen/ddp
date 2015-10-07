ddp
===
`dd` with a progress bar.

- - -

_ddp_ is a small utility that wraps around your local _dd_ command and displays
a progress bar while it is running.

This util starts your local _dd_ as a child process and periodically sends
it the INFO signal (USR1 on Linux). This makes _dd_ output it's progress in
bytes which is parsed and used to update the progress bar.

The commandline arguments are parsed to guess the target filesize, it uses the
filesize specified by the **if=** option or the **count=** option multiplied by
the **bs=** option (the blocksize).


Installation
------------

#### Use golang:
```
$ go get -u github.com/koenbollen/ddp
```


Example
-------

A code-snippet says more then a thousand words:

```
$ ddp if=/dev/zero of=./testfile bs=1m count=1000
218.00 MB / 1000.00 MB [=======>--------------------------] 21.80 % 1.27 GB/s 2s

$ ddp if=./testfile >/dev/null
400.00 MB / 400.00 MB [=================================] 100.00 % 836.37 MB/s
```
