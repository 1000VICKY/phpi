// 簡易に標準出力に出力する関数を定義する
package echo

import (
	"errors"
	"os"
)

/**
 * Echo 関数を定義する
 * @param s string
 * @return int, error
 */
func Echo() func(interface{}) (int, error) {

	// クロージャを返却する
	return func(i interface{}) (int, error) {
		var s string
		var ok bool
		// 引数に渡されたinterface{}の型アサーション
		s, ok = i.(string)
		if ok == true {
			// 文字列型をバイト列化する
			var buffer []byte = []byte(s)
			var size *int = new(int)
			var err error = nil
			*size, err = os.Stdout.Write(buffer)
			// 上記戻り値をそのまま返却
			return *size, err
		} else {
			// 型アサーションに失敗時
			return 0, errors.New("[Failed to run type asserstion.]")
		}
	}
}
