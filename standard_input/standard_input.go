// 表示入力を実行
package standardInput;

import "os";
import "strings";
import "phpa_with_go/echo";

// 表示入力を実行する関数オブジェクトのみを保持する
type StandardInput struct {
    input func (*string)
}
// 標準入力関数をオブジェクトから取得
func (self *StandardInput) GetStandardInputFunction () (func(*string)){
    return self.input;
}
// オブジェクトに標準入力関数を設定
func (self *StandardInput) SetStandardInputFunction() {
    self.input = func(s *string) {
        var size int = 64;
        var writtenSize int = 0;
        var buffer []byte = make([]byte, size);
        var err interface{};
        var value error;
        var ok bool;
        for {
            writtenSize, err = os.Stdin.Read(buffer);
            value, ok = err.(error);
            // 型アサーションの検証結果
            if (ok == true && value != nil) {
                echo.Echo("[" + value.Error() + "]");
                break;
            }
            *s += string(buffer[:(writtenSize)]);
            if (writtenSize < size) {
                break;
            }
        }
        *s = strings.Trim(*s, "\r\n");
        // 入力終了
    };
};