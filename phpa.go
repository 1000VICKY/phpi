package main


import "fmt";
import "os";
import exe "os/exec";
import "io/ioutil";
import _ "io";
import _ "reflect";
import _ "strconv";
import "bufio";
import "path/filepath";
import "strings";
import "regexp";

func main () {
    // 利用変数初期化
    var input string;
    var line string;
    var ff *os.File;
    var err error;
    var tentativeFile string = "";
    // 文字列をバイト配列に
    var initializer string = "<?php ";


    // (1) 一時ファイル作成，且つ前回起動時のゴミファイルを削除する
    var format func (...interface{}) (int , error)  = fmt.Println;
    var myPrint func (...interface{}) (int , error) = fmt.Print;
    var initialize *os.File = nil;
    var initializeFileName string = "";
    var absolutePath string = "";
    var myError error = nil;
    // ※戻り値 => *os.File, error を返却
    initialize, myError = ioutil .TempFile("", "__php__main__");
    defer initialize.Close();
    if (myError != nil) {
        format(myError.Error());
        os.Exit(255);
    }
    initialize.Chmod(os.ModePerm);
    // ファイルポインタから当該のファイルパスを取得(rootからの絶対パスを取得)
    initializeFileName = initialize.Name();
    // 念の為Absメソッドで絶対パスを取得
    initializeFileName , _ = filepath.Abs(initializeFileName);
    // 取得した絶対パスからディレクトリ名のみを取得
    absolutePath = filepath.Dir(initializeFileName);
    // Globで該当ファイルをすべて取得
    var fileList []string = []string{};
    fileList, myError = filepath.Glob(absolutePath + "/" + "__php__main__*");
    if (myError != nil) {
        format(myError.Error());
        os.Exit(255);
    }
    for _, value := range fileList {
        err = os.Remove (value);
        if (err != nil) {
            format(err.Error());
            format("ファイル名=> [" + value + "] の削除に失敗しました。");
        }
    }

    // ダミー実行ポインタ
    ff , err = ioutil.TempFile("", "__php__main__");
    ff.Chmod(os.ModePerm);
    if (err != nil) {
        fmt.Println(err.Error());
        os.Exit(1);
    }
    _, err = ff.WriteAt([]byte(initializer), 0);
    if (err != nil) {
        format(err.Error());
        os.Exit(1);
    }
    tentativeFile = ff.Name();
    defer ff.Close();
    defer os.Remove(tentativeFile);


    var split []string;
    var count int = 0;
    var ss int = 0;
    var multiple int = 0;
    var multipleTemp int = 0;
    var backup []byte;
    var e error = nil;
    var reg *regexp.Regexp = nil;

    scanner := bufio.NewScanner(os.Stdin);
    for {

        // ループ開始時に正常動作するソースのバックアップを取得
        ff.Seek(0, 0);
        backup , e = ioutil.ReadAll(ff);
        if (e != nil) {
            format("バックアップに失敗!");
            format(e.Error());
            break;
        }

        ff.WriteAt(backup, 0);
        if (multiple == 1) {
            myPrint("... ");
        } else {
            myPrint(">>> ");
        }
        scanner.Scan();
        line = scanner.Text();

        if (line == "del") {
            ff, e = deleteFile(ff, "<?php ");
            line = "";
            input = "";
            count = 0;
            continue;
        } else if (line == "save") {
            // saveコマンドが入力された場合,その時点まで入力された
            // カレントディレクトリに保存する
            currentDir, e := os.Getwd();
            if (e != nil) {
                format("<カレントディレクトリの取得に失敗>");
                format(e.Error());
                break;
            }
            currentDir, e = filepath.Abs(currentDir);
            if (e != nil) {
                format(e.Error());
                break;
            }
            var saveFp *os.File = new (os.File);
            saveFp , e = ioutil.TempFile(currentDir, "save");
            if (e != nil) {
                format(e.Error());
                continue;
            }
            saveFp.Chmod(os.ModePerm);
            defer saveFp.Close();
            saveFp.WriteAt(backup, 0);
            saveFp.Close();
            line = "";
            input = "";
            continue;
        } else if (line == "") {
            // 空文字エンターの場合はループを飛ばす
            continue;
        }
        reg, e = regexp.Compile("\\\\[ ]*$");
        // 正規表現実行箇所エラーハンドリング
        if (e != nil) {
            format(e.Error());
            format("<正規表現:RunTime Error>");
            continue;
        }
        // 正規表現のマッチチェック
        res := reg.MatchString(line);
        if (res == true) {
            // 複数行入力フラグ
            if (multipleTemp == 0) {
                multipleTemp = 1;
                multiple = 1;
            } else if (multipleTemp == 1) {
                multipleTemp = 0;
                multiple = 0;
            } else {
                format("<不正な処理:Runtime Error>");
                break;
            }
        } else {
            // 複数行入力フラグ
            if (multipleTemp == 0) {
                multipleTemp = 0;
                multiple = 0;
            } else if (multipleTemp == 1) {
                multipleTemp = 1;
                multiple = 1;
            } else {
                format("<不正な処理:Runtime Error>");
                break;
            }
        }
        line = string(reg.ReplaceAll([]byte(line), []byte("")));
        input += line + "\n";
        if (multiple == 0) {
            ss, err = ff.Write([]byte(input));
            if (err != nil) {
                fmt.Println("<ファイルポインタへの書き込み失敗>");
                fmt.Println("=>" + err.Error());
                continue;
            }
            if (ss > 0) {
                // バックグラウンドでPHPをコマンドラインで実行
                output, e := exe . Command("php", tentativeFile).Output();
                // stdinから読み出したスクリプトが失敗した場合
                if ( e != nil) {
                    fmt.Println(string(output));
                    fmt.Println(e.Error());
                    ff.Truncate(0);
                    ff.Close();
                    ff, _ =  os.OpenFile(tentativeFile, os.O_RDWR|os.O_CREATE, os.ModePerm);
                    defer ff.Close();
                    ff.Chmod(os.ModePerm);
                    ff.WriteAt(backup, 0);
                    input = "";
                    line = "";
                    continue;
                }
                split = strings.Split(string(output), "\n");
                for k := count ; k < len(split); k++ {
                    fmt.Print(" ");
                    fmt.Println(split[k]);
                }
                ff.Write([]byte("echo(PHP_EOL);"));
                count = len(split);
            }
            // 入力文字数を初期化
            line = "";
            input = "";
        } else if (multiple == 1) {
            // 現在複数行で入力中
        } else {
            format("<不正な処理>");
        }
    }
}

    // 関数内で自作エラーオブジェクトを生成
    type MyError struct {
        ErrorMessage string;
    }
    func (this *MyError) Error() string {
        return this.ErrorMessage;
    }

func
    deleteFile ( fp *os.File, initialString string) (*os.File , error) {
    var size int ;
    var e error;
    var myError *MyError = new(MyError);

    fp.Truncate(0);
    fp.Seek(0, 0);
    size, e = fp.WriteAt([]byte(initialString), 0);
    fp.Seek(0, 0);
    if (e == nil && size >= 0) {
        myError = nil;
        return fp, myError;
    } else {
        myError.ErrorMessage = "一時ファイルの初期化に失敗しました。";
        return fp, myError;
    }
}
