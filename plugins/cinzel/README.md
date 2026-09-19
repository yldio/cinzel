# cinzel plugin

Teaches your agent to use [cinzel](https://github.com/yldio/cinzel), which
converts CI/CD pipelines between HCL and YAML.

## Install

```
/plugin marketplace add yldio/cinzel
/plugin install cinzel@cinzel-plugin
```

Then install the CLI itself, which the plugin drives:

```sh
brew tap yldio/cinzel && brew install --cask cinzel
# or
go install github.com/yldio/cinzel@latest
```

## What you get

**`cinzel` skill** — loads on its own when a repo has a `cinzel/` directory or
a `.cinzelrc.yaml`, or when you ask to convert a pipeline to or from HCL. It
covers both providers, both directions, the pin and upgrade commands, and the
handful of things about the HCL shape that surprise people.

**`/parse` and `/unparse`** — run the conversion over a whole directory and
show you the diff. Both take the provider as their argument:

```
/parse github
/unparse gitlab
```

## Using the skill outside Claude Code

The skill is a plain `SKILL.md`. Codex, opencode and pi read the same format
from their own skills directory, so one copy serves all of them:

```sh
git clone https://github.com/yldio/cinzel /tmp/cinzel
mkdir -p ~/.agents/skills
cp -r /tmp/cinzel/plugins/cinzel/skills/cinzel ~/.agents/skills/

ln -s ../../.agents/skills/cinzel     ~/.claude/skills/cinzel
ln -s ../../.agents/skills/cinzel     ~/.codex/skills/cinzel
ln -s ../../../.agents/skills/cinzel  ~/.config/opencode/skills/cinzel
ln -s ../../../.agents/skills/cinzel  ~/.pi/agent/skills/cinzel
```

Link only the ones you use. Agents that read `AGENTS.md` rather than skill
files can be pointed at the file instead:

```markdown
- `~/.agents/skills/cinzel/SKILL.md` — converting CI/CD pipelines to or from HCL
```
