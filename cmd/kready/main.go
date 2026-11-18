// Command kready 检查 Kubernetes 集群交付就绪状态。
//
// 当前为脚手架阶段，尚无检查逻辑。
package main

import (
	"fmt"
	"os"
)

// version 在发布时通过 -ldflags 注入。
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("kready", version)
		return
	}
	fmt.Fprintln(os.Stderr, "kready: 尚未实现检查功能（脚手架阶段）")
	os.Exit(2)
}
