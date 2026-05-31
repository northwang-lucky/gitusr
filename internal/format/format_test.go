package format

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// captureStderr runs f and returns whatever was written to stderr.
func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// captureStdout runs f and returns whatever was written to stdout.
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintErr(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	output := captureStderr(func() {
		PrintErr("something went wrong")
	})

	expected := "gitusr error: something went wrong\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestPrintErr_EmptyMsg(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	output := captureStderr(func() {
		PrintErr("")
	})

	expected := "gitusr error: \n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestPrintUserInfo_Repo(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	user := domain.User{Name: "Alice", Email: "alice@example.com"}
	opts := PrintOptions{Global: false, ShowSuccess: false}

	output := captureStdout(func() {
		PrintUserInfo(user, opts)
	})

	expected := "Your repo git user is:\n\nuser.name  = Alice\nuser.email = alice@example.com\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestPrintUserInfo_Global(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	user := domain.User{Name: "Bob", Email: "bob@example.com"}
	opts := PrintOptions{Global: true, ShowSuccess: false}

	output := captureStdout(func() {
		PrintUserInfo(user, opts)
	})

	expected := "Your global git user is:\n\nuser.name  = Bob\nuser.email = bob@example.com\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestPrintUserInfo_WithSuccess(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	user := domain.User{Name: "Charlie", Email: "charlie@example.com"}
	opts := PrintOptions{Global: false, ShowSuccess: true}

	output := captureStdout(func() {
		PrintUserInfo(user, opts)
	})

	expected := "Success!\nYour repo git user is:\n\nuser.name  = Charlie\nuser.email = charlie@example.com\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestFormatUserList_Empty(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	result := FormatUserList(nil)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}

	result = FormatUserList([]domain.User{})
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestFormatUserList_Single(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	users := []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}

	result := FormatUserList(users)
	expected := "0: Name: Alice | Email: alice@example.com\n"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatUserList_MultipleAligned(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	users := []domain.User{
		{Name: "Alice", Email: "a@t.com"},
		{Name: "Bob", Email: "b@t.com"},
		{Name: "Charlie", Email: "c@t.com"},
	}

	result := FormatUserList(users)

	// Names are padded to "Charlie" length (7 chars), plus a literal
	// space before the pipe from the format string " | Email: ".
	lines := strings.Split(strings.TrimSuffix(result, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	expected0 := "0: Name: Alice   | Email: a@t.com"
	expected1 := "1: Name: Bob     | Email: b@t.com"
	expected2 := "2: Name: Charlie | Email: c@t.com"

	if lines[0] != expected0 {
		t.Errorf("line 0:\nexpected %q\ngot      %q", expected0, lines[0])
	}
	if lines[1] != expected1 {
		t.Errorf("line 1:\nexpected %q\ngot      %q", expected1, lines[1])
	}
	if lines[2] != expected2 {
		t.Errorf("line 2:\nexpected %q\ngot      %q", expected2, lines[2])
	}
}

func TestPrintErr_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	output := captureStderr(func() {
		PrintErr("something went wrong")
	})

	expected := "gitusr error: something went wrong\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestPrintErr_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	output := captureStderr(func() {
		PrintErr("错误信息")
	})

	expected := "gitusr 错误：错误信息\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestPrintUserInfo_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	user := domain.User{Name: "Alice", Email: "alice@example.com"}
	opts := PrintOptions{Global: false, ShowSuccess: false}

	output := captureStdout(func() {
		PrintUserInfo(user, opts)
	})

	expected := "Your repo git user is:\n\nuser.name  = Alice\nuser.email = alice@example.com\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestPrintUserInfo_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	user := domain.User{Name: "Alice", Email: "alice@example.com"}
	opts := PrintOptions{Global: false, ShowSuccess: true}

	output := captureStdout(func() {
		PrintUserInfo(user, opts)
	})

	expected := "成功！\n您的 repo git 用户为：\n\nuser.name  = Alice\nuser.email = alice@example.com\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestFormatUserList_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	users := []domain.User{
		{Name: "Alice", Email: "a@t.com"},
		{Name: "Bob", Email: "b@t.com"},
	}

	result := FormatUserList(users)

	lines := strings.Split(strings.TrimSuffix(result, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	expected0 := "0: Name: Alice | Email: a@t.com"
	expected1 := "1: Name: Bob   | Email: b@t.com"

	if lines[0] != expected0 {
		t.Errorf("line 0:\nexpected %q\ngot      %q", expected0, lines[0])
	}
	if lines[1] != expected1 {
		t.Errorf("line 1:\nexpected %q\ngot      %q", expected1, lines[1])
	}
}

func TestFormatUserList_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	users := []domain.User{
		{Name: "Alice", Email: "a@t.com"},
		{Name: "Bob", Email: "b@t.com"},
	}

	result := FormatUserList(users)

	lines := strings.Split(strings.TrimSuffix(result, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	expected0 := "0：姓名：Alice | 邮箱：a@t.com"
	expected1 := "1：姓名：Bob   | 邮箱：b@t.com"

	if lines[0] != expected0 {
		t.Errorf("line 0:\nexpected %q\ngot      %q", expected0, lines[0])
	}
	if lines[1] != expected1 {
		t.Errorf("line 1:\nexpected %q\ngot      %q", expected1, lines[1])
	}
}
