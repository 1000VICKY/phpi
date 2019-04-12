// 表示入力を実行
package standardInput;

import "os";
import "strings";
import "phpa/echo";

// 表示入力を実行する関数オブジェクトのみを保持する
type StandardInput struct {
    input func (*string) bool;
    // バッファサイズを指定
    size int ;
}
// 標準入力関数をオブジェクトから取得
func (self *StandardInput) GetStandardInputFunction () (func(*string) bool ){
    return self.input;
}
// バッファサイズを任意に指定する
func (self *StandardInput) SetBufferSize (size int ) {
    self.size = size;
}
// オブジェクトに標準入力関数を設定
func (self *StandardInput) SetStandardInputFunction() {
    // 無名関数を変数へ保持
    self.input = func(s *string) bool {
        var size = 64;
        var writtenSize int = 0;
        var buffer []byte = make([]byte, size);
        var err interface{};
        var value error;
        var ok bool;
        for {
            // interface{}型のerr変数に意図的にエラーオブジェクトを保持
            writtenSize, err = os.Stdin.Read(buffer);
            // 型アサーションを実施
            value, ok = err.(error);
            // 型アサーションの検証結果
            if (ok == true && value != nil) {
                echo.Echo("[" + value.Error() + "]");
                return false;
            }
            *s += string(buffer[:writtenSize]);
            if (writtenSize < size) {
                break;
            }
        }
        *s = strings.Trim(*s, "\r\n");
        // 入力終了
        return true;
    };
};
