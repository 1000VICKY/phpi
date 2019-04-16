// +build windows

package main

import (
	"bufio"
	_ "errors"
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
	. "regexp"
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
	/*
		// プロセスの監視
		var signal_chan chan os.Signal = make(chan os.Signal)
		// OSによってシグナルのパッケージを変更
		signal.Notify(
			signal_chan,
			os.Interrupt,
			os.Kill,
			unix.SIGKILL,
			unix.SIGHUP,
			unix.SIGINT,
			unix.SIGTERM,
			unix.SIGQUIT,
			unix.SIGTSTP,
			unix.Signal(0x13),
			unix.Signal(0x14), // Windowsの場合 SIGTSTPを認識しないためリテラルで指定する
		)
	*/

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
	var line *string = new(string)
	line = new(string)
	var tentativeFile *string
	tentativeFile = new(string)
	var ff *os.File
	var err error
	var writtenByte *int
	writtenByte = new(int)

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
	var ss int = 0
	var multiple int = 0
	var backup []byte = make([]byte, 0)
	var currentDir string

	// 開き括弧
	var openBrace *Regexp
	openBrace = new(Regexp)
	var openCount int
	openCount = 0
	// 閉じ括弧
	var closeBrace *Regexp
	closeBrace = new(Regexp)
	var closeCount int
	closeCount = 0
	openBrace, _ = Compile("^.*{[ \t]*.*$")
	closeBrace, _ = Compile("^.*}.*$")
	// ヒアドキュメントを入力された場合
	var startHereDocument *Regexp
	startHereDocument = new(Regexp)
	startHereDocument, err = Compile("^.*<<<[ ]*([_a-zA-Z0-9]+)$")
	if err != nil {
		echo.Echo(err.Error() + "\r\n")
		os.Exit(255)
	}
	// ヒアドキュメントで入力された場合
	var hereFlag bool = false
	// マッチしたヒアドキュメントタグを取得するため
	var hereTag [][]string = make([][]string, 1)
	var ID string = ""
	var endHereDocument *Regexp = new(Regexp)
	var saveFp *os.File = new(os.File)
	//var syntaxChan chan int = make(chan int)
	for {
		runtime.GC()
		debug.SetGCPercent(100)
		debug.FreeOSMemory()
		// ループ開始時に正常動作するソースのバックアップを取得
		ff.Seek(0, 0)
		backup, err = ioutil.ReadAll(ff)
		if err != nil {
			echo.Echo(err.Error() + "\r\n")
			break
		}

		ff.WriteAt(backup, 0)
		if multiple == 1 {
			echo.Echo(" .... ")
		} else {
			echo.Echo("php > ")
		}
		*line = ""

		// 標準入力開始
		stdin(line)

		// ヒアドキュメントで入力された場合
		if hereFlag == false {
			hereTag = startHereDocument.FindAllStringSubmatch(*line, -1)
			if len(hereTag) > 0 {
				if len(hereTag[0]) > 0 {
					ID = hereTag[0][1]
					hereFlag = true
					echo.Echo("(" + ID + ")" + "\r\n")
				}
			} else {
				hereFlag = false
			}
		} else {
			endHereDocument, err = Compile("^" + ID + "[ ]*;$")
			if endHereDocument.MatchString(*line) {
				hereFlag = false
			} else {
				hereFlag = true
			}
		}

		if *line == "del" {
			ff, err = deleteFile(ff, initializer)
			if err != nil {
				echo.Echo(err.Error() + "\r\n")
				os.Exit(255)
			}
			*line = ""
			input = ""
			count = 0
			continue
		} else if *line == "save" {
			// saveキーワードが入力された場合
			// OSによってパスの差し替え(build タグによって差し替えるためif文は削除)
			currentDir, err = os.Getwd()
			currentDir += "\\save.php"
			saveFp, err = os.Create(currentDir)
			if err != nil {
				echo.Echo(err.Error() + "\r\n")
				continue
			}
			saveFp.Chmod(os.ModePerm)
			*writtenByte, err = saveFp.WriteAt(backup, 0)
			if err != nil {
				saveFp.Close()
				echo.Echo(err.Error() + "\r\n")
				os.Exit(255)
			}
			echo.Echo("[" + currentDir + ":Completed saving input code which you wrote.]" + "\r\n")
			saveFp.Close()
			*line = ""
			input = ""
			continue
		} else if *line == "exit" {
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
			input = ""
			continue
		} else if *line == "" {
			// 空文字エンターの場合はループを飛ばす
			continue
		}

		ob := openBrace.FindAllStringSubmatch(*line, -1)
		if len(ob) > 0 {
			if len(ob[0]) > 0 {
				openCount = openCount + len(ob[0])
			}
		}

		cb := closeBrace.FindAllStringSubmatch(*line, -1)
		if len(cb) > 0 {
			if len(cb[0]) > 0 {
				closeCount = closeCount + len(cb[0])
			}
		}
		// ブレースによる複数入力フラグがfalseの場合
		if openCount == 0 && closeCount == 0 {
			multiple = 0
			if hereFlag == true {
				multiple = 1
			} else if hereFlag == false {
				multiple = 0
			}
		} else if openCount != closeCount {
			multiple = 1
		} else if openCount == closeCount {
			multiple = 0
			openCount = 0
			closeCount = 0
		} else {
			panic("[Runtime Error happened!]")
		}
		input += *line + "\n"
		if multiple == 0 {
			ss, err = ff.Write([]byte(input))
			// 並行処理でスクリプトが正常実行できるまでループを繰り返す
			// go SyntaxCheck(tentativeFile, syntaxChan)
			if err != nil {
				echo.Echo("[Failed to write input code to file pointer.]" + "\r\n")
				echo.Echo("    " + err.Error() + "\r\n")
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
			panic("[Runtime Error which system could not understand happeds.]")
		}
	}
}

// シンタックスチェックのみを実行する
// SyntaxCheck バックグランドでコマンドが正常終了したかどうかを検証する
func SyntaxCheck(filePath *string, c chan int) (bool, error) {
	defer debug.SetGCPercent(100)
	defer runtime.GC()
	defer debug.FreeOSMemory()
	var e error = nil
	var command *exe.Cmd = new(exe.Cmd)
	// バックグラウンドでPHPをコマンドラインで実行
	// php -l コマンドで構文検証を行う
	command = exe.Command("php", "-l", *filePath)
	e = command.Run()
	if e != nil {
		c <- 0
		return false, e
	} else {
		c <- 1
		return true, nil
	}
}

func tempFunction(fp *os.File, filePath *string, beforeOffset int, temporaryBackup []byte) (int, error) {
	defer debug.SetGCPercent(100)
	defer runtime.GC()
	defer debug.FreeOSMemory()
	var e error = nil
	// バックグラウンドでPHPをコマンドラインで実行
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
