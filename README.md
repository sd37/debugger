# debugger
A debugger from scratch in Go.

You can find the slides with the name Debugger_from_scratch_slides.pdf

The go library for ptrace syscall is platform dependent, hence the Dockerfile can be used to run the code. Since ptrace allows one process to look and modify the other process and it's register, the docker needs to run with secure computing off.

docker run -v $PWD:<path_to_mount> --security-opt seccomp=unconfined -it debugger /bin/sh

