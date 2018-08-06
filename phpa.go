package main

import (
	"bufio"
	"errors"
	"fmt"
	_ "io"
	"io/ioutil"
	"os"
	exe "os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
)

var format func(...interface{}) (int, error) = fmt.Println
var myPrint func(...interface{}) (int, error) = fmt.Print

func main() {
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
	openBrace, _ = regexp.Compile("^.*{[ \t]*$")
	closeBrace, _ = regexp.Compile("^[ \t]*}[ \t;]*$")
	// [save]というキーワードを入力した場合の正規表現
	var saveRegex *regexp.Regexp = new(regexp.Regexp)
	saveRegex, err = regexp.Compile("^[ ]*save[ ]*$")
	if err != nil {
		format(err)
		os.Exit(255)
	}
	var scanner *bufio.Scanner = new(bufio.Scanner)
	scanner = bufio.NewScanner(os.Stdin)
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
			myPrint("... ")
		} else {
			myPrint(">>> ")
		}
		scanner.Scan()
		*line = scanner.Text()

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
			var saveFp *os.File = new(os.File)
			if runtime.GOOS == "windows" {
				currentDir += "\\save.php" // C:\\aaa\\bbb\\save.php
			} else {
				currentDir += "/save.php" // /aaa/bbb/save.php
			}
			saveFp, err = os.Create(currentDir)
			if err != nil {
				format(err)
				continue
			}
			saveFp.Chmod(os.ModePerm)
			defer saveFp.Close()
			*writtenByte, err = saveFp.WriteAt(backup, 0)
			if err != nil {
				format(err)
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
				count, err = tempFunction(ff, tentativeFile, &count, backup)
				if err != nil {
					continue
				}
			}
		} else if multiple == 1 {
			continue
		} else {
			format("<Runtime Error>")
		}
	}
}

func tempFunction(fp *os.File, filePath *string, beforeOffset *int, temporaryBackup []byte) (int, error) {
	defer debug.FreeOSMemory()
	var output []byte = make([]byte, 0)
	var e error = nil
	var index *int = new(int)
	runtime.GC()
	// バックグラウンドでPHPをコマンドラインで実行
	// (1)まずは終了コードを取得
	e = exe.Command("php", *filePath).Run()
	if e != nil {
		// スクリプトを実行した結果、実行失敗の場合
		output, e = exe.Command("php", *filePath).Output()
		fmt.Println("    " + e.Error())
		fp.Truncate(0)
		fp.Seek(0, 0)
		fp.WriteAt(temporaryBackup, 0)
		debug.FreeOSMemory()
		return *beforeOffset, e
	}
	output, _ = exe.Command("php", *filePath).Output()
	output = output[*beforeOffset:]
	*index = len(output) + *beforeOffset
	var strOutput []string = strings.Split(string(output), "\n")
	for _, value := range strOutput {
		fmt.Print("    ")
		format(value)
	}
	output = nil
	strOutput = nil
	fp.Write([]byte("echo(PHP_EOL);"))
	// プログラムが確保したメモリを強制的にOSへ返却
	debug.FreeOSMemory()
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
