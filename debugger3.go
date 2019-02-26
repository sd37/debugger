package main

import (
	"bufio"
	"debug/elf"
	"debug/gosym"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

var ws unix.WaitStatus
var regs unix.PtraceRegs

func catchError(err error) {
	if err != nil {
		panic(err)
	}
}

func setBreakPoint(pid int, addr uintptr) []byte {

	originalData := make([]byte, 1)

	_, err := unix.PtracePeekData(pid, addr, originalData)
	catchError(err)

	_, err = unix.PtracePokeData(pid, addr, []byte{0xCC})
	catchError(err)

	tempData := make([]byte, 1)
	_, err = unix.PtracePeekData(pid, addr, tempData)
	fmt.Println("Check if breakpoint was set: ", tempData)

	return originalData
}

func unsetBreakPoint(pid int, addr uintptr, originalData []byte) {
	_, err := unix.PtracePokeData(pid, addr, originalData)
	catchError(err)
}

func getProgramCounter(pid int) uint64 {
	var regs unix.PtraceRegs
	catchError(unix.PtraceGetRegs(pid, &regs))

	return regs.PC()
}

func setProgramCounter(pid int, pc uint64) {
	var regs unix.PtraceRegs
	catchError(unix.PtraceGetRegs(pid, &regs))

	regs.SetPC(pc)
	catchError(unix.PtraceSetRegs(pid, &regs))
}

func decreaseProgramCounter(pid int, pc uint64) {
	setProgramCounter(pid, getProgramCounter(pid)-uint64(1))
}

func singleStep(pid int) {
	catchError(unix.PtraceSingleStep(pid))
}

func continueExecution(pid int) {
	catchError(unix.PtraceCont(pid, 0))
}

func getPCFromLine() {

	executable, err := elf.Open("hello/hello")
	catchError(err)

	//Returns section in elf file
	pcToLineSection := executable.Section(".gopclntab")
	//uncompresses the section data returned by Section
	pcToLineData, err := pcToLineSection.Data()

	//returns tge go symbol table
	symbolTableSection := executable.Section(".gosymtab")
	symbolTableData, err := symbolTableSection.Data()

	lineTableForText := gosym.NewLineTable(pcToLineData, executable.Section(".text").Addr)

	//NewTable decodes the Go symbol table (the ".gosymtab" section in ELF),
	//returning an in-memory representation.
	newSymbolTable, err := gosym.NewTable(symbolTableData, lineTableForText)
	catchError(err)

	// fmt.Println(newSymbolTable.Files)

	pc, fn, err := newSymbolTable.LineToPC("/go/src/debugger/hello/hello.go", 8)
	catchError(err)

	fmt.Println(pc)
	fmt.Println(fn)
}

func getLineFromPC() {
	executable, err := elf.Open("hello/hello")
	catchError(err)

	//Returns section in elf file
	pcToLineSection := executable.Section(".gopclntab")
	//uncompresses the section data returned by Section
	pcToLineData, err := pcToLineSection.Data()

	symbolTableSection := executable.Section(".gosymtab")
	symbolTableData, err := symbolTableSection.Data()

	lineTableForText := gosym.NewLineTable(pcToLineData, executable.Section(".text").Addr)

	//NewTable decodes the Go symbol table (the ".gosymtab" section in ELF),
	//returning an in-memory representation.
	newSymbolTable, err := gosym.NewTable(symbolTableData, lineTableForText)
	catchError(err)

	sym := newSymbolTable.LookupFunc("main.main")
	filename, lineno, _ := newSymbolTable.PCToLine(sym.Entry)

	fmt.Println(filename)
	fmt.Println(lineno)
}

func printRegisters(pid int) {
	var regs unix.PtraceRegs
	catchError(unix.PtraceGetRegs(pid, &regs))

	fmt.Println(regs.Rax)
	fmt.Println(regs.Rdi)
}

func processInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")

	for {
		text, _ := reader.ReadString('\n')

		if strings.Contains(text, "breakpoint") {
			//lineNumber := parseLineNumber(text)

		} else if strings.Contains(text, "stop") {

		} else if strings.Contains(text, "cont") {

		}
	}
}

func main() {

	getLineFromPC()

	getPCFromLine()

	//Start another process that needs to be debugged
	target := "hello/hello"

	cmd := exec.Command(target)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.SysProcAttr = &unix.SysProcAttr{
		Ptrace: true,
	}

	cmd.Start()

	err := cmd.Wait()
	log.Printf("State: %v\n", err)

	pid := cmd.Process.Pid

	ppid := os.Getppid()

	catchError(unix.PtraceSetOptions(pid, syscall.PTRACE_O_TRACECLONE))

	breakpoint := uintptr(getProgramCounter(pid) + 2)
	original := setBreakPoint(pid, breakpoint)

	continueExecution(pid)

	_, err = unix.Wait4(-1*ppid, &ws, unix.WALL, nil)
	catchError(err)

	printRegisters(pid)

	unsetBreakPoint(pid, breakpoint, original)

	setProgramCounter(pid, uint64(breakpoint))
	singleStep(pid)

	_, err = unix.Wait4(-1*ppid, &ws, unix.WALL, nil)
	catchError(err)

	printRegisters(pid)

	// catchError(unix.PtraceSetOptions(pid, unix.PTRACE_O_TRACECLONE))

	// catchError(unix.PtraceSingleStep(pid))

	// steps := 1
	// for {
	// 	pid, err = unix.Wait4(-1*ppid, &ws, unix.WALL, nil)
	// 	catchError(err)

	// 	if pid == -1 {
	// 		catchError(err)
	// 	}

	// 	if pid == cmd.Process.Pid && ws.Exited() {
	// 		break
	// 	}

	// 	if !ws.Exited() {
	// 		catchError(unix.PtraceSingleStep(pid))
	// 		steps += 1
	// 	}
	// }

	// fmt.Println("Steps: ", steps)

}
