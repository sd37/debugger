package main

import (
	"bufio"
	"debug/elf"
	"debug/gosym"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

var targetfile string
var line int
var pc uint64
var fn *gosym.Func
var symTable *gosym.Table
var regs unix.PtraceRegs
var ws unix.WaitStatus
var originalData []byte
var breakpointSet bool

func catchError(err error) {
	if err != nil {
		panic(err)
	}
}

func execute(target string) {
	var filename string

	cmd := exec.Command(target)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.SysProcAttr = &unix.SysProcAttr{
		Ptrace: true,
	}

	cmd.Start()
	err := cmd.Wait()
	//The first time this will stop because of cloning
	if err != nil {
		fmt.Printf("Wait returned: %v\n\n", err)
	}

	pid := cmd.Process.Pid
	pgid, _ := unix.Getpgid(pid)

	//Allow tracing of the target binary
	catchError(unix.PtraceSetOptions(pid, unix.PTRACE_O_TRACECLONE))

	if getInput(pid) {
		catchError(unix.PtraceCont(pid, 0))
	} else {
		catchError(unix.PtraceSingleStep(pid))
	}

	for {
		wpid, err := unix.Wait4(-1*pgid, &ws, 0, nil)
		catchError(err)
		if ws.Exited() {
			if wpid == pid {
				break
			}
		} else {
			//Breakpoint is set only if we receive a SIGTRAP and teh process hasn't exited
			if ws.StopSignal() == unix.SIGTRAP && ws.TrapCause() != unix.PTRACE_EVENT_CLONE {
				catchError(unix.PtraceGetRegs(wpid, &regs))
				filename, line, fn = symTable.PCToLine(regs.Rip)
				fmt.Printf("Stopped at %s at %d in %s\n", fn.Name, line, filename)
				outputStack(symTable, wpid, regs.Rip, regs.Rsp, regs.Rbp)

				if breakpointSet {
					replaceMachineCode(wpid, pc, originalData)
					breakpointSet = false
				}

				if getInput(wpid) {
					catchError(unix.PtraceCont(wpid, 0))
				} else {
					catchError(unix.PtraceSingleStep(wpid))
				}
			} else {
				catchError(unix.PtraceCont(wpid, 0))
			}
		}
	}
}

func replaceMachineCode(pid int, breakpoint uint64, code []byte) {
	originalData = make([]byte, len(code))
	unix.PtracePeekData(pid, uintptr(breakpoint), originalData)
	unix.PtracePokeData(pid, uintptr(breakpoint), code)
}

func outputStack(symTable *gosym.Table, pid int, ip uint64, sp uint64, bp uint64) {

	_, _, fn = symTable.PCToLine(ip)

	var i uint64
	var nextbp uint64

	for {
		i = 0
		frameSize := bp - sp + 8

		// Read the next stack frame
		b := make([]byte, frameSize)
		_, err := unix.PtracePeekData(pid, uintptr(sp), b)
		if err != nil {
			panic(err)
		}

		// The address to return to is at the top of the frame
		content := binary.LittleEndian.Uint64(b[i : i+8])
		_, lineno, nextfn := symTable.PCToLine(content)
		if nextfn != nil {
			fn = nextfn
			fmt.Printf("  called by %s line %d\n", fn.Name, lineno)
		}

		// Rest of the frame
		for i = 8; sp+i <= bp; i += 8 {
			content := binary.LittleEndian.Uint64(b[i : i+8])
			if sp+i == bp {
				nextbp = content
			}
			fmt.Printf("  %X %X  \n", sp+i, content)
		}

		//Stop the stack trace at the top most level
		if fn.Name == "main.main" || fn.Name == "runtime.main" {
			break
		}

		// Move to the next frame
		sp = sp + i
		bp = nextbp
	}

	fmt.Println()
}

func generateSymbolTable(prog string) {
	exe, err := elf.Open(prog)
	if err != nil {
		panic(err)
	}
	defer exe.Close()

	addr := exe.Section(".text").Addr

	lineTableData, err := exe.Section(".gopclntab").Data()
	if err != nil {
		panic(err)
	}
	lineTable := gosym.NewLineTable(lineTableData, addr)
	if err != nil {
		panic(err)
	}

	symTableData, err := exe.Section(".gosymtab").Data()
	if err != nil {
		panic(err)
	}

	symTable, err = gosym.NewTable(symTableData, lineTable)
	if err != nil {
		panic(err)
	}
}

func getInput(pid int) bool {
	sub := false
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("\n(C)ontinue, (S)tep, set (B)reakpoint or (Q)uit? > ")
	for {
		scanner.Scan()
		input := scanner.Text()
		switch strings.ToUpper(input) {
		case "C":
			return true
		case "S":
			return false
		case "B":
			fmt.Printf("  Enter line number in %s: > ", targetfile)
			sub = true
		case "Q":
			os.Exit(0)
		default:
			if sub {
				line, _ = strconv.Atoi(input)
				setBreakPoint(pid, targetfile, line)
				return true
			}
			fmt.Printf("Please enter again: %s\n", input)
			fmt.Printf("\n(C)ontinue, (S)tep, set (B)reakpoint or (Q)uit? > ")
		}
	}
}

func setBreakPoint(pid int, filename string, line int) {
	var err error
	pc, _, err = symTable.LineToPC(filename, line)
	if err != nil {
		fmt.Printf("Can't find breakpoint for %s, %d\n", filename, line)
		breakpointSet = false
	}

	replaceMachineCode(pid, pc, []byte{0xCC})
	breakpointSet = true

}

func main() {
	target := os.Args[1]
	generateSymbolTable(target)
	fn = symTable.LookupFunc("main.main")
	targetfile, line, fn = symTable.PCToLine(fn.Entry)
	execute(target)
}
