package main

import "fmt"
import "os"
import exe "os/exec"
import "io/ioutil"
import _ "io"
import _ "reflect"
import "strconv"
import "bufio"
import "path/filepath"
import "strings"
import "regexp"

var format func(...interface{}) (int, error) = fmt.Println
var myPrint func(...interface{}) (int, error) = fmt.Print

func main() {
	const initializer = "<?php " + "\n"
	// 利用変数初期化
	var input string
	var line string
	var ff *os.File
	var err error
	var tentativeFile string = ""

	// (1) 一時ファイル作成，且つ前回起動時のゴミファイルを削除する
	var initialize *os.File = nil
	var initializeFileName string = ""
	var absolutePath string = ""
	var myError error = nil
	// ※戻り値 => *os.File, error を返却
	initialize, myError = ioutil.TempFile("", "__php__main__")
	defer initialize.Close()
	if myError != nil {
		format(myError.Error())
		os.Exit(255)
	}
	initialize.Chmod(os.ModePerm)
	// ファイルポインタから当該のファイルパスを取得(rootからの絶対パスを取得)
	initializeFileName = initialize.Name()
	// 念の為Absメソッドで絶対パスを取得
	initializeFileName, myError = filepath.Abs(initializeFileName)
	if myError != nil {
		format(myError.Error())
		os.Exit(255)
	}
	// 取得した絶対パスからディレクトリ名のみを取得
	absolutePath = filepath.Dir(initializeFileName)
	// Globで該当ファイルをすべて取得
	var fileList []string = make([]string, 0)
	fileList, myError = filepath.Glob(absolutePath + "/" + "__php__main__*")
	if myError != nil {
		format(myError.Error())
		os.Exit(255)
	}
	for key, value := range fileList {
		myError = os.Remove(value)
		if myError != nil {
			k := strconv.Itoa(key)
			format(myError.Error())
			format("インデックスキー => [" + k + "]" + "ファイル名=> [" + value + "] の削除に失敗しました。")
		}
	}

	// ダミー実行ポインタ
	ff, myError = ioutil.TempFile("", "__php__main__")
	if myError != nil {
		format(myError.Error())
		os.Exit(1)
	}
	ff.Chmod(os.ModePerm)
	_, myError = ff.WriteAt([]byte(initializer), 0)
	if myError != nil {
		format(err.Error())
		os.Exit(1)
	}
	tentativeFile = ff.Name()
	defer ff.Close()
	defer os.Remove(tentativeFile)

	var count int = 0
	var ss int = 0
	var multiple int = 0
	var multipleTemp int = 0
	var backup []byte = make([]byte, 0)
	var currentDir string

	// 末尾はバックスラッシュの場合，以降再びバックスラッシュで終わるまで
	// スクリプトを実行しない
	var reg *regexp.Regexp = nil
	reg, myError = regexp.Compile("\\\\[ ]*$")
	// 正規表現実行箇所エラーハンドリング
	if myError != nil {
		format(myError.Error())
		format("<正規表現:RunTime Error>")
		os.Exit(255)
	}
	// [save]というキーワードを入力した場合の正規表現
	var saveRegex *regexp.Regexp = new(regexp.Regexp)
	saveRegex, myError = regexp.Compile("[ ]*save[ ]*$")
	if myError != nil {
		format(myError.Error())
		format("<正規表現:Runtime Error>")
		os.Exit(255)
	}
	var scanner *bufio.Scanner = new(bufio.Scanner)
	scanner = bufio.NewScanner(os.Stdin)
	for {
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
		line = scanner.Text()

		if line == "del" {
			ff, myError = deleteFile(ff, "<?php ")
			line = ""
			input = ""
			count = 0
			continue
		} else if saveRegex.MatchString(line) {
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
			saveFp, myError = ioutil.TempFile(currentDir, "save")
			if myError != nil {
				format(myError.Error())
				continue
			}
			saveFp.Chmod(os.ModePerm)
			defer saveFp.Close()
			saveFp.WriteAt(backup, 0)
			saveFp.Close()
			line = ""
			input = ""
			continue
		} else if line == "" {
			// 空文字エンターの場合はループを飛ばす
			continue
		}

		// 正規表現のマッチチェック
		res := reg.MatchString(line)
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
		line = string(reg.ReplaceAll([]byte(line), []byte("")))
		input += line + "\n"
		if multiple == 0 {
			ss, myError = ff.Write([]byte(input))
			if myError != nil {
				format("<ファイルポインタへの書き込み失敗>")
				format("=>" + myError.Error())
				continue
			}
			if ss > 0 {
				input = ""
				line = ""
				count, myError = tempFunction(ff, tentativeFile, count, backup)
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
func tempFunction(temporaryFp *os.File, temporaryFilePath string, beforeOffset int, temporaryBackup []byte) (int, error) {
	var output []byte = make([]byte, 0)
	var split []string = make([]string, 0)
	var k int = 0
	var e error = new(MyError)

	// バックグラウンドでPHPをコマンドラインで実行
	output, e = exe.Command("php", temporaryFilePath).Output()
	// stdinから読み出したスクリプトが失敗した場合
	if e != nil {
		format(string(output))
		temporaryFp.Truncate(0)
		temporaryFp.Seek(0, 0)
		temporaryFp.WriteAt(temporaryBackup, 0)
		return beforeOffset, e
	}
	split = strings.Split(string(output), "\n")
	for k = beforeOffset; k < len(split); k++ {
		myPrint("    ")
		format(split[k])
		split[k] = ""
	}
	temporaryFp.Write([]byte("echo(PHP_EOL);"))
	beforeOffset = len(split)
	return beforeOffset, e
}

// 関数内で自作エラーオブジェクトを生成
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
