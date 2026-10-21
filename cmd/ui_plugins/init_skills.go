package ui_plugins

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
)

const (
	// agentSkillsSubdir is the repository-root skills location inside the
	// templates archive (framework-agnostic skills).
	agentSkillsSubdir = ".agents/skills"
)

// agentSkillsDir is the OS-specific destination path (relative to a project
// root) the skills are delivered into.
var agentSkillsDir = filepath.Join(".agents", "skills")

// pathOnlyFrontmatter matches the per-skill opt-out that keeps a repo-wide skill
// out of framework starters (delivered only by `init --path`). It is matched
// anywhere in the SKILL.md YAML frontmatter, nesting-agnostic, to stay
// consistent with the templates-repo PRB skill-collision check.
var pathOnlyFrontmatter = regexp.MustCompile(`(?m)^\s*sailpoint-merge:\s*path-only\s*$`)

// skillSummary tallies what a skills-delivery pass did, so silent (non-prompted)
// writes are still surfaced to the user.
type skillSummary struct {
	added       int
	skipped     int
	overwritten int
}

func (s skillSummary) empty() bool {
	return s.added == 0 && s.skipped == 0 && s.overwritten == 0
}

func (s skillSummary) String() string {
	return fmt.Sprintf("added %d, skipped %d (already present), overwritten %d", s.added, s.skipped, s.overwritten)
}

// deliverSkills merges the repository-wide (framework-agnostic) agent skills from
// the templates archive into destDir/.agents/skills/.
//
//   - framework == true  (new-workspace flow): skills whose frontmatter sets
//     `sailpoint-merge: path-only` are skipped (a framework-specific version is
//     authored in the starter instead). A name already present in the
//     destination is a collision with a scaffolded starter skill — a template
//     authoring error — and fails loudly.
//   - framework == false (attach flow): every repo-wide skill is delivered,
//     including path-only ones (this is their only delivery path). A name already
//     present belongs to the user; it is left untouched unless overwrite is
//     confirmed (prompt, or force).
//
// A template that ships no repo-wide skills is handled gracefully: the pass is a
// no-op returning an empty summary and no error. Only skill folders the CLI
// ships are ever touched.
func deliverSkills(r *bufio.Reader, deps initDeps, archive []byte, destDir string, framework, force bool) (skillSummary, error) {
	var sum skillSummary

	staging, err := os.MkdirTemp("", "sp-ui-plugin-skills-")
	if err != nil {
		return sum, fmt.Errorf("failed to create skills staging directory: %w", err)
	}
	defer os.RemoveAll(staging)

	counts, err := initialize.ExtractSubtrees(bytes.NewReader(archive), staging, []initialize.SubtreeMapping{
		{Subdir: agentSkillsSubdir, Dest: "."},
	})
	if err != nil {
		return sum, err
	}
	if counts[agentSkillsSubdir] == 0 {
		// The template carries no repository-wide skills — skip gracefully.
		return sum, nil
	}

	folders, err := listSkillFolders(staging)
	if err != nil {
		return sum, err
	}

	destSkills := filepath.Join(destDir, agentSkillsDir)
	for _, name := range folders {
		src := filepath.Join(staging, name)

		if framework && isPathOnlySkill(filepath.Join(src, "SKILL.md")) {
			// Opted out of framework starters; delivered only by `init --path`.
			continue
		}

		target := filepath.Join(destSkills, name)
		_, statErr := os.Stat(target)
		exists := statErr == nil

		if !exists {
			if err := copyTree(src, target); err != nil {
				return sum, err
			}
			sum.added++
			continue
		}

		if framework {
			// A starter scaffolded this same skill name — the root and framework
			// sources collide and would silently overwrite. Fail loudly rather
			// than interleave; mark the repo-wide skill `sailpoint-merge:
			// path-only` if it should not merge into frameworks.
			return sum, fmt.Errorf("agent skill %q is delivered by both the repository-wide skills and the scaffolded starter; rename one, or mark the repo-wide skill `sailpoint-merge: path-only`", name)
		}

		// Attach flow: the existing folder is the user's. Confirm before touching.
		if !force {
			overwrite, err := promptYesNoShared(r, deps.out, fmt.Sprintf("Agent skill %q already exists. Overwrite?", name))
			if err != nil {
				return sum, err
			}
			if !overwrite {
				fmt.Fprintf(deps.out, "Skipped skill %s\n", name)
				sum.skipped++
				continue
			}
		}
		if err := os.RemoveAll(target); err != nil {
			return sum, fmt.Errorf("failed to replace skill %q: %w", name, err)
		}
		if err := copyTree(src, target); err != nil {
			return sum, err
		}
		sum.overwritten++
	}

	return sum, nil
}

// removeStarterSkills deletes the skills the scaffolded starter bundled under
// destDir/.agents/skills/, used to honor --no-skills in the new-workspace flow.
// The empty .agents wrapper is removed too when nothing else remains in it.
func removeStarterSkills(destDir string) error {
	if err := os.RemoveAll(filepath.Join(destDir, agentSkillsDir)); err != nil {
		return err
	}
	// os.Remove removes .agents only when it is now empty; a non-empty wrapper
	// (other agent config) is left intact, so the error is intentionally ignored.
	_ = os.Remove(filepath.Join(destDir, ".agents"))
	return nil
}

// countSkillFolders returns the number of immediate skill directories in dir,
// or 0 if dir does not exist (no skills were delivered).
func countSkillFolders(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			n++
		}
	}
	return n, nil
}

// listSkillFolders returns the names of the immediate skill directories staged
// under dir, in sorted order (os.ReadDir sorts), so delivery and prompting are
// deterministic.
func listSkillFolders(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read staged skills: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// isPathOnlySkill reports whether the skill's SKILL.md frontmatter opts out of
// framework-starter delivery. A missing or unreadable SKILL.md is treated as not
// opted out (the default is to merge everywhere).
func isPathOnlySkill(skillMdPath string) bool {
	data, err := os.ReadFile(skillMdPath)
	if err != nil {
		return false
	}
	fm, ok := yamlFrontmatter(data)
	if !ok {
		return false
	}
	return pathOnlyFrontmatter.Match(fm)
}

// yamlFrontmatter returns the bytes between the leading `---` fence and the next
// `---` line. ok is false when the document has no frontmatter block.
func yamlFrontmatter(data []byte) ([]byte, bool) {
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, false
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return []byte(strings.Join(lines[1:i], "\n")), true
		}
	}
	return nil, false
}

// copyTree recursively copies the directory src to dst, creating parents as
// needed. Regular files and directories only (skills are Markdown); this mirrors
// extractGzipTar, which drops symlinks and special files.
func copyTree(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyTree(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
