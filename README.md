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

<div style='text-align:center; margin:auto;'>
<a href='http://www.recurse.com' title='Made with love at the Recurse Center'><img src='https://cloud.githubusercontent.com/assets/2883345/11322973/9e557144-910b-11e5-959a-8fdaaa4a88c5.png' height='14px'/></a>
</div>
