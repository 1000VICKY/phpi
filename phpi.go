// +build windows

package main

import (
	"bufio"
	_ "errors"
	"io"
	_ "io"
	"io/ioutil"
	"os"
	exe "os/exec"
	"os/signal"
	"path/filepath"
	. "phpi/echo"
	"phpi/goroutine"
	"phpi/standardInput"
	_ "reflect"
	_ "regexp"
	"strconv"
	_ "strings"
	"sync"
	"syscall"
	_ "syscall"
	_ "time"

	// 自作パッケージ

	_ "phpi/myreflect"

	"golang.org/x/sys/windows"

	// syscallライブラリの代替ツール

	_ "golang.org/x/sys/unix"
)

// 実行するPHPスクリプトの初期化
// バックティックでヒアドキュメント
const initializer = "<?php \r\n" +
	"ini_set(\"display_errors\", 1);\r\n" +
	"ini_set(\"error_reporting\", -1);\r\n"

var echo func(interface{}) (int, error)
var stdin (func(*string) bool)
var standard *standardInput.StandardInput

// *sync.WaitGroupを使ったスレッド処理
var wg *sync.WaitGroup = new(sync.WaitGroup)

// channelを使ったスレッド処理
var c chan int = make(chan int)
var cc chan int = make(chan int)

// グローパルなboolean型
var commonBool bool = false

func main() {
	echo = Echo()
	// 標準入力を取得するための関数オブジェクトを作成
	standard = new(standardInput.StandardInput)
	standard.SetStandardInputFunction()
	standard.SetBufferSize(1024 * 2)
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
	// 平行でGCを実施
	go goroutine.RunningFreeOSMemory()

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
		echo(err.Error() + "\r\n")
		os.Exit(255)
	}
	ff.Chmod(os.ModePerm)
	*writtenByte, err = ff.WriteAt([]byte(initializer), 0)
	if err != nil {
		echo(err.Error() + "\r\n")
		os.Exit(255)
	}
	// ファイルポインタに書き込まれたバイト数を検証する
	if *writtenByte != len(initializer) {
		echo("[Couldn't complete process to initialize script file.]\r\n")
		os.Exit(255)
	}
	// ファイルポインタオブジェクトから絶対パスを取得する
	*tentativeFile, err = filepath.Abs(ff.Name())
	if err != nil {
		echo(err.Error() + "\r\n")
		os.Exit(255)
	}

	var count int = 0
	//var ss int = 0
	var multiple int = 0
	//var backup []byte = make([]byte, 0)
	var currentDir string

	// saveコマンド入力用
	var saveFp *os.File
	saveFp = new(os.File)

	var fixedInput string
	input = initializer
	fixedInput = input
	var exitCode int

	for {
		if multiple == 1 {
			echo("(" + strconv.Itoa(exitCode) + ")" + " .... ")
		} else {
			echo("(" + strconv.Itoa(exitCode) + ")" + "php > ")
		}
		*line = ""

		// 標準入力開始
		stdin(line)
		temp := *line

		if temp == "del" {
			ff, err = deleteFile(ff, initializer)
			if err != nil {
				echo(err.Error() + "\r\n")
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
				echo(err.Error() + "\r\n")
				continue
			}
			saveFp.Chmod(os.ModePerm)
			input = fixedInput
			*writtenByte, err = saveFp.WriteAt([]byte(input), 0)
			if err != nil {
				saveFp.Close()
				echo(err.Error() + "\r\n")
				os.Exit(255)
			}
			echo("[" + currentDir + ":Completed saving input code which you wrote.]" + "\r\n")
			saveFp.Close()
			*line = ""
			multiple = 0
			exitCode = 0
			continue
		} else if temp == "exit" {
			// コンソールを終了させる
			echo("[Would you really like to quit a console which you are running in terminal? Pushing Enter key or other]\r\n")
			var quitText *string
			quitText = new(string)
			stdin(quitText)
			if *quitText == "" {
				ff.Close()
				os.Remove(*tentativeFile)
				os.Exit(0)
			} else {
				echo("[Canceled to quit this console app in terminal.]\r\n")
			}
			*line = ""
			continue
		} else if temp == "restore" || temp == "clear" {
			input = fixedInput
			os.Truncate(*tentativeFile, 0)
			ff.WriteAt([]byte(input), 0)
			multiple = 0
			exitCode = 0
			continue
		} else if temp == "" {
			// 空文字エンターの場合はループを飛ばす
			continue
		}

		input += *line + "\n"

		_, err = ff.WriteAt([]byte(input), 0)
		if err != nil {
			// temporary fileへの書き込みに失敗した場合
			echo(err.Error())
			continue
		}

		// *sync.WaitGroup 及び 共有メモリを使用したバージョン
		wg.Add(1)
		go SyntaxCheckUsingWaitGroup(tentativeFile, wg, &commonBool, &exitCode)
		wg.Wait()
		if commonBool == true {
			*line = ""
			fixedInput = input + "echo (PHP_EOL);"
			count, err = tempFunction(ff, tentativeFile, count, false)
			if err != nil {
				echo(err.Error())
				continue
			}
			multiple = 0
			input += " echo(PHP_EOL);\r\n "
		} else {
			//_, err = tempFunction(ff, tentativeFile, count, true)
			multiple = 1
		}

		// //channel を使った場合

		// // 並行処理でスクリプトが正常実行できるまでループを繰り返す
		// // wg.Add(1)
		// go SyntaxCheckUsingChannel(tentativeFile, syntax, cc chan int ,)
		// // チャンネルから値を取得
		// si := <-syntax
		// exitCode = <-cc
		// //		wg.Wait()
		// if si == 1 {
		// 	*line = ""
		// 	fixedInput = input + "echo (PHP_EOL);"
		// 	count, err = tempFunction(ff, tentativeFile, count, false)
		// 	if err != nil {
		// 		echo(err.Error())
		// 		continue
		// 	}
		// 	multiple = 0
		// 	input += " echo(PHP_EOL);\r\n "
		// } else {
		// 	_, err = tempFunction(ff, tentativeFile, count, true)
		// 	multiple = 1
		// }

	}
}

// SyntaxCheckUsingChannel
func SyntaxCheckUsingChannel(filePath *string, c chan int, cc chan int) (bool, error) {
	var e error = nil
	var command *exe.Cmd
	// バックグラウンドでPHPをコマンドラインで実行
	command = exe.Command("php", *filePath)
	e = command.Run()
	//wg.Done()
	// ガベージコレクション
	/*
		debug.SetGCPercent(100)
		runtime.GC()
		debug.FreeOSMemory()
	*/
	if e == nil {
		// コマンド成功時
		c <- 1
		cc <- command.ProcessState.ExitCode()
		return true, nil
	} else {
		// コマンド実行失敗時
		c <- 0
		cc <- command.ProcessState.ExitCode()
		return false, e
	}
}

/**
 *SyntaxCheckUsingWaitGroup WaitGroupオブジェクトを使ったバージョン
 * @param string filePath
 * @param *sync.WaitGroup w
 * @param *int exitedStatus
 *
 * @return bool, error
 */
func SyntaxCheckUsingWaitGroup(filePath *string, w *sync.WaitGroup, b *bool, exitedStatus *int) (bool, error) {
	var e error = nil
	var command *exe.Cmd
	var waitStatus syscall.WaitStatus
	var ok bool
	// バックグラウンドでPHPをコマンドラインで実行
	command = exe.Command("php", *filePath)
	command.Run()
	// command.ProcessState.Sys()は interface{}を返却する
	waitStatus, ok = command.ProcessState.Sys().(syscall.WaitStatus)
	// 型アサーション成功時
	if ok == true {
		*exitedStatus = waitStatus.ExitStatus()
		var ps *os.ProcessState
		ps = command.ProcessState
		if ps.Success() {
			// コマンド成功時
			w.Done()
			*b = true
			return true, nil
		} else {
			// コマンド実行失敗時
			w.Done()
			*b = false
			return false, e
		}
	} else {
		// コマンド実行失敗時
		w.Done()
		*b = false
		return false, e
	}
}

func tempFunction(fp *os.File, filePath *string, beforeOffset int, errorCheck bool) (int, error) {
	echo := Echo()
	var e error
	var stdout io.ReadCloser
	var command *exe.Cmd
	var ii int = 0
	var scanText string
	var code bool

	if errorCheck == true {
		command = exe.Command("php", *filePath)
		// バックグラウンドでPHPをコマンドラインで実行
		e = command.Run()
		// バックグランドでの実行が失敗の場合
		if e != nil {
			// 実行したスクリプトの終了コードを取得
			code = command.ProcessState.Success()
			if code != true {
				scanText = ""
				command = exe.Command("php", *filePath)
				stdout, _ := command.StdoutPipe()
				command.Start()
				scanner := bufio.NewScanner(stdout)
				ii = 0
				for scanner.Scan() {
					if ii >= beforeOffset {
						scanText = scanner.Text()
						if len(scanText) > 0 {
							echo("     " + scanner.Text() + "\r\n")
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
							echo("     " + scanner.Text() + "\r\n")
						}
					}
				}
				command.Wait()
				echo("\r\n")
				command = nil
				stdout = nil
				return beforeOffset, e
			}
		}
	}
	// Run()メソッドで利用したcommandオブジェクトを再利用
	command = exe.Command("php", *filePath)
	stdout, e = command.StdoutPipe()
	if e != nil {
		echo(e.Error() + "\r\n")
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
					echo("     " + scanText + "\r\n")
				}
			} else {
				scanText = scanner.Text()
			}
			ii++
		} else {
			break
		}
		scanText = ""
	}
	command.Wait()
	command = nil
	stdout = nil
	scanText = ""
	echo("\r\n")
	fp.Write([]byte("echo(PHP_EOL);\r\n"))

	/*
		debug.SetGCPercent(100)
		runtime.GC()
		debug.FreeOSMemory()
	*/
	return ii, e
}

func deleteFile(fp *os.File, initialString string) (*os.File, error) {
	var err error
	fp.Truncate(0)
	fp.Seek(0, 0)
	_, err = fp.WriteAt([]byte(initialString), 0)
	fp.Seek(0, 0)

	/*
		debug.SetGCPercent(100)
		runtime.GC()
		debug.FreeOSMemory()
	*/
	return fp, err
}
