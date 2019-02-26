package main

import (
	"debug/elf"
	"debug/gosym"
	"fmt"
	"log"
	"os"
	"os/exec"

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
	var regs unix.PtraceRegs

	originalData := make([]byte, 1)

	_, err := unix.PtracePeekData(pid, addr, originalData)
	catchError(err)

	_, err = unix.PtracePokeData(pid, addr, []byte{0xCC})
	catchError(err)

	return originalData
}

func unsetBreakPoint(pid int, addr uintptr, originalData []byte) {
	_, err := unix.PtracePokeData(pid, addr, originalData)
	catchError(err)
}

func getProgramCounter(pid int) {
	var regs unix.PtraceRegs
	catchError(unix.PtraceGetRegs(pid, &regs))

	return regs.PC()
}

func setProgramCounter(pid int, pc uint64) {
	var regs unix.PtraceRegs
	catchError(unix.PtraceGetRegs(pid, &regs))

	regs.SetPC(pc)
	catchError(unix.PtraceSetRegs())
}

func decreaseProgramCounter(pid int) {
	var regs unix.PtraceGetRegs
	setProgramCounter(getProgramCounter(pid) - 1)
}

func singleStep(pid int) {
	catchError(unix.PtraceSingleStep(pid))
}

func continueExecution(pid int) {
	catchError(unix.PtraceCont(pid, 0))
}

func getLineFromPC() {
	executable, err := elf.Open("hello/hello")
	catchError(err)

	//Returns section in elf file
	pcToLineSection := executable.Section(".gopclntab")
	//uncompresses the section data returned by Section
	pcToLineData, err := pcToLineSection.Data()

	symbolTableSection := executable.SectionByType(elf.SHT_PROGBITS)
	symbolTableSection = executable.Section(".gosymtab")
	fmt.Println(symbolTableSection)
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

func main() {

	getLineFromPC()

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
	fmt.Println("ppid from os: ", ppid)

	pgid, err1 := unix.Getpgid(pid)
	catchError(err1)
	fmt.Println("Get pgid from os: ", pgid)

	catchError(unix.PtraceSetOptions(pid, unix.PTRACE_O_TRACECLONE))

	catchError(unix.PtraceSingleStep(pid))

	steps := 1
	for {
		pid, err = unix.Wait4(-1*ppid, &ws, unix.WALL, nil)
		catchError(err)

		if pid == -1 {
			catchError(err)
		}

		if pid == cmd.Process.Pid && ws.Exited() {
			break
		}

		if !ws.Exited() {
			catchError(unix.PtraceSingleStep(pid))
			steps += 1
		}
	}

	fmt.Println("Steps: ", steps)

}
