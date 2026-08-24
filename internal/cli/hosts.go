package cli

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/hosts"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// NewHostsCmd creates the "hosts" command group that manages the ordered
// host→user routing rules consumed by the clone hook. The rule list order is
// significant: exact matches beat wildcards, and earlier rules win.
func NewHostsCmd(store domain.UserStore, hostStore domain.HostRuleStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hosts",
		Short: i18n.T("cli.hosts.short", nil),
	}

	cmd.AddCommand(
		newHostsSetCmd(store, hostStore),
		newHostsListCmd(hostStore),
		newHostsRemoveCmd(hostStore),
		newHostsMoveCmd(hostStore),
	)

	return cmd
}

// newHostsSetCmd creates the "hosts set" command. Setting an existing host
// updates its email in place without changing the rule order.
func newHostsSetCmd(store domain.UserStore, hostStore domain.HostRuleStore) *cobra.Command {
	return &cobra.Command{
		Use:   "set <host> <email>",
		Short: i18n.T("cli.hosts.set.short", nil),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := strings.ToLower(strings.TrimSpace(args[0]))
			email := strings.TrimSpace(args[1])

			if err := hosts.ValidateHost(host); err != nil {
				return errors.New(i18n.T("cli.hosts.set.invalid_host", map[string]interface{}{"Host": host}))
			}

			_, found, err := findSavedUser(store, email)
			if err != nil {
				return err
			}
			if !found {
				return errors.New(i18n.T("cli.hosts.set.email_not_found", map[string]interface{}{"Email": email}))
			}

			rules, err := hostStore.List()
			if err != nil {
				return err
			}

			// 就地更新已存在的 host,保持配置顺序不变
			updated := false
			for i := range rules {
				if rules[i].Host == host {
					rules[i].Email = email
					updated = true
					break
				}
			}
			if !updated {
				rules = append(rules, domain.HostRule{Host: host, Email: email})
			}

			if err := hostStore.SaveAll(rules); err != nil {
				return err
			}

			fmt.Println(i18n.T("cli.hosts.set.success", map[string]interface{}{
				"Host": host, "Email": email,
			}))
			return nil
		},
	}
}

// newHostsListCmd creates the "hosts list" command with numbered rows that
// can be referenced by "hosts move".
func newHostsListCmd(hostStore domain.HostRuleStore) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   i18n.T("cli.hosts.list.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			rules, err := hostStore.List()
			if err != nil {
				return err
			}

			if len(rules) == 0 {
				fmt.Println(i18n.T("cli.hosts.list.empty", nil))
				return nil
			}

			for i, rule := range rules {
				fmt.Println(i18n.T("cli.hosts.list_row", map[string]interface{}{
					"Index": i, "Host": rule.Host, "Email": rule.Email,
				}))
			}
			return nil
		},
	}
}

// newHostsRemoveCmd creates the "hosts remove" command, deleting a rule by
// its exact hostname or wildcard pattern.
func newHostsRemoveCmd(hostStore domain.HostRuleStore) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <host>",
		Short: i18n.T("cli.hosts.remove.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := strings.ToLower(strings.TrimSpace(args[0]))

			rules, err := hostStore.List()
			if err != nil {
				return err
			}

			idx := slices.IndexFunc(rules, func(r domain.HostRule) bool {
				return strings.ToLower(r.Host) == host
			})
			if idx < 0 {
				return errors.New(i18n.T("cli.hosts.remove.not_found", map[string]interface{}{"Host": host}))
			}

			rules = slices.Delete(rules, idx, idx+1)
			if err := hostStore.SaveAll(rules); err != nil {
				return err
			}

			fmt.Println(i18n.T("cli.hosts.remove.success", map[string]interface{}{"Host": host}))
			return nil
		},
	}
}

// newHostsMoveCmd creates the "hosts move" command reordering a rule.
// The target is a 0-based index, "first", "last", "up" or "down".
func newHostsMoveCmd(hostStore domain.HostRuleStore) *cobra.Command {
	return &cobra.Command{
		Use:   "move <host> <target>",
		Short: i18n.T("cli.hosts.move.short", nil),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := strings.ToLower(strings.TrimSpace(args[0]))
			target := strings.ToLower(strings.TrimSpace(args[1]))

			rules, err := hostStore.List()
			if err != nil {
				return err
			}

			idx := slices.IndexFunc(rules, func(r domain.HostRule) bool {
				return strings.ToLower(r.Host) == host
			})
			if idx < 0 {
				return errors.New(i18n.T("cli.hosts.remove.not_found", map[string]interface{}{"Host": host}))
			}

			newIdx, err := resolveMoveTarget(target, idx, len(rules), host)
			if err != nil {
				return err
			}

			rule := rules[idx]
			rules = slices.Delete(rules, idx, idx+1)
			rules = slices.Insert(rules, newIdx, rule)

			if err := hostStore.SaveAll(rules); err != nil {
				return err
			}

			fmt.Println(i18n.T("cli.hosts.move.success", map[string]interface{}{
				"Host": host, "Index": newIdx,
			}))
			return nil
		},
	}
}

// resolveMoveTarget converts a move target string into a destination index.
// It validates bounds and reports moved-at-boundary errors for up/down.
func resolveMoveTarget(target string, current, length int, host string) (int, error) {
	switch target {
	case "first":
		return 0, nil
	case "last":
		return length - 1, nil
	case "up":
		if current == 0 {
			return 0, errors.New(i18n.T("cli.hosts.move.at_boundary", map[string]interface{}{
				"Host": host, "Boundary": "first",
			}))
		}
		return current - 1, nil
	case "down":
		if current == length-1 {
			return 0, errors.New(i18n.T("cli.hosts.move.at_boundary", map[string]interface{}{
				"Host": host, "Boundary": "last",
			}))
		}
		return current + 1, nil
	}

	index, err := strconv.Atoi(target)
	if err != nil {
		return 0, errors.New(i18n.T("cli.hosts.move.invalid_target", map[string]interface{}{"Target": target}))
	}
	if index < 0 || index >= length {
		return 0, errors.New(i18n.T("cli.hosts.move.out_of_range", map[string]interface{}{
			"Index": index, "Max": length - 1,
		}))
	}
	return index, nil
}

// findSavedUser returns the saved user whose email matches exactly.
func findSavedUser(store domain.UserStore, email string) (domain.User, bool, error) {
	users, err := store.List()
	if err != nil {
		return domain.User{}, false, err
	}
	for _, u := range users {
		if u.Email == email {
			return u, true, nil
		}
	}
	return domain.User{}, false, nil
}
