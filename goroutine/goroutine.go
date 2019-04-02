// 自作パッケージ
package myPackage;

import "os";
import "fmt";
import
"syscall";
import
"time";
import
"runtime";
import
"runtime/debug";

func MonitoringSignal(sig chan os.Signal, exit chan int) {
    var s os.Signal;
    for {
        s, _ = <-sig;
        if (s == syscall.SIGHUP) {
            fmt.Printf("[syscall.SIGHUP]\r\n")
            // 割り込みを無視
            exit <- 0;
        } else if (s == syscall.SIGTERM) {
            fmt.Println("[syscall.SIGTERM].\r\n")
            exit <- 1;
        } else if (s == os.Kill) {
            fmt.Printf("[os.Kill].\r\n")
            // 割り込みを無視
            exit <- 0;
        } else if (s == os.Interrupt) {
            fmt.Printf("[os.Interrupt].\r\n")
            // 割り込みを無視
            exit <- 0;
        } else if (s == syscall.Signal(0x14)) {
            fmt.Printf("[syscall.SIGTSTP].\r\n")
            // 割り込みを無視
            exit <- 0;
        } else if (s == syscall.SIGQUIT) {
            fmt.Println("[syscall.SIGQUIT].\r\n");
            exit <- 1;
        }
    }
}

func CrushingSignal(exit chan int ) {
    var code int = 0;
    for {
        code, _ = <-exit;
        if (code == 1 ) {
            os.Exit(code);
        } else {
            fmt.Println("<割り込みを無視しました。>");
        }
    }
};

func RunningFreeOSMemory() {
    // 定期時間ごとにガベージコレクションを動作させる
    for
    {
        time.Sleep(10 * time.Second);
        runtime.GC();
        debug.FreeOSMemory();
    }
};

type My struct {
    name string;
    age int;
};
// 名前をメンバに代入
func (this *My) SetName(name string) *My {
    this.name = name;
    return (this);
}
// nameメンバを取得
func (this *My) GetName() string {
    return this.name;
}
func (this *My) SetAge (age int ) *My {
    this.age = age;
    return (this);
}
func (this *My) GetAge () int {
    return (this.age);
}