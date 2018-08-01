package main

import (
	"bufio"
	"fmt"
	_ "io"
	"io/ioutil"
	"os"
	exe "os/exec"
	"path/filepath"
	_ "reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	_ "strconv"
	"strings"
	"unsafe"
)

var format func(...interface{}) (int, error) = fmt.Println
var myPrint func(...interface{}) (int, error) = fmt.Print
var split []string = make([]string, 0)

func main() {
	const initializer = "<?php " + "\n"
	// 利用変数初期化
	var input string
	var line *string
	line = new(string)
	var ff *os.File
	var err error
	var tentativeFile *string = new(string)
	var myError error = nil
	var writtenByte *int = new(int)
	// ダミー実行ポインタ
	ff, myError = ioutil.TempFile("", "__php__main__")
	if myError != nil {
		format(myError.Error())
		os.Exit(1)
	}
	ff.Chmod(os.ModePerm)
	*writtenByte, myError = ff.WriteAt([]byte(initializer), 0)
	if myError != nil {
		format(err.Error())
		os.Exit(1)
	}
	// ファイルポインタに書き込まれたバイト数を検証する
	if *writtenByte != len(initializer) {
		format("<スクリプトファイルの初期化に失敗しました.>")
		os.Exit(255)
	}
	// ファイルポインタオブジェクトから絶対パスを取得する
	*tentativeFile, myError = filepath.Abs(ff.Name())
	if myError != nil {
		format(myError)
		os.Exit(255)
	}
	defer ff.Close()
	defer os.Remove(*tentativeFile)

	var count int = 0
	var ss int = 0
	var multiple int = 0
	var multipleTemp int = 0
	var backup []byte = make([]byte, 0)
	var currentDir string

	// 末尾はバックスラッシュの場合，以降再びバックスラッシュで終わるまで
	// スクリプトを実行しない
	var reg *regexp.Regexp = nil
	var openBrace *regexp.Regexp = new(regexp.Regexp)
	var openCount int = 0
	var closeBrace *regexp.Regexp = new(regexp.Regexp)
	var closeCount int = 0
	reg, myError = regexp.Compile("((\\\\)+[ ]*|_+[ ]*)$")
	// 正規表現実行箇所エラーハンドリング
	if myError != nil {
		format(myError.Error())
		format("<正規表現:RunTime Error>")
		os.Exit(255)
	}
	openBrace, _ = regexp.Compile("^.*{$")
	closeBrace, _ = regexp.Compile("^[ ]*}[ ]*$")
	// [save]というキーワードを入力した場合の正規表現
	var saveRegex *regexp.Regexp = new(regexp.Regexp)
	saveRegex, myError = regexp.Compile("^[ ]*save[ ]*$")
	if myError != nil {
		format(myError.Error())
		format("<正規表現:Runtime Error>")
		os.Exit(255)
	}
	var scanner *bufio.Scanner = new(bufio.Scanner)
	scanner = bufio.NewScanner(os.Stdin)
	for {
		runtime.GC()
		// ループ開始時に正常動作するソースのバックアップを取得
		ff.Seek(0, 0)
		backup, myError = ioutil.ReadAll(ff)
		if myError != nil {
			format("バックアップに失敗!")
			format(myError.Error())
			break
		}

		ff.WriteAt(backup, 0)
		if multiple == 1 {
			myPrint("... ")
		} else {
			myPrint(">>> ")
		}
		scanner.Scan()
		*line = scanner.Text()

		if *line == "del" {
			ff, myError = deleteFile(ff, "<?php ")
			*line = ""
			input = ""
			count = 0
			continue
		} else if saveRegex.MatchString(*line) {
			/*
			   [save]コマンドが入力された場合，その時点まで入力されたスクリプトを
			   カレントディレクトリに保存する
			*/
			currentDir, myError = os.Getwd()
			if myError != nil {
				format("<カレントディレクトリの取得に失敗>")
				format(myError.Error())
				break
			}
			currentDir, myError = filepath.Abs(currentDir)
			if myError != nil {
				format(myError.Error())
				break
			}
			var saveFp *os.File = new(os.File)
			if runtime.GOOS == "windows" {
				currentDir += "\\save.php"
			} else {
				currentDir += "/save.php"
			}
			saveFp, myError = os.Create(currentDir)
			if myError != nil {
				format(myError.Error())
				continue
			}
			saveFp.Chmod(os.ModePerm)
			defer saveFp.Close()
			*writtenByte, myError = saveFp.WriteAt(backup, 0)
			if myError != nil {
				format(myError.Error())
				os.Exit(255)
			}
			format(currentDir + ":入力した内容を保存しました。")
			saveFp.Close()
			*line = ""
			input = ""
			continue
		} else if *line == "" {
			// 空文字エンターの場合はループを飛ばす
			continue
		}

		ob := openBrace.MatchString(*line)
		if ob == true {
			openCount = openCount + 1
		}
		cb := closeBrace.MatchString(*line)
		if cb == true {
			closeCount = closeCount + 1
		}
		// ブレースによる複数入力フラグがfalseの場合
		if openCount == 0 && closeCount == 0 {
			// 正規表現のマッチチェック
			res := reg.MatchString(*line)
			if res == true {
				// 複数行入力フラグ
				if multipleTemp == 0 {
					multipleTemp = 1
					multiple = 1
				} else if multipleTemp == 1 {
					multipleTemp = 0
					multiple = 0
				} else {
					format("<不正な処理:Runtime Error>")
					break
				}
			} else {
				// 複数行入力フラグ
				if multipleTemp == 0 {
					multipleTemp = 0
					multiple = 0
				} else if multipleTemp == 1 {
					multipleTemp = 1
					multiple = 1
				} else {
					format("<不正な処理:Runtime Error>")
					break
				}
			}
		} else if openCount != closeCount {
			multiple = 1
		} else if openCount == closeCount {
			multiple = 0
			openCount = 0
			closeCount = 0
		} else {

		}
		*line = string(reg.ReplaceAll([]byte(*line), []byte("")))
		input += *line + "\n"
		if multiple == 0 {
			ss, myError = ff.Write([]byte(input))
			if myError != nil {
				format("<ファイルポインタへの書き込み失敗>")
				format("=>" + myError.Error())
				continue
			}
			if ss > 0 {
				input = ""
				*line = ""
				count, myError = tempFunction(ff, tentativeFile, &count, backup)
				if err != nil {
					format(myError.Error())
					continue
				}
			}
		} else if multiple == 1 {
			// 現在複数行で入力中
		} else {
			format("<不正な処理>")
		}
	}
}

var output []byte = make([]byte, 0)

func tempFunction(temporaryFp *os.File, temporaryFilePath *string, beforeOffset *int, temporaryBackup []byte) (int, error) {
	var e error = new(MyError)
	var index *int = new(int)
	var start *int = new(int)
	var _int *int = new(int)
	_intPointer := unsafe.Pointer(_int)
	var _float *float64 = (*float64)(_intPointer)
	format(_float)
	format(*_float)
	*_float = 0.5
	format(*_float)
	format(*_int)
	runtime.GC()
	// バックグラウンドでPHPをコマンドラインで実行
	output, e = exe.Command("php", *temporaryFilePath).Output()
	// stdinから読み出したスクリプトが失敗した場合
	if e != nil {
		format(string(output))
		temporaryFp.Truncate(0)
		temporaryFp.Seek(0, 0)
		temporaryFp.WriteAt(temporaryBackup, 0)
		return *beforeOffset, e
	}
	output = output[*beforeOffset:]
	*index = len(output) + *beforeOffset
	var strOutput []string = strings.Split(string(output), "\n")
	maxLength := len(strOutput)
	for *start = 0; *start < maxLength; *start++ {
		fmt.Print("    ")
		format(strOutput[*start])
	}
	output = nil
	strOutput = nil
	start = nil
	temporaryFp.Write([]byte("echo(PHP_EOL);"))
	runtime.GC()
	// プログラムが確保したメモリを強制的にOSへ返却
	debug.FreeOSMemory()
	return *index, e
}

type MyError struct {
	ErrorMessage string
}

func (this *MyError) Error() string {
	return this.ErrorMessage
}

func deleteFile(fp *os.File, initialString string) (*os.File, error) {
	var size int
	var e error
	var myError *MyError = new(MyError)
	fp.Truncate(0)
	fp.Seek(0, 0)
	size, e = fp.WriteAt([]byte(initialString), 0)
	fp.Seek(0, 0)
	if e == nil && size >= 0 {
		myError = nil
		return fp, myError
	} else {
		myError.ErrorMessage = "一時ファイルの初期化に失敗しました。"
		return fp, myError
	}
}
