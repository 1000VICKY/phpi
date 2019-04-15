// PHPの組み込み関数をシミュレーション
package main

import "os"
import "fmt"
import "errors"

// 空のインターフェース
type myInterface interface {
	A()
}

// MyFunction
func MyFunction() func(...interface{}) (int, error) {
	var echo (func(...interface{}) (int, error))

	echo = fmt.Println
	return echo
}

func main() {
	echo := MyFunction()
	echo("文字列を出力")
	var fp *os.File = new(os.File)
	var err error
	var filePath string = "./senbiki.dat"
	var ok bool
	// ファイルを作成
	fp, err = Fopen(filePath, "w")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(255)
	}

	ok, err = FileExists(filePath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(255)
	}
	fmt.Println(ok)
	// ファイルポインタを閉じる
	Fclose(fp)
}

// FileExists 指定したファイルが存在するかどうかを検証
func FileExists(filePath string) (bool, error) {
	fmt.Println(filePath)
	var fileInfo os.FileInfo
	var err error
	// FileInfoオブジェクトを取得
	fileInfo, err = os.Stat(filePath)
	fmt.Println(fileInfo.Name())
	if err != nil {
		// 検証に失敗した場合
		return false, err
	} else {
		return true, err
	}
}

// Fopen PHPのfopenをシミュレーション
func Fopen(filePath string, mode string) (*os.File, error) {
	// 変数の宣言
	var permission os.FileMode
	var fp *os.File
	var err error
	var openFlag *int

	// 変数の初期化
	permission = 0755
	fp = new(os.File)
	err = nil
	openFlag = new(int)

	// 指定されたフラグによってファイルの処理を分ける
	if mode == "w" {
		// 新規作成型
		*openFlag = os.O_CREATE | os.O_WRONLY
		// ファイルサイズを0にする
		os.Truncate(filePath, 0)
	} else if mode == "w+" {
		// 新規作成
		// 読み出しおよび書き出しで開く
		*openFlag = os.O_CREATE | os.O_RDWR
		// ファイルサイズを0にする
		os.Truncate(filePath, 0)
	} else if mode == "a" {
		// 追記型
		// 書き込みのみ許可
		*openFlag = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	} else if mode == "a+" {
		// 追記型
		// 書き込み及び読み込み許可
		*openFlag = os.O_CREATE | os.O_APPEND | os.O_RDWR
	} else if mode == "r" {
		// 読み込みのみ型
		*openFlag = os.O_RDONLY
	} else if mode == "r+" {
		// 読み込み及び書き込み型で開く
		*openFlag = os.O_RDWR
	} else {
		return nil, errors.New("A open mode type isn't specified to open a file which you wrote on a source.\r\n")
	}
	// 指定した条件でOpenFileを実行
	fp, err = os.OpenFile(filePath, *openFlag, permission)
	return fp, err
}

// Fwrite シミュレーション
func Fwrite(filePointer *os.File, text string) (int, error) {
	var buffer []byte
	var writtenByte int
	var err *error = new(error)
	// 文字列をbyte列にキャスト
	buffer = []byte(text)
	writtenByte, *err = filePointer.Write(buffer)
	if *err != nil {
		// 書き込みに失敗した場合
		return 0, *err
	}
	return writtenByte, *err
}

// Fclose シミュレーション
func Fclose(f *os.File) (bool, error) {
	// nilでないことを確認
	if f == nil {
		return false, errors.New("[Value passed to function is not type *os.File.]")
	}
	var err *error = new(error)
	*err = f.Close()
	if err != nil {
		// closeに失敗
		return false, *err
	}
	// 正常にclose
	return true, *err
}
