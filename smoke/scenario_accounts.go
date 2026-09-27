package smoke

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// The accounts scenario drives the game package's account functions in a
// temp data dir: create two accounts with a save each, switch between them,
// export and import, back up, recover from a recovery code, and wipe. It
// checks that nothing panics, that exports round-trip, that a wipe leaves
// the other account's files byte for byte alone and writes a backup, and
// that bad input (tampered exports, garbage codes) is refused cleanly.

type acctRun struct {
	res   *Result
	steps []string
}

// step runs f, turning a panic or an error into a failure. It reports
// whether f succeeded.
func (a *acctRun) step(name string, f func() error) (ok bool) {
	defer func() {
		if rec := recover(); rec != nil {
			a.res.fail("account_panic", "%s panicked: %v", name, rec).Detail = string(debug.Stack())
			a.steps = append(a.steps, "| "+name+" | panicked |")
			ok = false
		}
	}()
	if err := f(); err != nil {
		a.res.fail("account_"+strings.ReplaceAll(name, " ", "_"), "%s: %v", name, err)
		a.steps = append(a.steps, "| "+name+" | "+cell(err.Error())+" |")
		return false
	}
	a.steps = append(a.steps, "| "+name+" | ok |")
	return true
}

func activeID() string {
	for _, s := range game.ListAccounts() {
		if s.Active {
			return s.AccountID
		}
	}
	return ""
}

func accountIDs() []string {
	var ids []string
	for _, s := range game.ListAccounts() {
		ids = append(ids, s.AccountID)
	}
	sort.Strings(ids)
	return ids
}

func hasID(id string) bool {
	for _, x := range accountIDs() {
		if x == id {
			return true
		}
	}
	return false
}

// slotFiles reads every file in an account's slot, keyed by relative path.
// Call it with some account active: DataDir is then <root>/accounts/<active>.
func slotFiles(id string) (map[string][]byte, error) {
	root := filepath.Join(filepath.Dir(game.DataDir()), id)
	out := map[string][]byte{}
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		rel, _ := filepath.Rel(root, p)
		out[rel] = b
		return err
	})
	return out, err
}

func saveNames() []string {
	names, _ := game.ListSaves()
	return names
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func runAccounts(e *Env, res *Result) {
	a := &acctRun{res: res}
	themeKey := theme.All()[len(theme.All())-1].Key
	var alpha, beta *game.Account

	a.step("create alpha", func() error {
		acct, err := game.CreateAccount("Smoke Alpha")
		if err != nil {
			return err
		}
		alpha = acct
		if _, err := alpha.UnlockTheme(themeKey); err != nil {
			return err
		}
		ge := game.NewGameEngine()
		ge.SetAccount(alpha)
		ge.SeedRNG(e.SeedBase)
		ge.StepTicks(50)
		return ge.SaveGame("alpha-run")
	})
	a.step("create beta", func() error {
		acct, err := game.CreateAccount("Smoke Beta")
		if err != nil {
			return err
		}
		beta = acct
		if activeID() != beta.AccountID {
			return fmt.Errorf("creating beta did not make it active (active %q)", activeID())
		}
		ge := game.NewGameEngine()
		ge.SetAccount(beta)
		return ge.SaveGame("beta-run")
	})
	if alpha == nil || beta == nil {
		res.Summary = "could not create the two accounts"
		return
	}
	if alpha.AccountID == beta.AccountID {
		res.fail("account_ids", "two differently named accounts got the same id %s", alpha.AccountID)
	}
	a.step("saves are per account", func() error {
		if s := saveNames(); !contains(s, "beta-run") || contains(s, "alpha-run") {
			return fmt.Errorf("with beta active the save list is %v", s)
		}
		return nil
	})
	a.step("switch to alpha", func() error {
		if _, err := game.SwitchAccount(alpha.AccountID); err != nil {
			return err
		}
		if activeID() != alpha.AccountID {
			return fmt.Errorf("active account is %q after switching to alpha", activeID())
		}
		if s := saveNames(); !contains(s, "alpha-run") || contains(s, "beta-run") {
			return fmt.Errorf("with alpha active the save list is %v", s)
		}
		return nil
	})
	a.step("switch to a missing account", func() error {
		if _, err := game.SwitchAccount(strings.Repeat("0", 32)); err == nil {
			return fmt.Errorf("switching to an account that does not exist succeeded")
		}
		if activeID() != alpha.AccountID {
			return fmt.Errorf("a failed switch changed the active account to %q", activeID())
		}
		return nil
	})
	a.step("re-create alpha by name", func() error {
		acct, err := game.CreateAccount("Smoke Alpha")
		if err != nil {
			return err
		}
		if acct.AccountID != alpha.AccountID || !acct.HasTheme(themeKey) {
			return fmt.Errorf("re-entering the name gave id %s (want %s), theme kept %v", acct.AccountID, alpha.AccountID, acct.HasTheme(themeKey))
		}
		return nil
	})

	var blob []byte
	a.step("export", func() error {
		b, err := game.ExportAccountByID(alpha.AccountID)
		if err != nil {
			return err
		}
		blob = b
		if !bytes.Contains(b, []byte(alpha.AccountID)) {
			return fmt.Errorf("the export does not carry the account id")
		}
		return nil
	})
	a.step("import round-trip", func() error {
		for _, merge := range []bool{true, false} {
			acct, err := game.ImportAccountExport(blob, merge)
			if err != nil {
				return fmt.Errorf("merge=%v: %w", merge, err)
			}
			if acct.AccountID != alpha.AccountID || !acct.HasTheme(themeKey) {
				return fmt.Errorf("merge=%v: import gave id %s, theme kept %v", merge, acct.AccountID, acct.HasTheme(themeKey))
			}
		}
		again, err := game.ExportAccountByID(alpha.AccountID)
		if err != nil {
			return err
		}
		ta, _ := jsonTree(blob)
		tb, _ := jsonTree(again)
		if d := firstDiff(ta, tb, nil); d != "" {
			return fmt.Errorf("export -> import -> export changed the data at %s", d)
		}
		return nil
	})
	a.step("refuse tampered and garbage exports", func() error {
		var m map[string]interface{}
		if err := json.Unmarshal(blob, &m); err != nil {
			return err
		}
		m["display_name"] = "Mallory"
		bad, _ := json.Marshal(m)
		for name, b := range map[string][]byte{"tampered": bad, "empty": nil, "garbage": []byte("\x00not json"), "empty object": []byte("{}"),
			"unsigned": []byte(fmt.Sprintf(`{"version":1,"account_id":%q}`, alpha.AccountID))} {
			if _, err := game.ImportAccountExport(b, true); err == nil {
				return fmt.Errorf("a %s export was accepted", name)
			}
		}
		return nil
	})

	var backup string
	a.step("backup", func() error {
		p, err := game.BackupAccount(alpha.AccountID)
		if err != nil {
			return err
		}
		backup = p
		for _, want := range []string{"account.json", filepath.Join("saves", "alpha-run.json")} {
			if _, err := os.Stat(filepath.Join(p, want)); err != nil {
				return fmt.Errorf("backup %s is missing %s", p, want)
			}
		}
		if activeID() != alpha.AccountID {
			return fmt.Errorf("backing up changed the active account")
		}
		return nil
	})

	var betaCode string
	a.step("recovery code", func() error {
		code, err := game.RecoveryCodeForID(alpha.AccountID)
		if err != nil {
			return err
		}
		acct, err := game.ImportRecoveryCode(code)
		if err != nil {
			return err
		}
		if acct.AccountID != alpha.AccountID {
			return fmt.Errorf("the recovery code restored id %s, want %s", acct.AccountID, alpha.AccountID)
		}
		if betaCode, err = game.RecoveryCodeForID(beta.AccountID); err != nil {
			return err
		}
		// A typo must be caught, not mint a stranger's account.
		typo := []byte(code)
		for i := len(typo) - 1; i >= 0; i-- {
			if typo[i] != '-' {
				if typo[i] == 'A' {
					typo[i] = 'B'
				} else {
					typo[i] = 'A'
				}
				break
			}
		}
		if acct, err := game.ImportRecoveryCode(string(typo)); err == nil && acct.AccountID == alpha.AccountID {
			return fmt.Errorf("a code with a changed character still restored alpha")
		} else if err == nil {
			res.warn("recovery_typo", "a recovery code with one changed character was accepted as a different account (%s) instead of failing its checksum", acct.AccountID)
		}
		for _, g := range []string{"", "AGEF-", "not a code", strings.Repeat("Z", 200), "💥"} {
			if _, err := game.ImportRecoveryCode(g); err == nil {
				return fmt.Errorf("garbage recovery code %q was accepted", g)
			}
		}
		return nil
	})

	a.step("wipe beta leaves alpha alone", func() error {
		before, err := slotFiles(alpha.AccountID)
		if err != nil {
			return err
		}
		p, err := game.WipeAccountByID(beta.AccountID)
		if err != nil {
			return err
		}
		if p == "" {
			return fmt.Errorf("wiping beta wrote no backup")
		}
		if _, err := os.Stat(filepath.Join(p, "account.json")); err != nil {
			return fmt.Errorf("the wipe backup %s has no account.json", p)
		}
		if hasID(beta.AccountID) {
			return fmt.Errorf("beta is still listed after the wipe")
		}
		if activeID() != alpha.AccountID {
			return fmt.Errorf("wiping beta changed the active account to %q", activeID())
		}
		after, err := slotFiles(alpha.AccountID)
		if err != nil {
			return err
		}
		if d := firstDiff(before, after, nil); d != "" {
			return fmt.Errorf("wiping beta changed alpha's files: %s", d)
		}
		return nil
	})
	a.step("recover a wiped identity", func() error {
		acct, err := game.ImportRecoveryCode(betaCode)
		if err != nil {
			return err
		}
		if acct.AccountID != beta.AccountID {
			return fmt.Errorf("beta's code restored %s", acct.AccountID)
		}
		return nil
	})
	a.step("wipe the active account and restore it from its export", func() error {
		if _, err := game.SwitchAccount(alpha.AccountID); err != nil {
			return err
		}
		if _, err := game.WipeAccountByID(alpha.AccountID); err != nil {
			return err
		}
		if hasID(alpha.AccountID) {
			return fmt.Errorf("alpha is still listed after wiping it")
		}
		acct, err := game.ImportAccountExport(blob, true)
		if err != nil {
			return err
		}
		if acct.AccountID != alpha.AccountID || !acct.HasTheme(themeKey) {
			return fmt.Errorf("restoring alpha gave id %s, theme %v", acct.AccountID, acct.HasTheme(themeKey))
		}
		return nil
	})
	res.Summary = fmt.Sprintf("%d step(s), %d account(s) left, backup at %s", len(a.steps), len(accountIDs()), filepath.Base(backup))
	res.section("Steps", "| step | result |\n|---|---|\n%s", strings.Join(a.steps, "\n"))
}
