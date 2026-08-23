// gitu keeps selected, intentionally untracked files in one personal Git vault.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type config struct {
	Version     int      `json:"version"`
	ProjectID   string   `json:"projectId"`
	StorageMode string   `json:"storageMode"`
	Paths       []string `json:"paths"`
}
type manifest struct {
	Version      int      `json:"version"`
	ProjectID    string   `json:"projectId"`
	StorageMode  string   `json:"storageMode"`
	Branch       string   `json:"branch,omitempty"`
	HostRemote   string   `json:"hostRemote,omitempty"`
	Paths        []string `json:"paths"`
	LastSyncedAt string   `json:"lastSyncedAt"`
}
type projectState struct {
	Version     int    `json:"version"`
	ProjectID   string `json:"projectId"`
	StorageMode string `json:"storageMode"`
}
type project struct{ root, config string }

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "vault":
		err = vaultCommand(os.Args[2:])
	case "branch":
		err = branchCommand(os.Args[2:])
	case "project":
		err = projectCommand(os.Args[2:])
	case "init":
		err = initCommand(os.Args[2:])
	case "add":
		err = addCommand(os.Args[2:])
	case "list":
		err = listCommand()
	case "status":
		err = statusCommand()
	case "sync":
		err = syncCommand(os.Args[2:])
	case "restore":
		err = restoreCommand(os.Args[2:])
	case "help", "-h", "--help":
		usage()
		return
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `gitu — Git history for explicitly selected untracked files

Usage:
  gitu vault init --remote URL      clone or connect your one personal vault
  gitu vault pull                   fetch the latest vault history
  gitu init [--project ID]          connect this working repository to a vault folder

  gitu branch enable|status         use independent vault files for each Git branch
  gitu branch rename --from OLD --to NEW
                                    move a branch's vault history after a Git rename
  gitu project rename --to ID       move vault history after a project identity rename
  gitu add [--allow-sensitive] PATH... register only these untracked paths
  gitu list | status | sync [-m MESSAGE] | restore [--force] [PATH...]

The vault is a single ordinary Git repository. Files are namespaced by project;
the host repository is never modified.
`)
}

func vaultCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("use `gitu vault init --remote URL` or `gitu vault pull`")
	}
	switch args[0] {
	case "init":
		fs := flag.NewFlagSet("vault init", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		remote := fs.String("remote", "", "personal vault Git URL")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *remote == "" {
			return errors.New("--remote is required; choose one personal private repository")
		}
		dir, err := vaultDir()
		if err != nil {
			return err
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			if err := run(dir, "git", "remote", "set-url", "origin", *remote); err != nil {
				return err
			}
			fmt.Printf("Connected existing vault at %s\n", dir)
			return nil
		}
		if _, err := os.Stat(dir); err == nil {
			entries, readErr := os.ReadDir(dir)
			if readErr != nil {
				return readErr
			}
			if len(entries) != 0 {
				return fmt.Errorf("vault directory %s exists but is not a Git repository", dir)
			}
		}
		if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
			return err
		}
		if err := run(".", "git", "clone", "--quiet", *remote, dir); err != nil {
			return err
		}
		fmt.Printf("Vault ready at %s\n", dir)
		return nil
	case "pull":
		dir, err := readyVault()
		if err != nil {
			return err
		}
		if err := run(dir, "git", "pull", "--ff-only"); err != nil {
			return err
		}
		fmt.Println("Vault is up to date.")
		return nil
	default:
		return fmt.Errorf("unknown vault command %q", args[0])
	}
}

func branchCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("use `gitu branch enable`, `gitu branch status`, or `gitu branch rename --from OLD --to NEW`")
	}
	p, cfg, err := openProject()
	if err != nil {
		return err
	}
	switch args[0] {
	case "enable":
		if cfg.StorageMode == "branch" {
			fmt.Println("Branch-scoped storage is already enabled.")
			return nil
		}
		cfg.StorageMode = "branch"
		if err := saveConfig(p, cfg); err != nil {
			return err
		}
		fmt.Println("Branch-scoped storage enabled. Existing shared history remains in the vault; the current branch starts a separate history on its next sync.")
		return nil
	case "status":
		if cfg.StorageMode != "branch" {
			fmt.Println("Shared storage is active. Run `gitu branch enable` to isolate files by branch.")
			return nil
		}
		branch, err := currentBranch(p.root)
		if err != nil {
			return err
		}
		fmt.Printf("Branch-scoped storage is active for %q.\n", branch)
		return nil
	case "rename":
		fs := flag.NewFlagSet("branch rename", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		from := fs.String("from", "", "previous Git branch name")
		to := fs.String("to", "", "new Git branch name")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if cfg.StorageMode != "branch" {
			return errors.New("branch-scoped storage is not enabled")
		}
		if *from == "" || *to == "" {
			return errors.New("both --from and --to are required")
		}
		if branchKey(*from) == branchKey(*to) {
			return errors.New("the branch names resolve to the same vault identity")
		}
		dir, err := readyVault()
		if err != nil {
			return err
		}
		oldPath := filepath.Join(dir, "projects", cfg.ProjectID, "branches", branchKey(*from))
		newPath := filepath.Join(dir, "projects", cfg.ProjectID, "branches", branchKey(*to))
		if _, err := os.Stat(oldPath); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("no vault history exists for branch %q", *from)
		}
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("vault history already exists for branch %q", *to)
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
		if err := commitVaultMove(dir, filepath.Join("projects", cfg.ProjectID), "Rename branch "+*from+" to "+*to); err != nil {
			return err
		}
		fmt.Printf("Moved vault history from %q to %q.\n", *from, *to)
		return nil
	default:
		return fmt.Errorf("unknown branch command %q", args[0])
	}
}

func projectCommand(args []string) error {
	if len(args) == 0 || args[0] != "rename" {
		return errors.New("use `gitu project rename --to NEW-ID`")
	}
	fs := flag.NewFlagSet("project rename", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	to := fs.String("to", "", "new stable project ID")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if slug(*to) == "" {
		return errors.New("--to must contain letters or numbers")
	}
	p, cfg, err := openProject()
	if err != nil {
		return err
	}
	newID := slug(*to)
	if newID == cfg.ProjectID {
		fmt.Println("Project already has that vault identity.")
		return nil
	}
	dir, err := readyVault()
	if err != nil {
		return err
	}
	oldPath := filepath.Join(dir, "projects", cfg.ProjectID)
	newPath := filepath.Join(dir, "projects", newID)
	if _, err := os.Stat(oldPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("no vault history exists for project %q", cfg.ProjectID)
	}
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("vault history already exists for project %q", newID)
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	cfg.ProjectID = newID
	if err := saveConfig(p, cfg); err != nil {
		return err
	}
	if err := commitVaultMove(dir, "projects", "Rename project vault identity to "+newID); err != nil {
		return err
	}
	fmt.Printf("Moved vault history to project %q. Use `gitu init --project %s` in other checkouts.\n", newID, newID)
	return nil
}

func initCommand(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	projectID := fs.String("project", "", "stable vault project ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	p, err := discoverProject()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p.config), 0755); err != nil {
		return err
	}
	var cfg config
	if data, err := os.ReadFile(p.config); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("read config: %w", err)
		}
	}
	if *projectID != "" {
		cfg.ProjectID = slug(*projectID)
		if cfg.ProjectID == "" {
			return errors.New("project ID must contain letters or numbers")
		}
	}
	if cfg.ProjectID == "" {
		remote, err := hostRemote(p.root)
		if err != nil || remote == "" {
			return errors.New("this repository has no origin; choose a stable name with `gitu init --project my-project`")
		}
		cfg.ProjectID = "repo-" + shortHash(normalizeRemote(remote))
	}
	cfg.Version = 1
	if cfg.StorageMode == "" {
		cfg.StorageMode = "shared"
	}
	if dir, err := vaultDir(); err == nil {
		if state, ok := readProjectState(filepath.Join(dir, "projects", cfg.ProjectID, "project.json")); ok {
			cfg.StorageMode = state.StorageMode
		}
		manifestPath := filepath.Join(dir, "projects", cfg.ProjectID, "manifest.json")
		if cfg.StorageMode == "branch" {
			_, path, _, locationErr := vaultLocation(dir, p, cfg)
			if locationErr != nil {
				return locationErr
			}
			manifestPath = path
		}
		if incoming, ok := readManifest(manifestPath); ok && len(cfg.Paths) == 0 {
			cfg.Paths = incoming.Paths
			fmt.Printf("Imported %d registered path(s) from vault.\n", len(cfg.Paths))
		}
	}
	if err := saveConfig(p, cfg); err != nil {
		return err
	}
	fmt.Printf("Connected this checkout to vault project %q.\n", cfg.ProjectID)
	return nil
}

func addCommand(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	allowSensitive := fs.Bool("allow-sensitive", false, "allow sensitive paths")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return errors.New("provide at least one path")
	}
	p, cfg, err := openProject()
	if err != nil {
		return err
	}
	for _, raw := range fs.Args() {
		rel, err := cleanProjectPath(p, raw)
		if err != nil {
			return err
		}
		full := filepath.Join(p.root, rel)
		if _, err := os.Lstat(full); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		if err := validateCandidate(p, full, rel, *allowSensitive); err != nil {
			return err
		}
		if !contains(cfg.Paths, rel) {
			cfg.Paths = append(cfg.Paths, rel)
			fmt.Printf("Registered %s\n", rel)
		}
	}
	sort.Strings(cfg.Paths)
	return saveConfig(p, cfg)
}

func listCommand() error {
	_, cfg, err := openProject()
	if err != nil {
		return err
	}
	if len(cfg.Paths) == 0 {
		fmt.Println("No paths registered.")
		return nil
	}
	for _, path := range cfg.Paths {
		fmt.Println(path)
	}
	return nil
}

func statusCommand() error {
	p, cfg, err := openProject()
	if err != nil {
		return err
	}
	dir, err := readyVault()
	if err != nil {
		return err
	}
	if len(cfg.Paths) == 0 {
		fmt.Println("No paths registered.")
		return nil
	}
	changes := 0
	files, _, _, err := vaultLocation(dir, p, cfg)
	if err != nil {
		return err
	}
	for _, rel := range cfg.Paths {
		state, err := comparePaths(filepath.Join(p.root, rel), filepath.Join(files, rel))
		if err != nil {
			return err
		}
		if state != "up to date" {
			changes++
			fmt.Printf("%-13s %s\n", state, rel)
		}
	}
	if changes == 0 {
		fmt.Println("All registered paths are up to date.")
	}
	return nil
}

func syncCommand(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	message := fs.String("m", "", "commit message")
	if err := fs.Parse(args); err != nil {
		return err
	}
	p, cfg, err := openProject()
	if err != nil {
		return err
	}
	if len(cfg.Paths) == 0 {
		return errors.New("no paths registered; use `gitu add PATH` first")
	}
	dir, err := readyVault()
	if err != nil {
		return err
	}
	if staged, err := gitHasStagedChanges(dir); err != nil {
		return err
	} else if staged {
		return errors.New("the vault has staged changes; commit or unstage them before syncing")
	}
	files, manifestPath, branch, err := vaultLocation(dir, p, cfg)
	if err != nil {
		return err
	}
	changed := false
	for _, rel := range cfg.Paths {
		source, target := filepath.Join(p.root, rel), filepath.Join(files, rel)
		state, err := comparePaths(source, target)
		if err != nil {
			return err
		}
		if state == "up to date" {
			continue
		}
		if err := os.RemoveAll(target); err != nil {
			return err
		}
		if _, err := os.Lstat(source); err == nil {
			if err := copyPath(source, target); err != nil {
				return err
			}
		}
		changed = true
	}
	remote, _ := hostRemote(p.root)
	m := manifest{Version: 1, ProjectID: cfg.ProjectID, StorageMode: cfg.StorageMode, Branch: branch, HostRemote: remote, Paths: cfg.Paths, LastSyncedAt: time.Now().UTC().Format(time.RFC3339)}
	if err := writeJSON(manifestPath, m); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, "projects", cfg.ProjectID, "project.json"), projectState{Version: 1, ProjectID: cfg.ProjectID, StorageMode: cfg.StorageMode}); err != nil {
		return err
	}
	if err := run(dir, "git", "add", "--", filepath.Join("projects", cfg.ProjectID)); err != nil {
		return err
	}
	hasChanges, err := gitHasStagedChanges(dir)
	if err != nil {
		return err
	}
	if !hasChanges && !changed {
		fmt.Println("Nothing to sync.")
		return nil
	}
	if *message == "" {
		*message = "Sync " + cfg.ProjectID + " — " + time.Now().Format("2006-01-02 15:04:05")
	}
	if err := run(dir, "git", "commit", "--quiet", "-m", *message); err != nil {
		return err
	}
	if err := run(dir, "git", "push"); err != nil {
		return err
	}
	fmt.Printf("Saved and pushed %q to your personal vault.\n", cfg.ProjectID)
	return nil
}

func restoreCommand(args []string) error {
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	force := fs.Bool("force", false, "overwrite existing files")
	if err := fs.Parse(args); err != nil {
		return err
	}
	p, cfg, err := openProject()
	if err != nil {
		return err
	}
	dir, err := readyVault()
	if err != nil {
		return err
	}
	paths := cfg.Paths
	if fs.NArg() > 0 {
		paths = nil
		for _, raw := range fs.Args() {
			rel, err := cleanProjectPath(p, raw)
			if err != nil {
				return err
			}
			if !contains(cfg.Paths, rel) {
				return fmt.Errorf("%s is not registered", rel)
			}
			paths = append(paths, rel)
		}
	}
	files, _, _, err := vaultLocation(dir, p, cfg)
	if err != nil {
		return err
	}
	for _, rel := range paths {
		source, target := filepath.Join(files, rel), filepath.Join(p.root, rel)
		if _, err := os.Lstat(source); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Skipped %s (not in vault)\n", rel)
			continue
		}
		if _, err := os.Lstat(target); err == nil && !*force {
			return fmt.Errorf("%s already exists; rerun with --force to overwrite", rel)
		}
		if *force {
			if err := os.RemoveAll(target); err != nil {
				return err
			}
		}
		if err := copyPath(source, target); err != nil {
			return err
		}
		fmt.Printf("Restored %s\n", rel)
	}
	return nil
}

func discoverProject() (project, error) {
	root, err := output(".", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return project{}, errors.New("run this inside a Git working tree")
	}
	gitDir, err := output(root, "git", "rev-parse", "--git-dir")
	if err != nil {
		return project{}, err
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(root, gitDir)
	}
	return project{root: root, config: filepath.Join(gitDir, "gittrackuntracked", "config.json")}, nil
}
func openProject() (project, config, error) {
	p, err := discoverProject()
	if err != nil {
		return project{}, config{}, err
	}
	data, err := os.ReadFile(p.config)
	if errors.Is(err, os.ErrNotExist) {
		return project{}, config{}, errors.New("not initialized; run `gitu init`")
	}
	if err != nil {
		return project{}, config{}, err
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return project{}, config{}, fmt.Errorf("read config: %w", err)
	}
	if cfg.Version != 1 || cfg.ProjectID == "" {
		return project{}, config{}, errors.New("invalid project configuration")
	}
	if cfg.StorageMode == "" {
		cfg.StorageMode = "shared"
	}
	if cfg.StorageMode != "shared" && cfg.StorageMode != "branch" {
		return project{}, config{}, errors.New("invalid storage mode in project configuration")
	}
	return p, cfg, nil
}
func vaultDir() (string, error) {
	if override := os.Getenv("GITU_VAULT"); override != "" {
		return filepath.Abs(override)
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "gitu", "vault"), nil
}
func readyVault() (string, error) {
	dir, err := vaultDir()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); errors.Is(err, os.ErrNotExist) {
		return "", errors.New("personal vault is not ready; run `gitu vault init --remote URL`")
	}
	return dir, nil
}
func vaultLocation(dir string, p project, cfg config) (files, manifestPath, branch string, err error) {
	base := filepath.Join(dir, "projects", cfg.ProjectID)
	if cfg.StorageMode == "branch" {
		branch, err = currentBranch(p.root)
		if err != nil {
			return "", "", "", err
		}
		base = filepath.Join(base, "branches", branchKey(branch))
	}
	return filepath.Join(base, "files"), filepath.Join(base, "manifest.json"), branch, nil
}
func saveConfig(p project, cfg config) error { return writeJSON(p.config, cfg) }
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}
func readManifest(path string) (manifest, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, false
	}
	var value manifest
	return value, json.Unmarshal(data, &value) == nil && value.Version == 1
}
func readProjectState(path string) (projectState, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return projectState{}, false
	}
	var value projectState
	return value, json.Unmarshal(data, &value) == nil && value.Version == 1 && (value.StorageMode == "shared" || value.StorageMode == "branch")
}

func cleanProjectPath(p project, raw string) (string, error) {
	root, err := filepath.EvalSymlinks(p.root)
	if err != nil {
		return "", err
	}
	var abs string
	if filepath.IsAbs(raw) {
		abs, err = filepath.EvalSymlinks(raw)
		if err != nil {
			abs = filepath.Clean(raw)
		}
	} else {
		abs = filepath.Join(root, raw)
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("path must be inside the project and cannot be its root")
	}
	return filepath.ToSlash(rel), nil
}
func validateCandidate(p project, full, rel string, allowSensitive bool) error {
	return filepath.WalkDir(full, func(current string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		candidate, err := filepath.Rel(p.root, current)
		if err != nil {
			return err
		}
		candidate = filepath.ToSlash(candidate)
		if entry.IsDir() {
			return nil
		}
		tracked, err := isTracked(p.root, candidate)
		if err != nil {
			return err
		}
		if tracked {
			return fmt.Errorf("%s contains %s, which is already tracked by the host repository", rel, candidate)
		}
		if looksSensitive(candidate) && !allowSensitive {
			return fmt.Errorf("%s contains sensitive-looking file %s (use --allow-sensitive only if intentional)", rel, candidate)
		}
		return nil
	})
}
func isTracked(root, rel string) (bool, error) {
	result, err := output(root, "git", "ls-files", "--error-unmatch", "--", rel)
	if err == nil && result != "" {
		return true, nil
	}
	if exitCode(err) == 1 {
		return false, nil
	}
	return false, err
}
func looksSensitive(rel string) bool {
	value, base := strings.ToLower(filepath.ToSlash(rel)), strings.ToLower(filepath.Base(rel))
	if base == ".env" || strings.HasPrefix(base, ".env.") || base == "id_rsa" || base == "id_ed25519" {
		return true
	}
	for _, word := range []string{"secret", "credential", "password", "token", "private", "apikey", "api_key"} {
		if strings.Contains(base, word) {
			return true
		}
	}
	for _, suffix := range []string{".pem", ".key", ".p12", ".pfx"} {
		if strings.HasSuffix(base, suffix) {
			return true
		}
	}
	return strings.Contains(value, "/.aws/") || strings.HasPrefix(value, ".aws/")
}

func comparePaths(left, right string) (string, error) {
	leftInfo, leftErr := os.Lstat(left)
	rightInfo, rightErr := os.Lstat(right)
	if errors.Is(leftErr, os.ErrNotExist) && errors.Is(rightErr, os.ErrNotExist) {
		return "up to date", nil
	}
	if errors.Is(leftErr, os.ErrNotExist) {
		return "missing local", nil
	}
	if errors.Is(rightErr, os.ErrNotExist) {
		return "not synced", nil
	}
	if leftErr != nil {
		return "", leftErr
	}
	if rightErr != nil {
		return "", rightErr
	}
	if leftInfo.Mode() != rightInfo.Mode() {
		return "changed", nil
	}
	equal, err := pathsEqual(left, right)
	if err != nil {
		return "", err
	}
	if equal {
		return "up to date", nil
	}
	return "changed", nil
}
func pathsEqual(left, right string) (bool, error) {
	li, err := os.Lstat(left)
	if err != nil {
		return false, err
	}
	ri, err := os.Lstat(right)
	if err != nil {
		return false, err
	}
	if li.Mode() != ri.Mode() {
		return false, nil
	}
	if li.Mode()&os.ModeSymlink != 0 {
		l, e := os.Readlink(left)
		if e != nil {
			return false, e
		}
		r, e := os.Readlink(right)
		return l == r, e
	}
	if !li.IsDir() {
		return filesEqual(left, right)
	}
	leftEntries, err := os.ReadDir(left)
	if err != nil {
		return false, err
	}
	rightEntries, err := os.ReadDir(right)
	if err != nil {
		return false, err
	}
	if len(leftEntries) != len(rightEntries) {
		return false, nil
	}
	for i := range leftEntries {
		if leftEntries[i].Name() != rightEntries[i].Name() {
			return false, nil
		}
		equal, err := pathsEqual(filepath.Join(left, leftEntries[i].Name()), filepath.Join(right, rightEntries[i].Name()))
		if err != nil || !equal {
			return equal, err
		}
	}
	return true, nil
}
func filesEqual(left, right string) (bool, error) {
	li, err := os.Stat(left)
	if err != nil {
		return false, err
	}
	ri, err := os.Stat(right)
	if err != nil {
		return false, err
	}
	if li.Size() != ri.Size() {
		return false, nil
	}
	l, err := os.Open(left)
	if err != nil {
		return false, err
	}
	defer l.Close()
	r, err := os.Open(right)
	if err != nil {
		return false, err
	}
	defer r.Close()
	lb, rb := make([]byte, 32*1024), make([]byte, 32*1024)
	for {
		ln, le := l.Read(lb)
		rn, re := r.Read(rb)
		if ln != rn || !bytes.Equal(lb[:ln], rb[:rn]) {
			return false, nil
		}
		if le == io.EOF && re == io.EOF {
			return true, nil
		}
		if le != nil {
			return false, le
		}
		if re != nil {
			return false, re
		}
	}
}
func copyPath(source, target string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(source)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return os.Symlink(link, target)
	}
	if info.IsDir() {
		if err := os.MkdirAll(target, info.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := copyPath(filepath.Join(source, entry.Name()), filepath.Join(target, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func hostRemote(root string) (string, error) {
	return output(root, "git", "remote", "get-url", "origin")
}
func currentBranch(root string) (string, error) {
	branch, err := output(root, "git", "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil || branch == "" {
		return "", errors.New("branch-scoped storage is unavailable in detached HEAD; check out a named branch first")
	}
	return branch, nil
}
func normalizeRemote(remote string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(remote)), ".git")
}
func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:12]
}
func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var out strings.Builder
	dash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out.WriteRune(r)
			dash = false
		} else if !dash {
			out.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(out.String(), "-")
}
func branchKey(branch string) string { return slug(branch) + "-" + shortHash(branch) }
func commitVaultMove(dir, path, message string) error {
	if err := run(dir, "git", "add", "-A", "--", path); err != nil {
		return err
	}
	if err := run(dir, "git", "commit", "--quiet", "-m", message); err != nil {
		return err
	}
	return run(dir, "git", "push")
}
func gitHasStagedChanges(dir string) (bool, error) {
	command := exec.Command("git", "diff", "--cached", "--quiet")
	command.Dir = dir
	err := command.Run()
	if exitCode(err) == 1 {
		return true, nil
	}
	return false, err
}
func contains(paths []string, want string) bool {
	for _, path := range paths {
		if path == want {
			return true
		}
	}
	return false
}
func run(dir, name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %s", name, strings.Join(args, " "), strings.TrimSpace(string(output)))
	}
	return nil
}
func output(dir, name string, args ...string) (string, error) {
	command := exec.Command(name, args...)
	command.Dir = dir
	result, err := command.Output()
	return strings.TrimSpace(string(result)), err
}
func exitCode(err error) int {
	var e *exec.ExitError
	if errors.As(err, &e) {
		return e.ExitCode()
	}
	return -1
}
