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
);
import _"time";
import _ "reflect";

// 自作パッケージ
import "./goroutine";

var format func(...interface{}) (int, error) = fmt.Println
var myPrint func(...interface{}) (int, error) = fmt.Print

func main() {
    // プロセスの監視
    signal_chan := make(chan os.Signal);
    signal.Notify(
        signal_chan,
        os.Interrupt,
        os.Kill,
        syscall.SIGHUP,
        syscall.SIGINT,
        syscall.SIGTERM,
        syscall.Signal(0x14),
        syscall.SIGQUIT,
    );

    // シグナルを取得後終了フラグとするチャンネル
    var exit_chan chan int = make(chan int );
    // シグナルを監視
    go myPackage.MonitoringSignal(signal_chan, exit_chan);
    // コンソールを停止するシグナルを握りつぶす
    go myPackage.CrushingSignal(exit_chan);
    go myPackage.RunningFreeOSMemory();

    const initializer = "<?php " + "\n"
    // 利用変数初期化
    var input string
    var line *string = new(string)
    var ff *os.File
    var err error;
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
    // マッチしたヒアドキュメントタグを取得するため
    var hereTag [][]string = make([][]string, 1);
    var ID string = "";
    var endHereDocument *regexp.Regexp = new (regexp.Regexp);
    var scanner *bufio.Scanner = nil;
    for {
        debug.SetGCPercent(100);
        runtime.GC();
        debug.FreeOSMemory();
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
        scanner = bufio.NewScanner(os.Stdin);
        scanner.Scan();
        *line = scanner.Text();

        // ヒアドキュメントで入力された場合
        if (hereFlag == false) {
            hereTag = startHereDocument.FindAllStringSubmatch(*line, -1);
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
            continue;
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
                break;
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
    debug.SetGCPercent(100);
    runtime.GC();
    defer debug.FreeOSMemory()
    var strOutput []string = make([]string, 0)
    var output []byte = make([]byte, 0)
    var e error = nil
    var index *int = new(int)
    // バックグラウンドでPHPをコマンドラインで実行
    // (1)まずは終了コードを取得
    e = exe.Command("php", *filePath).Run()
    if e != nil {
        fmt.Println("終了コードが0以外")
        var ok bool = true;
        var exitError *exe.ExitError = nil
        var exitStatus int = 0;
        var s syscall.WaitStatus;
        // 型アサーション
        exitError, ok = e.(*exe.ExitError);
        if (ok == true) {
            // 型アサーション
            s, ok = exitError.Sys().(syscall.WaitStatus);
            if (ok == true) {
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
                        fmt.Println("     " + value)
                        strOutput[key] = ""
                    }
                    fmt.Println("     " + e.Error())
                    fp.Truncate(0)
                    fp.Seek(0, 0)
                    fp.WriteAt(temporaryBackup, 0);
                    return beforeOffset, e
                }
            } else {
                panic(errors.New("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus."))
            }
        }
    }
    output, _ = exe.Command("php", *filePath).Output()
    strOutput = strings.Split(string(output), "\n")
    if len(strOutput) < beforeOffset {
        beforeOffset = len(strOutput)
    }
    strOutput = strOutput[beforeOffset:]
    *index = len(strOutput) + beforeOffset;
    var value *string = new(string);
    for _, *value = range strOutput {
        fmt.Print("     " + *value + "\r\n");
        *value = "";
    }
    output = nil
    strOutput = nil;
    fp.Write([]byte("echo(PHP_EOL);"))
    return *index, e
}

func deleteFile(fp *os.File, initialString string) (*os.File, error) {
    runtime.GC();
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