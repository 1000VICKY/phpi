// 自作パッケージ
package goroutine

import (
	"fmt"
	"os"
	"phpa/echo"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"
)

func MonitoringSignal(sig chan os.Signal, exit chan int) {
	var s os.Signal
	for {
		s, _ = <-sig
		if s == syscall.SIGHUP {
			echo.Echo("[syscall.SIGHUP]\r\n")
			// 割り込みを無視
			exit <- 0
		} else if s == syscall.SIGTERM {
			echo.Echo("[syscall.SIGTERM].\r\n")
			exit <- 1
		} else if s == os.Kill {
			echo.Echo("[os.Kill].\r\n")
			// 割り込みを無視
			exit <- 0
		} else if s == os.Interrupt {
			if runtime.GOOS != "darwin" {
				echo.Echo("[os.Interrupt].\r\n")
			}
			// 割り込みを無視
			exit <- 0
		} else if s == syscall.Signal(0x14) {
			if runtime.GOOS != "darwin" {
				echo.Echo("[syscall.SIGTSTP].\r\n")
			}
			// 割り込みを無視
			exit <- 0
		} else if s == syscall.SIGQUIT {
			echo.Echo("[syscall.SIGQUIT].\r\n")
			exit <- 1
		}
	}
}

func CrushingSignal(exit chan int) {
	var code int = 0
	for {
		code, _ = <-exit
		if code == 1 {
			os.Exit(code)
		} else {
			if runtime.GOOS != "darwin" {
				echo.Echo("[Ignored interrupt]")
			}
		}
	}
}

func RunningFreeOSMemory() {
	var mem runtime.MemStats
	// 定期時間ごとにガベージコレクションを動作させる
	for {
		runtime.ReadMemStats(&mem)
		fmt.Println(mem.Alloc, mem.TotalAlloc, mem.HeapAlloc, mem.HeapSys)
		time.Sleep(5 * time.Second)
		runtime.GC()
		debug.FreeOSMemory()
	}
}

type My struct {
	name string
	age  int
}

// 名前をメンバに代入
func (this *My) SetName(name string) *My {
	this.name = name
	return (this)
}

// nameメンバを取得
func (this *My) GetName() string {
	return this.name
}
func (this *My) SetAge(age int) *My {
	this.age = age
	return (this)
}
func (this *My) GetAge() int {
	return (this.age)
}
