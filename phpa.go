package main

import (
//  "bufio"
    "errors"
    "fmt"
    _ "io"
    "io/ioutil"
    "os"
    exe "os/exec"
    "os/signal"
    "path/filepath"
    "regexp"
    "runtime"
    "runtime/debug"
    "strings"
    "syscall"
    //"github.com/nemith/goline"
    "github.com/chzyer/readline"
)

var format func(...interface{}) (int, error) = fmt.Println
var myPrint func(...interface{}) (int, error) = fmt.Print

// 入力履歴を保持する
var inputList map[string]string;

func main() {


    // プロセスの監視
    signal_chan := make(chan os.Signal, 1)
    signal.Notify(signal_chan,
        os.Interrupt,
        syscall.SIGHUP,
        syscall.SIGINT,
        syscall.SIGTERM,
        syscall.SIGQUIT)
    exit_chan := make(chan int)
    go func() {
        for {
            s := <-signal_chan
            switch s {
            case syscall.SIGHUP:
                fmt.Printf("\r\nシェルを終了させるには<exit>と入力してください\r\n")
            case syscall.SIGINT:
                fmt.Printf("\r\nシェルを終了させるには<exit>と入力してください\r\n")
            case syscall.SIGTERM:
                fmt.Println("force stop")
                exit_chan <- 0
            case syscall.SIGQUIT:
                fmt.Println("stop and core dump")
                exit_chan <- 0
            default:
                fmt.Println("Unknown signal.")
                exit_chan <- 1
            }
        }
    }()
    const initializer = "<?php " + "\n"
    // 利用変数初期化
    var input string
    var line *string = new(string)
    var ff *os.File
    var err error
    var tentativeFile *string = new(string)
    var writtenByte *int = new(int)
    // ダミー実行ポインタ
    ff, err = ioutil.TempFile("", "__php__main__")
    if err != nil {
        format(err)
        os.Exit(255)
    }
    ff.Chmod(os.ModePerm)
    *writtenByte, err = ff.WriteAt([]byte(initializer), 0)
    if err != nil {
        format(err)
        os.Exit(255)
    }
    // ファイルポインタに書き込まれたバイト数を検証する
    if *writtenByte != len(initializer) {
        format("<スクリプトファイルの初期化に失敗しました.>")
        os.Exit(255)
    }
    // ファイルポインタオブジェクトから絶対パスを取得する
    *tentativeFile, err = filepath.Abs(ff.Name())
    if err != nil {
        format(err)
        os.Exit(255);
    }
    defer ff.Close()
    defer os.Remove(*tentativeFile)

    var count int = 0
    var ss int = 0
    var multiple int = 0
    var backup []byte = make([]byte, 0)
    var currentDir string

    // 末尾はバックスラッシュの場合，以降再びバックスラッシュで終わるまで
    // スクリプトを実行しない
    var openBrace *regexp.Regexp = new(regexp.Regexp)
    var openCount int = 0
    var closeBrace *regexp.Regexp = new(regexp.Regexp)
    var closeCount int = 0
    openBrace, _ = regexp.Compile("^.*{[ \t]*.*$")
    closeBrace, _ = regexp.Compile("^.*}.*$")
    // [save]というキーワードを入力した場合の正規表現
    var saveRegex *regexp.Regexp = new(regexp.Regexp)
    saveRegex, err = regexp.Compile("^[ ]*save[ ]*$")
    if err != nil {
        format(err)
        os.Exit(255)
    }
    //var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
    inputList = make(map[string]string);
    var promptMessage *string = new(string);
    for {
        runtime.GC()
        // ループ開始時に正常動作するソースのバックアップを取得
        ff.Seek(0, 0)
        backup, err = ioutil.ReadAll(ff)
        if err != nil {
            format(err.Error())
            break
        }

        ff.WriteAt(backup, 0)
        *line = "";
        // コマンドライン入力記入
        if (multiple == 1) {
            *promptMessage = "\033[32m ... >\033[0m ";
        } else {
            *promptMessage = "\033[32m php >\033[0m ";
        }
        l, _ := readline.NewEx(&readline.Config{
            Prompt:          *promptMessage,
            HistoryFile:     "/tmp/readline.tmp",
    //      AutoComplete:    completer,
    //        InterruptPrompt: "^C",
    //        EOFPrompt:       "exit",

    //      HistorySearchFold:   true,
    //      FuncFilterInputRune: filterInput,
        });
        if multiple == 1 {
/*
            gl := goline.NewGoLine(goline.StringPrompt("... "))
            data, _ := gl.Line();
            *line = data;
*/
            *line, _ = l.Readline();
        } else {
/*
            gl := goline.NewGoLine(goline.StringPrompt("php > "))
            data, _ := gl.Line();
            *line = data;
*/
            *line, _ =  l.Readline();
        }
        //scanner.Scan()
        //*line = scanner.Text()
        if *line == "del" {
            ff, err = deleteFile(ff, "<?php ")
            if err != nil {
                format(err)
                os.Exit(255)
            }
            *line = ""
            input = ""
            count = 0
            continue
        } else if saveRegex.MatchString(*line) {
            /*
               [save]コマンドが入力された場合，その時点まで入力されたスクリプトを
               カレントディレクトリに保存する
            */
            currentDir, err = os.Getwd()
            if err != nil {
                format(err)
                os.Exit(255)
            }
            currentDir, err = filepath.Abs(currentDir)
            if err != nil {
                format(err)
                break
            }
            if runtime.GOOS == "windows" {
                currentDir += "\\save.php"
            } else {
                currentDir += "/save.php"
            }
            saveFp := new(os.File)
            saveFp, err = os.Create(currentDir)
            if err != nil {
                format(err)
                continue
            }
            saveFp.Chmod(os.ModePerm)
            *writtenByte, err = saveFp.WriteAt(backup, 0)
            if err != nil {
                saveFp.Close()
                fmt.Println(err)
                os.Exit(255)
            }
            format(currentDir + ":入力した内容を保存しました。")
            saveFp.Close()
            *line = ""
            input = ""
            continue
        } else if *line == "exit" {
            os.Exit(0)
        } else if *line == "" {
            // 空文字エンターの場合はループを飛ばす
//          fmt.Println("\n");
            continue
        }

        //ob := openBrace.MatchString(*line)
        ob := openBrace.FindAllStringSubmatch(*line, -1)
        if len(ob) > 0 {
            if len(ob[0]) > 0 {
                //if ob == true {
                openCount = openCount + len(ob[0])
            }
        }
        //cb := closeBrace.MatchString(*line)
        cb := closeBrace.FindAllStringSubmatch(*line, -1)
        if len(cb) > 0 {
            if len(cb[0]) > 0 {
                //if cb == true {
                closeCount = closeCount + len(cb[0])
            }
        }
        // ブレースによる複数入力フラグがfalseの場合
        if openCount == 0 && closeCount == 0 {
            multiple = 0
        } else if openCount != closeCount {
            multiple = 1
        } else if openCount == closeCount {
            multiple = 0
            openCount = 0
            closeCount = 0
        } else {
            panic("Runtime Error happened!:")
        }
        input += *line + "\n"
        if multiple == 0 {
            ss, err = ff.Write([]byte(input))
            if err != nil {
                format("<ファイルポインタへの書き込み失敗>")
                format("    " + err.Error())
                continue
            }
            if ss > 0 {
                input = ""
                *line = ""
                count, err = tempFunction(ff, tentativeFile, count, backup)
                if err != nil {
                    continue
                }
            }
        } else if multiple == 1 {
            continue
        } else {
            panic("<Runtime Error>")
        }
    }
    code := <-exit_chan
    os.Exit(code)
}

func tempFunction(fp *os.File, filePath *string, beforeOffset int, temporaryBackup []byte) (int, error) {
    defer debug.FreeOSMemory()
    var strOutput []string = make([]string, 0)
    var output []byte = make([]byte, 0)
    var e error = nil
    var index *int = new(int)
    runtime.GC()
    // バックグラウンドでPHPをコマンドラインで実行
    // (1)まずは終了コードを取得
    e = exe.Command("php", *filePath).Run()
    if e != nil {
        fmt.Println("終了コードが0以外")
        var ok bool
        var exitError *exe.ExitError = nil
        var exitStatus int = 0
        if exitError, ok = e.(*exe.ExitError); ok == true {
            if s, ok := exitError.Sys().(syscall.WaitStatus); ok == true {
                exitStatus = s.ExitStatus()
                if exitStatus != 0 {
                    // スクリプトを実行した結果、実行失敗の場合
                    output, e = exe.Command("php", *filePath).Output()
                    castStr := string(output)
                    // 改行で区切って[]string型に代入する
                    strOutput = strings.Split(castStr, "\n")
                    if len(strOutput) >= beforeOffset {
                        strOutput = strings.Split(castStr, "\n")[beforeOffset:]
                    }
                    for key, value := range strOutput {
                        fmt.Println("    " + value)
                        strOutput[key] = ""
                    }
                    fmt.Println("    " + e.Error())
                    fp.Truncate(0)
                    fp.Seek(0, 0)
                    fp.WriteAt(temporaryBackup, 0)
                    debug.FreeOSMemory()
                    return beforeOffset, e
                }
            } else {
                panic(errors.New("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus."))
            }
        }
    }
    output, _ = exe.Command("php", *filePath).Output()
    /*
        var s string = string(output);
        var sarray []string = strings.Split(s, "\n");
    */
    strOutput = strings.Split(string(output), "\n")
    if len(strOutput) < beforeOffset {
        beforeOffset = len(strOutput)
    }
    strOutput = strOutput[beforeOffset:]
    *index = len(strOutput) + beforeOffset
    for key, value := range strOutput {
        fmt.Println("    " + value)
        strOutput[key] = ""
    }
    output = nil
    strOutput = nil
    fp.Write([]byte("echo(PHP_EOL);"))
    // プログラムが確保したメモリを強制的にOSへ返却
    debug.FreeOSMemory()
    //fmt.Println("    " + strconv.Itoa(*index) + "byte")
    return *index, e
}

func deleteFile(fp *os.File, initialString string) (*os.File, error) {
    defer debug.FreeOSMemory()
    var size int
    var err error
    fp.Truncate(0)
    fp.Seek(0, 0)
    size, err = fp.WriteAt([]byte(initialString), 0)
    fp.Seek(0, 0)
    if err == nil && size == len(initialString) {
        return fp, err
    } else {
        return fp, errors.New("一時ファイルの初期化に失敗しました。")
    }
}
