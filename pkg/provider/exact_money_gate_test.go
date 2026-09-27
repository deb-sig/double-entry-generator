package provider

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// ExactMoney 迁移门禁（2026-09-27）
//
// 引擎只认 `order.ExactMoney` 为权威金额，`order.Money`(float64) 是遗留视图。
// 但**当前没有一个 provider 写 ExactMoney**，全在 `float64(x)/100.0` 上丢精度。
// 这个测试把「谁迁了、谁还没迁」变成显式清单：
//
//   - migrated：必须设置 ExactMoney（改了要同时改这里）
//   - notYet：还没迁（存量债，允许但不是免费 —— 加新 provider 必须二选一登记）
//
// 任何未登记的新 provider 目录都会让测试失败，避免又悄悄多一个丢精度的实现。
var exactMoneyMigrated = []string{
	"jd", // 源即整数分，ExactFromCents 精确构造
}

var exactMoneyNotYet = []string{
	"abc_debit", "alipay", "bmo", "boc", "bocom_credit", "bocom_debit", "ccb",
	"cgb_credit", "citic", "cmb", "hsbchk", "htsec", "huobi", "hxsec", "icbc",
	"mt", "oklink", "spdb_debit", "td", "wechat",
}

func providerDirs(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read pkg/provider: %v", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs)
	return dirs
}

func setsExactMoney(t *testing.T, dir string) bool {
	t.Helper()
	found := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(string(data), "ExactMoney") {
			found = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return found
}

func TestExactMoneyCoverageIsTracked(t *testing.T) {
	migrated := map[string]bool{}
	for _, d := range exactMoneyMigrated {
		migrated[d] = true
	}
	notYet := map[string]bool{}
	for _, d := range exactMoneyNotYet {
		notYet[d] = true
	}

	var unlisted []string
	for _, dir := range providerDirs(t) {
		if !migrated[dir] && !notYet[dir] {
			unlisted = append(unlisted, dir)
		}
	}
	if len(unlisted) > 0 {
		t.Fatalf(
			"新 provider 未登记 ExactMoney 状态（必须选一边，别默认丢精度）：%v\n"+
				"  迁完 → 加进 exactMoneyMigrated；暂缓 → 加进 exactMoneyNotYet",
			unlisted,
		)
	}

	// 清单必须与代码一致：登记为已迁移的必须真的设了 ExactMoney。
	for _, dir := range exactMoneyMigrated {
		if !setsExactMoney(t, dir) {
			t.Errorf("%s 登记为已迁移，但没有设置 ExactMoney", dir)
		}
	}
	// 未迁移的不强制，但一旦迁了就要求同步登记（避免清单过期）。
	for _, dir := range exactMoneyNotYet {
		if setsExactMoney(t, dir) {
			t.Errorf("%s 已经设置 ExactMoney，请从 exactMoneyNotYet 移到 exactMoneyMigrated", dir)
		}
	}
}
