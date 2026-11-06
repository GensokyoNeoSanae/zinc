package app

import "fmt"

// cmdReadme 打印内置说明：zinc 做什么、三种操作怎么记、怎么确认备份没坏。
// 这是 `zinc readme`，与仓库里那份 README 无关——那份是给开发者看的，
// 这份是装完就能看到的。
func cmdReadme() {
	fmt.Println(M.ReadmeTitle)
	fmt.Println()
	fmt.Print(M.ReadmeText)
}
