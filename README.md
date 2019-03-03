# debugger :beetle: :hammer:
A debugger from scratch in Go, for Go binaries and by a Gopher.

You can find the slides with the name Debugger_from_scratch_slides.pdf

The go library for ptrace syscall is platform dependent(will only run on linux), hence the Dockerfile can be used to run the code. Since ptrace allows one process to look and modify the other process and it's register, the docker needs to run with secure computing off.

```
docker run -v $PWD:<path_to_mount> --security-opt seccomp=unconfined -it debugger /bin/sh
```
Compile target binary using: 
```
go build hello.go
```

Compile the debugger using:
```
go build debugger.go
```

To run:
```
./debugger <target_binary>
```

Assumption: There is only one source file in the target binary.
