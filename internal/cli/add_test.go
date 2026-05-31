package cli

import (
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// addTestStore implements domain.UserStore with configurable behaviour for add tests.
type addTestStore struct {
	initialized bool
	users       []domain.User
	listErr     error
	addErr      error
	addedUsers  []domain.User
}

func (m *addTestStore) List() ([]domain.User, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.users, nil
}

func (m *addTestStore) Add(user domain.User) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.users = append(m.users, user)
	m.addedUsers = append(m.addedUsers, user)
	return nil
}

func (m *addTestStore) Remove(index int) error         { return nil }
func (m *addTestStore) SaveAll(users []domain.User) error { return nil }
func (m *addTestStore) IsInitialized() bool              { return m.initialized }

func TestAdd_NotInitialized(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &addTestStore{initialized: false}
	cmd := NewAddCmd(store)

	_, stderr, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error when store is not initialized")
	}

	if !strings.Contains(err.Error(), "store not initialized") {
		t.Errorf("error should contain 'store not initialized', got: %q", err.Error())
	}

	// Cobra prints errors to stderr automatically
	if !strings.Contains(stderr, "store not initialized") {
		t.Errorf("stderr should contain 'store not initialized', got: %q", stderr)
	}
}

func TestAdd_DuplicateName(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &addTestStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	// Override the prompt function to return a duplicate name
	origAskNewUser := askNewUser
	askNewUser = func() (domain.User, error) {
		return domain.User{Name: "Alice", Email: "alice2@example.com"}, nil
	}
	defer func() { askNewUser = origAskNewUser }()

	cmd := NewAddCmd(store)
	_, stderr, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error for duplicate name")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should contain 'already exists', got: %q", err.Error())
	}

	if !strings.Contains(err.Error(), "Alice") {
		t.Errorf("error should contain duplicate name, got: %q", err.Error())
	}

	if !strings.Contains(stderr, "already exists") {
		t.Errorf("stderr should contain 'already exists', got: %q", stderr)
	}
}

func TestAdd_DuplicateEmail(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &addTestStore{
		initialized: true,
		users: []domain.User{
			{Name: "Bob", Email: "bob@example.com"},
		},
	}

	origAskNewUser := askNewUser
	askNewUser = func() (domain.User, error) {
		return domain.User{Name: "Bob2", Email: "bob@example.com"}, nil
	}
	defer func() { askNewUser = origAskNewUser }()

	cmd := NewAddCmd(store)
	_, stderr, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error for duplicate email")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should contain 'already exists', got: %q", err.Error())
	}

	if !strings.Contains(err.Error(), "bob@example.com") {
		t.Errorf("error should contain duplicate email, got: %q", err.Error())
	}

	if !strings.Contains(stderr, "already exists") {
		t.Errorf("stderr should contain 'already exists', got: %q", stderr)
	}
}

func TestAdd_Success(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &addTestStore{
		initialized: true,
		users: []domain.User{
			{Name: "Existing", Email: "existing@example.com"},
		},
	}

	origAskNewUser := askNewUser
	askNewUser = func() (domain.User, error) {
		return domain.User{Name: "Charlie", Email: "charlie@example.com"}, nil
	}
	defer func() { askNewUser = origAskNewUser }()

	cmd := NewAddCmd(store)
	stdout, stderr, err := executeCmd(cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stderr != "" {
		t.Errorf("stderr should be empty, got: %q", stderr)
	}

	if !strings.Contains(stdout, "Success!") {
		t.Errorf("stdout should contain 'Success!', got: %q", stdout)
	}

	if !strings.Contains(stdout, "Charlie") {
		t.Errorf("stdout should contain user name 'Charlie', got: %q", stdout)
	}

	if !strings.Contains(stdout, "charlie@example.com") {
		t.Errorf("stdout should contain user email 'charlie@example.com', got: %q", stdout)
	}

	// Verify the user was added to the store
	if len(store.addedUsers) != 1 {
		t.Fatalf("expected 1 user added, got %d", len(store.addedUsers))
	}

	added := store.addedUsers[0]
	if added.Name != "Charlie" {
		t.Errorf("added user name = %q, want %q", added.Name, "Charlie")
	}
	if added.Email != "charlie@example.com" {
		t.Errorf("added user email = %q, want %q", added.Email, "charlie@example.com")
	}
}

// TestAddSuccess_En verifies the success message in English.
func TestAddSuccess_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &addTestStore{
		initialized: true,
		users:       []domain.User{},
	}

	origAskNewUser := askNewUser
	askNewUser = func() (domain.User, error) {
		return domain.User{Name: "Charlie", Email: "charlie@example.com"}, nil
	}
	defer func() { askNewUser = origAskNewUser }()

	cmd := NewAddCmd(store)
	stdout, stderr, err := executeCmd(cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stderr != "" {
		t.Errorf("stderr should be empty, got: %q", stderr)
	}

	want := "Success! User (name: Charlie | email: charlie@example.com) has been saved!"
	if !strings.Contains(stdout, want) {
		t.Errorf("stdout should contain %q, got: %q", want, stdout)
	}
}

// TestAddSuccess_ZhCN verifies the success message in Chinese.
func TestAddSuccess_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	store := &addTestStore{
		initialized: true,
		users:       []domain.User{},
	}

	origAskNewUser := askNewUser
	askNewUser = func() (domain.User, error) {
		return domain.User{Name: "张三", Email: "zhangsan@example.com"}, nil
	}
	defer func() { askNewUser = origAskNewUser }()

	cmd := NewAddCmd(store)
	stdout, stderr, err := executeCmd(cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stderr != "" {
		t.Errorf("stderr should be empty, got: %q", stderr)
	}

	want := "成功！用户（姓名：张三 | 邮箱：zhangsan@example.com）已保存！"
	if !strings.Contains(stdout, want) {
		t.Errorf("stdout should contain %q, got: %q", want, stdout)
	}
}

// TestAddDupName_En verifies the duplicate-name error in English.
// --- non-TTY flag tests ----------------------------------------------------

// TestAdd_WithFlags_Success verifies that --name and --email flags add a user
// without invoking the interactive prompt.
func TestAdd_WithFlags_Success(t *testing.T) {
	initI18nForTest("en")

	store := &addTestStore{
		initialized: true,
		users:       []domain.User{},
	}

	cmd := NewAddCmd(store)
	stdout, stderr, err := executeCmd(cmd, "--name", "FlagUser", "--email", "flag@test.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stderr != "" {
		t.Errorf("stderr should be empty, got: %q", stderr)
	}

	if !strings.Contains(stdout, "Success!") {
		t.Errorf("stdout should contain 'Success!', got: %q", stdout)
	}

	if !strings.Contains(stdout, "FlagUser") {
		t.Errorf("stdout should contain user name 'FlagUser', got: %q", stdout)
	}

	if !strings.Contains(stdout, "flag@test.com") {
		t.Errorf("stdout should contain email 'flag@test.com', got: %q", stdout)
	}

	// Verify the user was added to the store
	if len(store.addedUsers) != 1 {
		t.Fatalf("expected 1 user added, got %d", len(store.addedUsers))
	}

	added := store.addedUsers[0]
	if added.Name != "FlagUser" {
		t.Errorf("added user name = %q, want %q", added.Name, "FlagUser")
	}
	if added.Email != "flag@test.com" {
		t.Errorf("added user email = %q, want %q", added.Email, "flag@test.com")
	}
}

// TestAdd_WithFlags_Partial verifies that --name without --email returns an
// error about requiring both flags.
func TestAdd_WithFlags_Partial(t *testing.T) {
	initI18nForTest("en")

	store := &addTestStore{
		initialized: true,
		users:       []domain.User{},
	}

	cmd := NewAddCmd(store)
	_, stderr, err := executeCmd(cmd, "--name", "OnlyName")

	if err == nil {
		t.Fatal("expected error when only --name is provided")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error should contain 'required', got: %q", err.Error())
	}

	if !strings.Contains(stderr, "required") {
		t.Errorf("stderr should contain 'required', got: %q", stderr)
	}

	// Verify no user was added
	if len(store.addedUsers) != 0 {
		t.Errorf("expected 0 users added, got %d", len(store.addedUsers))
	}

	// Also test --email only
	cmd2 := NewAddCmd(&addTestStore{initialized: true, users: []domain.User{}})
	_, _, err2 := executeCmd(cmd2, "--email", "only@email.com")

	if err2 == nil {
		t.Fatal("expected error when only --email is provided")
	}

	if !strings.Contains(err2.Error(), "required") {
		t.Errorf("error should contain 'required', got: %q", err2.Error())
	}
}

// TestAdd_WithFlags_InvalidEmail verifies that an invalid --email value is
// rejected with a format error.
func TestAdd_WithFlags_InvalidEmail(t *testing.T) {
	initI18nForTest("en")

	store := &addTestStore{
		initialized: true,
		users:       []domain.User{},
	}

	cmd := NewAddCmd(store)
	_, stderr, err := executeCmd(cmd, "--name", "BadEmail", "--email", "not-an-email")

	if err == nil {
		t.Fatal("expected error for invalid email format")
	}

	if !strings.Contains(err.Error(), "invalid email") {
		t.Errorf("error should contain 'invalid email', got: %q", err.Error())
	}

	if !strings.Contains(stderr, "invalid email") {
		t.Errorf("stderr should contain 'invalid email', got: %q", stderr)
	}

	// Verify no user was added
	if len(store.addedUsers) != 0 {
		t.Errorf("expected 0 users added, got %d", len(store.addedUsers))
	}
}

// TestAdd_WithFlags_DuplicateName verifies that adding a duplicate name via
// flags returns the same duplicate error as interactive mode.
func TestAdd_WithFlags_DuplicateName(t *testing.T) {
	initI18nForTest("en")

	store := &addTestStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewAddCmd(store)
	_, stderr, err := executeCmd(cmd, "--name", "Alice", "--email", "alice2@example.com")

	if err == nil {
		t.Fatal("expected error for duplicate name")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should contain 'already exists', got: %q", err.Error())
	}

	if !strings.Contains(err.Error(), "Alice") {
		t.Errorf("error should contain duplicate name 'Alice', got: %q", err.Error())
	}

	if !strings.Contains(stderr, "already exists") {
		t.Errorf("stderr should contain 'already exists', got: %q", stderr)
	}

	// Verify no user was added
	if len(store.addedUsers) != 0 {
		t.Errorf("expected 0 users added, got %d", len(store.addedUsers))
	}
}

// --- existing tests ----------------------------------------------------

func TestAddDupName_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &addTestStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	origAskNewUser := askNewUser
	askNewUser = func() (domain.User, error) {
		return domain.User{Name: "Alice", Email: "alice2@example.com"}, nil
	}
	defer func() { askNewUser = origAskNewUser }()

	cmd := NewAddCmd(store)
	_, stderr, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error for duplicate name")
	}

	want := `user with name "Alice" already exists`
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error should contain %q, got: %q", want, err.Error())
	}

	if !strings.Contains(stderr, want) {
		t.Errorf("stderr should contain %q, got: %q", stderr, want)
	}
}

// TestAddDupName_ZhCN verifies the duplicate-name error in Chinese.
func TestAddDupName_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	store := &addTestStore{
		initialized: true,
		users: []domain.User{
			{Name: "张三", Email: "zhangsan@example.com"},
		},
	}

	origAskNewUser := askNewUser
	askNewUser = func() (domain.User, error) {
		return domain.User{Name: "张三", Email: "zhangsan2@example.com"}, nil
	}
	defer func() { askNewUser = origAskNewUser }()

	cmd := NewAddCmd(store)
	_, stderr, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error for duplicate name")
	}

	want := `用户名为"张三"的用户已存在`
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error should contain %q, got: %q", want, err.Error())
	}

	if !strings.Contains(stderr, want) {
		t.Errorf("stderr should contain %q, got: %q", stderr, want)
	}
}
