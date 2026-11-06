package app

import (
	"fmt"
	"os"
	"strings"
)

// askConfirm 在危险操作前问一次 y/n。confirm_tips=false 时直接放行，
// 让脚本或"我知道自己在干什么"的用户一路跑下去。head 是要先打印的说明，
// 由调用方格式化好再传进来（避免把变量当格式串用）。
//
// 返回 false 表示用户放弃：调用方打印 M.ConfirmCancelled 后正常退出
// ——取消不是错误，退出码保持 0。
func askConfirm(set *Settings, head string) bool {
	if !set.confirmTips() {
		return true
	}
	if head != "" {
		fmt.Print(head)
		if !strings.HasSuffix(head, "\n") {
			fmt.Println()
		}
	}
	fmt.Print(M.ConfirmPrompt)
	for {
		line := strings.ToLower(strings.TrimSpace(readLine()))
		switch line {
		case "y", "yes", "是":
			return true
		case "n", "no", "否", "":
			// 空输入（含 stdin 已结束）按「否」处理：宁可什么都不做，
			// 也不能因为读不到输入就默认同意。
			return false
		}
		// 其他字符不当作任何答案，重新问。裁决要求：y/N 大小写不敏感，
		// 别的字母直接屏蔽，不能当成默认值提交。
		fmt.Print(M.ConfirmAgain)
	}
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
