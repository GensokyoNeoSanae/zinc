package app

import (
	"fmt"
	"runtime"
)

// Version 是构建时注入的版本号，例如 v0.7.0。
// build.ps1 通过 -ldflags "-X .../internal/app.Version=v0.7.0" 写入；
// 直接 go run / go test 构建时保持 dev。
var Version = "dev"

// versionLine 是 `zinc --version` 的唯一输出行：单行、无装饰，
// 方便脚本按空白切分取第二个字段。
func versionLine() string {
	return fmt.Sprintf("zinc %s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH)
}

// printVersion 打印版本。调用方负责退出码：查版本不是错误。
func printVersion() {
	fmt.Println(versionLine())
}
