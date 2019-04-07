// 簡易に標準出力に出力する関数を定義する
package echo;

import "os";

/**
 * Echo 関数を定義する
 * @param s string
 * @return int, error
 */
func Echo (s string ) (int, error) {
    // 文字列をバイト列化
    var buffer []byte = []byte(s);
    var size *int = new(int);
    var err error = nil;
    *size, err = os.Stdout.Write(buffer);
    // 上記戻り値をそのまま返却
    return *size, err;
}