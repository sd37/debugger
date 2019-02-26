package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func errors(err error) {
	if err != nil {
		panic(err)
	}
}

func main1() {

	// target := "hello/hello"

	// cmd := exec.Command(target)
	// cmd.Stderr = os.Stderr
	// cmd.Stdin = os.Stdin
	// cmd.Stdout = os.Stdout
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Ptrace: true,
	// }

	// cmd.Start()
	// err := cmd.Wait()
	// if err != nil {
	// 	fmt.Printf("Wait returned: %v\n\n", err)
	// }

	// pid := cmd.Process.Pid
	// fmt.Println("PID of hello: ", pid)
	//_, _ := syscall.Getpgid(pid)

	pid := 5461
	fmt.Println("parenttt")
	errors(unix.PtraceAttach(pid))
	var addr uintptr = 4527539
	data := make([]byte, 5, 5)

	// fmt.Println("PEEKING DATA")
	// _, err := unix.PtracePeekData(pid, addr, data)
	// errors(err)
	fmt.Println("POKING DATA")
	data[0] = 0xCC
	_, err := unix.PtracePokeData(pid, addr, data)
	errors(err)
	fmt.Println(data)
	errors(unix.PtraceSingleStep(pid))
	var ws unix.WaitStatus
	unix.Wait4(pid, &ws, 0, nil)

	// errors(unix.PtraceAttach(pid))
	//errors(unix.PtraceSingleStep(pid))

	// var ws unix.WaitStatus

	fmt.Println("hi")
}
