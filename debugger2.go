package main

// import (
// 	"fmt"
// 	"os"
// 	"os/exec"

// 	"golang.org/x/sys/unix"
// )

// var ws unix.WaitStatus
// var regs unix.PtraceRegs

// func catchError(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func main() {

// 	//Start another process that needs to be debugged
// 	target := "hello/hello"

// 	cmd := exec.Command(target)
// 	cmd.Stderr = os.Stderr
// 	cmd.Stdin = os.Stdin
// 	cmd.Stdout = os.Stdout
// 	cmd.SysProcAttr = &unix.SysProcAttr{
// 		Ptrace: true,
// 	}

// 	cmd.Start()

// 	pid := cmd.Process.Pid

// 	// catchError(unix.PtraceAttach(pid))

// 	err := cmd.Wait()
// 	if err != nil {
// 		fmt.Printf("Wait returned: %v\n\n", err)
// 	}

// 	regs := 0x455ac0
// 	catchError(unix.PtraceGetRegs(pid, &regs))
// 	fmt.Println("Regs: ", regs.Rip)

// 	reg_data := make([]byte, 8)
// 	_, err = unix.PtracePeekText(pid, uintptr(regs.Rip), reg_data)
// 	catchError(err)

// 	fmt.Println("Reg Data:", reg_data)

// 	catchError(unix.PtraceCont(pid, 0))

// 	// pid := cmd.Process.Pid
// 	// pgid, _ := unix.Getpgid(pid)
// 	// fmt.Println("PID of hello: ", pid)

// 	// catchError(unix.PtraceSingleStep(pid))

// 	// for {
// 	// catchError(unix.PtraceSingleStep(pid))

// 	// catchError(unix.PtraceGetRegs(pid, &regs))
// 	// fmt.Println("Regs: ", regs.Rsp)
// 	// reg_data := make([]byte, 5)
// 	// _, err := unix.PtracePeekText(pid, uintptr(regs.Rsp), reg_data)
// 	// catchError(err)

// 	// 	// fmt.Println("Reg Data:", reg_data)
// 	// 	//unix.Wait4(-1*pgid, &ws, 0, nil)
// 	// 	//fmt.Println(ws)

// 	// }

// 	fmt.Println("hi")
// }
