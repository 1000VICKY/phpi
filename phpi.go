// +build windows

package main

import (
	"bufio"
	_ "errors"
	"fmt"
	_ "fmt"
	_ "io"
	"io/ioutil"
	"os"
	exe "os/exec"
	"os/signal"
	"path/filepath"
	"phpa/echo"
	"phpa/goroutine"
	"phpa/standardInput"
	_ "reflect"
	_ "regexp"
	"runtime"
	"runtime/debug"
	_ "strings"
	_ "syscall"
	_ "time"

	// 自作パッケージ

	_ "phpa/myreflect"

	"golang.org/x/sys/windows"

	// syscallライブラリの代替ツール

	_ "golang.org/x/sys/unix"
)

func main() {
	var stdin (func(*string) bool) = nil
	var standard *standardInput.StandardInput = new(standardInput.StandardInput)
	standard.SetStandardInputFunction()
	stdin = standard.GetStandardInputFunction()

	// プロセスの監視
	var signal_chan chan os.Signal = make(chan os.Signal)
	// OSによってシグナルのパッケージを変更
	signal.Notify(
		signal_chan,
		os.Interrupt,
		os.Kill,
		windows.SIGKILL,
		windows.SIGHUP,
		windows.SIGINT,
		windows.SIGTERM,
		windows.SIGQUIT,
		windows.Signal(0x13),
		windows.Signal(0x14), // Windowsの場合 SIGTSTPを認識しないためリテラルで指定する
	)

	// シグナルを取得後終了フラグとするチャンネル
	var exit_chan chan int = make(chan int)
	// シグナルを監視
	go goroutine.MonitoringSignal(signal_chan, exit_chan)
	// コンソールを停止するシグナルを握りつぶす
	go goroutine.CrushingSignal(exit_chan)
	go goroutine.RunningFreeOSMemory()

	// 実行するPHPスクリプトの初期化
	// バックティックでヒアドキュメント
	const initializer = "<?php \r\n" +
		"ini_set(\"display_errors\", 1);\r\n" +
		"ini_set(\"error_reporting\", -1);\r\n"

	// 利用変数初期化
	var input string
	var line *string
	line = new(string)

	var tentativeFile *string
	tentativeFile = new(string)

	var writtenByte *int
	writtenByte = new(int)

	var ff *os.File
	var err error
	// ダミー実行ポインタ
	ff, err = ioutil.TempFile("", "__php__main__")
	if err != nil {
		echo.Echo(err.Error() + "\r\n")
		os.Exit(255)
	}
	ff.Chmod(os.ModePerm)
	*writtenByte, err = ff.WriteAt([]byte(initializer), 0)
	if err != nil {
		echo.Echo(err.Error() + "\r\n")
		os.Exit(255)
	}
	// ファイルポインタに書き込まれたバイト数を検証する
	if *writtenByte != len(initializer) {
		echo.Echo("[Couldn't complete process to initialize script file.]\r\n")
		os.Exit(255)
	}
	// ファイルポインタオブジェクトから絶対パスを取得する
	*tentativeFile, err = filepath.Abs(ff.Name())
	if err != nil {
		echo.Echo(err.Error() + "\r\n")
		os.Exit(255)
	}
	defer ff.Close()
	defer os.Remove(*tentativeFile)

	var count int = 0
	//var ss int = 0
	var multiple int = 0
	//var backup []byte = make([]byte, 0)
	var currentDir string

	// saveコマンド入力用
	var saveFp *os.File
	saveFp = new(os.File)

	// 入力されたソースコードをバックグラウンドで検証する
	var syntax chan int
	syntax = make(chan int)
	var errorString chan string
	errorString = make(chan string)

	var fixedInput string
	input = initializer
	fixedInput = input
	var errorMessage string
	for {
		debug.SetGCPercent(100)
		debug.FreeOSMemory()
		runtime.GC()
		// ループ開始時に正常動作するソースのバックアップを取得
		// ff.Seek(0, 0)
		// backup, err = ioutil.ReadAll(ff)
		// if err != nil {
		// 	echo.Echo(err.Error() + "\r\n")
		// 	break
		// }
		// ff.WriteAt(backup, 0)

		if multiple == 1 {
			echo.Echo("(" + errorMessage + ")" + " .... ")
		} else {
			echo.Echo("(" + errorMessage + ")" + "php > ")
		}
		*line = ""

		// 標準入力開始
		stdin(line)
		temp := *line

		if temp == "del" {
			ff, err = deleteFile(ff, initializer)
			if err != nil {
				echo.Echo(err.Error() + "\r\n")
				os.Exit(255)
			}
			*line = ""
			input = initializer
			fixedInput = input
			count = 0
			multiple = 0
			continue
		} else if temp == "save" {
			currentDir, err = os.Getwd()
			currentDir += "\\save.php"
			saveFp, err = os.Create(currentDir)
			if err != nil {
				echo.Echo(err.Error() + "\r\n")
				continue
			}
			saveFp.Chmod(os.ModePerm)
			input = fixedInput
			*writtenByte, err = saveFp.WriteAt([]byte(input), 0)
			if err != nil {
				saveFp.Close()
				echo.Echo(err.Error() + "\r\n")
				os.Exit(255)
			}
			echo.Echo("[" + currentDir + ":Completed saving input code which you wrote.]" + "\r\n")
			saveFp.Close()
			*line = ""
			multiple = 0
			continue
		} else if temp == "exit" {
			// コンソールを終了させる
			echo.Echo("[Would you really like to quit a console which you are running in terminal? yes/or]\r\n")
			var quitText *string
			quitText = new(string)
			stdin(quitText)
			if *quitText == "yes" {
				os.Exit(0)
			} else {
				echo.Echo("[Canceled to quit this console app in terminal.]\r\n")
			}
			*line = ""
			continue
		} else if temp == "restore" {
			input = fixedInput
			multiple = 0
			continue
		} else if temp == "" {
			// 空文字エンターの場合はループを飛ばす
			continue
		}

		input += *line + "\n"

		_, err = ff.WriteAt([]byte(input), 0)
		if err != nil {
			// temporary fileへの書き込みに失敗した場合
			echo.Echo(err.Error())
			continue
		}
		fmt.Println(input)
		// 並行処理でスクリプトが正常実行できるまでループを繰り返す
		go SyntaxCheck(tentativeFile, syntax, errorString)
		// チャンネルから値を取得
		si := <-syntax
		errorMessage = <-errorString
		if si == 1 {
			*line = ""
			fixedInput = input
			count, err = tempFunction(ff, tentativeFile, count)
			if err != nil {
				continue
			}
			multiple = 0
			input += " echo(PHP_EOL);\r\n "
		} else {
			multiple = 1
		}
	}
}

func SyntaxCheck(filePath *string, c chan int, errorString chan string) (bool, error) {
	defer debug.SetGCPercent(100)
	defer runtime.GC()
	defer debug.FreeOSMemory()
	var e error = nil
	var command *exe.Cmd
	// バックグラウンドでPHPをコマンドラインで実行
	command = exe.Command("php", *filePath)
	e = command.Run()
	fmt.Println(e)
	fmt.Println("======")
	if e == nil {
		// コマンド成功時
		c <- 1
		errorString <- command.ProcessState.String()
		return true, nil
	} else {
		// コマンド実行失敗時
		c <- 0
		errorString <- command.ProcessState.String()
		return false, e
	}
}

func tempFunction(fp *os.File, filePath *string, beforeOffset int) (int, error) {
	defer debug.SetGCPercent(100)
	defer runtime.GC()
	defer debug.FreeOSMemory()
	var e error
	command := exe.Command("php", *filePath)
	e = command.Run()
	// バックグラウンドでPHPをコマンドラインで実行
	/*
		command := exe.Command("php", *filePath)
		e = command.Run()
		if e != nil {
			// 実行したスクリプトの終了コードを取得
			var code bool = command.ProcessState.Success()
			if code != true {
				var scanText string = ""
				command = exe.Command("php", *filePath)
				stdout, _ := command.StdoutPipe()
				command.Start()
				scanner := bufio.NewScanner(stdout)
				var ii int = 0
				for scanner.Scan() {
					if ii >= beforeOffset {
						scanText = scanner.Text()
						if len(scanText) > 0 {
							echo.Echo("     " + scanner.Text() + "\r\n")
						}
					}
					ii++
				}
				if beforeOffset > ii {
					command = exe.Command("php", *filePath)
					stdout, _ := command.StdoutPipe()
					command.Start()
					scanner = bufio.NewScanner(stdout)
					for scanner.Scan() {
						scanText = scanner.Text()
						if len(scanText) > 0 {
							echo.Echo("     " + scanner.Text() + "\r\n")
						}
					}
				}
				command.Wait()
				echo.Echo("\r\n")
				fp.Truncate(0)
				fp.Seek(0, 0)
				fp.WriteAt(temporaryBackup, 0)
				command = nil
				stdout = nil
				return beforeOffset, e
			}
		}
	*/
	var ii int = 0
	var scanText string
	// Run()メソッドで利用したcommandオブジェクトを再利用
	command = exe.Command("php", *filePath)
	stdout, ee := command.StdoutPipe()
	if ee != nil {
		echo.Echo(ee.Error() + "\r\n")
		panic("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus.")
	}
	command.Start()
	scanner := bufio.NewScanner(stdout)
	for {
		// 読み取り可能な場合
		if scanner.Scan() == true {
			if ii >= beforeOffset {
				scanText = scanner.Text()
				if len(scanText) > 0 {
					echo.Echo("     " + scanText + "\r\n")
				}
			}
			ii++
		} else {
			break
		}
	}
	command.Wait()
	command = nil
	stdout = nil
	scanText = ""
	echo.Echo("\r\n")
	fp.Write([]byte("echo(PHP_EOL);\r\n"))
	return ii, e
}

func deleteFile(fp *os.File, initialString string) (*os.File, error) {
	defer debug.SetGCPercent(100)
	defer runtime.GC()
	defer debug.FreeOSMemory()
	var err error
	fp.Truncate(0)
	fp.Seek(0, 0)
	_, err = fp.WriteAt([]byte(initialString), 0)
	fp.Seek(0, 0)
	return fp, err
}
