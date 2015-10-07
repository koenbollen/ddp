package main

import "syscall"

// The signal that is send to the `dd` process every 200ms.
const InfoSignal = syscall.SIGUSR1
