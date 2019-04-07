package echo;

import "os";
// echo関数を定義
var Echo func (string) (int, error) = func (s string) (int, error) {
    // os.Stdout.Writeメソッドに渡す文字列を[]byteへ変換
    var buffer []byte = []byte(s);
    size, err := os.Stdout.Write(buffer);
    // Writeメソッドの戻り値をそのまま返却
    return size, err;
};