package app

import (
	"fmt"
	"os"
	"strings"
)

// askConfirm 在危险操作前问一次 y/n。confirm_tips=false 时直接放行，
// 让脚本或"我知道自己在干什么"的用户一路跑下去。
//
// 返回 false 表示用户放弃：调用方打印 M.ConfirmCancelled 后正常退出
// ——取消不是错误，退出码保持 0。
func askConfirm(set *Settings, head string, args ...any) bool {
	if !set.confirmTips() {
		return true
	}
	if head != "" {
		fmt.Printf(head, args...)
		if !strings.HasSuffix(head, "\n") {
			fmt.Println()
		}
	}
	fmt.Print(M.ConfirmPrompt)
	line := strings.ToLower(strings.TrimSpace(readLine()))
	return line == "y" || line == "yes" || line == "是"
}

// dirHasEntries 判断目录里是否已经有东西（含子目录）。用于决定恢复前
// 要不要停下来问一句：空目录直接恢复，非空目录才提示可能覆盖。
func dirHasEntries(dir string) bool {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(ents) > 0
}
