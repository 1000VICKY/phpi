package main

import (
    "bufio"
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
)

var format func(...interface{}) (int, error) = fmt.Println
var myPrint func(...interface{}) (int, error) = fmt.Print

func main() {
    // プロセスの監視
    signal_chan := make(chan os.Signal, 1)
    signal.Notify(signal_chan,
        os.Interrupt,
        os.Kill,
        syscall.SIGHUP,
        syscall.SIGINT,
        syscall.SIGTERM,
        syscall.SIGQUIT);

    // シグナルを取得後終了フラグとするチャンネル
    var exit_chan chan int;
    exit_chan = make(chan int)
    go func(sig chan os.Signal, exit chan int) {
        var s os.Signal;
        for {
            s, _ = <-sig;
            if (s == syscall.SIGHUP) {
                fmt.Printf("[syscall.SIGHUP]Input a word `exit`, if you would like to exit this console.\r\n")
                // 割り込みを無視
                exit <- 0;
            } else if (s == os.Kill) {
                fmt.Printf("[os.Kill]Input a word `exit`, if you would like to exit this console.\r\n")
                // 割り込みを無視
                exit <- 0;
            } else if (s == os.Interrupt) {
                fmt.Printf("[os.Interrupt]Input a word `exit`, if you would like to exit this console.\r\n")
                // 割り込みを無視
                exit <- 0;
            } else if (s == syscall.SIGTERM) {
                fmt.Println("Force stop.")
                exit <- 1;
            } else if (s == syscall.SIGQUIT) {
                fmt.Println("Stop and core dump.");
                exit <- 1;
            }
            /*
            switch s {
            case syscall.SIGHUP:
                fmt.Printf("[syscall.SIGHUP]Input a word `exit`, if you would like to exit this console.\r\n")
                // 割り込みを無視
                //exit <- 0;
            case os.Kill:
                fmt.Printf("[os.Kill]Input a word `exit`, if you would like to exit this console.\r\n")
                // 割り込みを無視
                //exit <- 0
            case os.Interrupt:
                fmt.Printf("[os.Interrupt]Input a word `exit`, if you would like to exit this console.\r\n")
                // 割り込みを無視
                //exit <- 0
            case syscall.SIGINT:
                fmt.Printf("[syscall.SIGINT]Input a word `exit`, if you would like to exit this console.\r\n");
                // 割り込みを無視
                //exit <- 0
            case syscall.SIGTERM:
                fmt.Println("force stop")
                exit <- 1
            case syscall.SIGQUIT:
                fmt.Println("stop and core dump")
                exit <- 1
            default:
                fmt.Println("Unknown signal.")
                exit <- 1
            }
            */
        }
    }(signal_chan, exit_chan);

    go func(exit chan int ) {
        var code int = 0;
        for {
            code, _ = <-exit;
            if (code == 1 ) {
                os.Exit(code);
            } else {
                fmt.Print("Ignore signal.");
            }
        }
    }(exit_chan);

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
        os.Exit(255)
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
    // ヒアドキュメントを入力された場合
    var startHereDocument *regexp.Regexp = new (regexp.Regexp);
    startHereDocument, err = regexp.Compile("^.*<<< *([_a-zA-Z0-9]+)$");
    if err != nil {
        format(err)
        os.Exit(255)
    }
    // ヒアドキュメントで入力された場合
    var hereFlag bool = false;
    var ID string = "";
    var endHereDocument *regexp.Regexp = new (regexp.Regexp);
    var scanner *bufio.Scanner = nil
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
        if multiple == 1 {
            myPrint(" ... ")
        } else {
            myPrint("php > ")
        }
        scanner = bufio.NewScanner(os.Stdin)
        scanner.Scan()
        *line = scanner.Text()
        scanner = nil

        // ヒアドキュメントで入力された場合
        if (hereFlag == false) {
            hereTag := startHereDocument.FindAllStringSubmatch(*line, -1);
            if (len(hereTag) > 0 ) {
                if (len(hereTag[0]) > 0) {
                    ID = hereTag[0][1];
                    hereFlag = true;
                }
            } else {
                hereFlag = false;
            }
        } else {
            endHereDocument, err = regexp.Compile("^" + ID + "[ ]*;$");
            if endHereDocument.MatchString(*line) {
                hereFlag = false;
            } else {
                hereFlag = true;
            }
        }


        if *line == "del" {
            ff, err = deleteFile(ff, "<?php ")
            if err != nil {
                format(err)
                os.Exit(255)
            }
            *line = ""
            input = ""
            count = 0;
            debug.FreeOSMemory();
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
            if (hereFlag == true) {
                multiple = 1
            } else if hereFlag == false {
                multiple = 0;
            }
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