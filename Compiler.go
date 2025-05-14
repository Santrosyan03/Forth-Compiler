package main

import (
        "fmt"
        "io/ioutil"
        "os"
        "os/exec"
        "path/filepath"
        "regexp"
        "strconv"
        "strings"
)

type Compiler struct {
        words        map[string]func(*Compiler, []string, *int) error
        variables    map[string]int
        labelCounter int
        output       []string
        dataSegment  []string
        bssSegment   []string
}

func NewCompiler() *Compiler {
        c := &Compiler{
                variables: make(map[string]int),
                output:    []string{},
                dataSegment: []string{
                        "section .data",
                        "fmt_str db \"%d \",0",
                        "stack_msg db \"Stack: \",0",
                        "stack_newline db \"\",10,0",
                        "empty_stack_msg db \"<empty>\",10,0",
                },
                bssSegment: []string{
                        "section .bss",
                        "stack_start resd 1",
                },
        }
        c.words = map[string]func(*Compiler, []string, *int) error{
                "+":        (*Compiler).compileAdd,
                "-":        (*Compiler).compileSub,
                "*":        (*Compiler).compileMul,
                "dup":      (*Compiler).compileDup,
                "swap":     (*Compiler).compileSwap,
                "tuck":     (*Compiler).compileTuck,
                "neg":      (*Compiler).compileNeg,
                "drop":     (*Compiler).compileDrop,
                "over":     (*Compiler).compileOver,
                "mod":      (*Compiler).compileMod,
                "nip":      (*Compiler).compileNip,
                ".":        (*Compiler).compileDot,
                ".s":       (*Compiler).compileDotS,
                "!":        (*Compiler).compileStore,
                "@":        (*Compiler).compileFetch,
                "variable": (*Compiler).compileVariable,
        }
        return c
}

func (c *Compiler) preprocess(src string) string {
        re := regexp.MustCompile(`\\[^\n]*`)
        src = re.ReplaceAllString(src, "")
        return strings.Join(strings.Fields(src), " ")
}

func (c *Compiler) newLabel(prefix string) string {
        c.labelCounter++
        return fmt.Sprintf("%s_%d", prefix, c.labelCounter)
}

func (c *Compiler) compileVariable(args []string, i *int) error {
        if *i+1 >= len(args) {
                return fmt.Errorf("missing variable name after 'variable'")
        }
        name := args[*i+1]
        if _, exists := c.variables[name]; exists {
                return fmt.Errorf("variable '%s' already exists", name)
        }
        c.variables[name] = len(c.bssSegment)
        c.bssSegment = append(c.bssSegment, fmt.Sprintf("%s: resd 1", name))
        *i += 2
        return nil
}

func (c *Compiler) compileLiteral(n int) {
        c.output = append(c.output, fmt.Sprintf("    push %d", n))
}

func (c *Compiler) compileVariableRef(name string) {
        c.output = append(c.output, fmt.Sprintf("    push %s", name))
}

func (c *Compiler) compileDup(args []string, i *int) error {
        c.output = append(c.output, "    mov eax, [esp]")
        c.output = append(c.output, "    push eax")
        *i++
        return nil
}

func (c *Compiler) compileDrop(args []string, i *int) error {
        c.output = append(c.output, "    add esp, 4")
        *i++
        return nil
}

func (c *Compiler) compileSwap(args []string, i *int) error {
        c.output = append(c.output,
                "    pop eax",
                "    pop ebx",
                "    push eax",
                "    push ebx",
        )
        *i++
        return nil
}

func (c *Compiler) compileAdd(args []string, i *int) error {
        c.output = append(c.output,
                "    pop eax",
                "    add [esp], eax",
        )
        *i++
        return nil
}

func (c *Compiler) compileSub(args []string, i *int) error {
        c.output = append(c.output,
                "    pop ebx",
                "    pop eax",
                "    sub eax, ebx",
                "    push eax",
        )
        *i++
        return nil
}

func (c *Compiler) compileMul(args []string, i *int) error {
        c.output = append(c.output,
                "    pop eax",
                "    pop ebx",
                "    imul eax, ebx",
                "    push eax",
        )
        *i++
        return nil
}

func (c *Compiler) compileTuck(args []string, i *int) error {
        if err := c.compileSwap(args, i); err != nil {
                return err
        }
        if err := c.compileOver(args, i); err != nil {
                return err
        }
        return nil
}

func (c *Compiler) compileNeg(args []string, i *int) error {
        c.output = append(c.output, "    neg dword [esp]")
        *i++
        return nil
}

func (c *Compiler) compileOver(args []string, i *int) error {
        c.output = append(c.output,
                "    mov eax, [esp+4]",
                "    push eax",
        )
        *i++
        return nil
}

func (c *Compiler) compileMod(args []string, i *int) error {
        c.output = append(c.output,
                "    xor edx, edx",
                "    pop ebx",
                "    pop eax",
                "    idiv ebx",
                "    push edx",
        )
        *i++
        return nil
}

func (c *Compiler) compileNip(args []string, i *int) error {
        if err := c.compileSwap(args, i); err != nil {
                return err
        }
        if err := c.compileDrop(args, i); err != nil {
                return err
        }
        return nil
}

func (c *Compiler) compileDot(args []string, i *int) error {
        c.output = append(c.output,
                "    push dword [esp]",
                "    push fmt_str",
                "    call printf",
                "    add esp, 8",
                "    push stack_newline",
                "    call printf",
                "    add esp, 4",
                "    pop eax",
        )
        *i++
        return nil
}

func (c *Compiler) compileDotS(args []string, i *int) error {
        startLabel := c.newLabel("print_stack")
        endLabel := c.newLabel("end_stack")
        c.output = append(c.output,
                "    push stack_msg",
                "    call printf",
                "    add esp, 4",
                "    mov ecx, esp",
                fmt.Sprintf("%s:", startLabel),
                "    cmp ecx, [stack_start]",
                fmt.Sprintf("    jae %s", endLabel),
                "    push ecx",
                "    push dword [ecx]",
                "    push fmt_str",
                "    call printf",
                "    add esp, 8",
                "    pop ecx",
                "    add ecx, 4",
                fmt.Sprintf("    jmp %s", startLabel),
                fmt.Sprintf("%s:", endLabel),
                "    push stack_newline",
                "    call printf",
                "    add esp, 4",
        )
        *i++
        return nil
}

func (c *Compiler) compileStore(args []string, i *int) error {
        c.output = append(c.output,
                "    pop ebx",
                "    pop eax",
                "    mov [ebx], eax",
        )
        *i++
        return nil
}

func (c *Compiler) compileFetch(args []string, i *int) error {
        c.output = append(c.output,
                "    pop eax",
                "    mov eax, [eax]",
                "    push eax",
        )
        *i++
        return nil
}

func (c *Compiler) Compile(inputFile, outputFile string) error {
        data, err := ioutil.ReadFile(inputFile)
        if err != nil {
                return err
        }
        src := c.preprocess(string(data))
        tokens := strings.Split(src, " ")
        for i := 0; i < len(tokens); {
                tok := tokens[i]
                if fn, ok := c.words[tok]; ok {
                        if tok == "variable" {
                                if err := fn(c, tokens, &i); err != nil {
                                        return err
                                }
                        } else {
                                if err := fn(c, tokens, &i); err != nil {
                                        return err
                                }
                        }
                } else if n, err := strconv.Atoi(tok); err == nil {
                        c.compileLiteral(n)
                        i++
                } else if _, exists := c.variables[tok]; exists {
                        c.compileVariableRef(tok)
                        i++
                } else {
                        return fmt.Errorf("unknown word or invalid number: %s", tok)
                }
        }
        return c.writeAsm(outputFile)
}

func (c *Compiler) writeAsm(filename string) error {
        asm := []string{
                "global main",
                "extern printf",
                "section .text",
                "main:",
                "    mov [stack_start], esp",
        }
        asm = append(asm, c.output...)
        asm = append(asm,
                "    xor eax, eax",
                "    ret",
        )
        asm = append(asm, c.dataSegment...)
        asm = append(asm, c.bssSegment...)

        return ioutil.WriteFile(filename, []byte(strings.Join(asm, "\n")), 0644)
}

func runCommand(name string, args ...string) error {
        cmd := exec.Command(name, args...)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        return cmd.Run()
}

func main() {
        if len(os.Args) < 2 {
                fmt.Printf("Usage: %s <tests.fs>\n", os.Args[0])
                os.Exit(1)
        }
        compiler := NewCompiler()
        inputFile := os.Args[1]

        ext := filepath.Ext(inputFile)
        base := inputFile[:len(inputFile)-len(ext)]
        outputFile := base + ".s"

        if err := compiler.Compile(inputFile, outputFile); err != nil {
                fmt.Fprintf(os.Stderr, "Compile error: %v\n", err)
                os.Exit(1)
        }

        fmt.Printf("Compiled %s to %s\n", inputFile, outputFile)

        fmt.Println("Assembling...")
        if err := runCommand("nasm", "-felf32", outputFile, "-o", "Compiler.o"); err != nil {
                fmt.Fprintf(os.Stderr, "Error assembling: %v\n", err)
                os.Exit(1)
        }

        fmt.Println("Linking...")
        if err := runCommand("gcc", "-m32", "Compiler.o", "-o", "Compiler", "-no-pie"); err != nil {
                fmt.Fprintf(os.Stderr, "Error linking: %v\n", err)
                os.Exit(1)
        }

        fmt.Println("Build complete. Running the program...\n")
        if err := runCommand("./Compiler"); err != nil {
                fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
                os.Exit(1)
        }
}
