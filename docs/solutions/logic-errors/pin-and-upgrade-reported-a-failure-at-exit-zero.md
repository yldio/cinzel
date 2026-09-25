---
title: "Pin and upgrade reported a failure at exit 0"
module: "internal/command"
problem_type: logic_error
component: "internal/command"
severity: medium
root_cause: "both summaries counted and printed a failed result, and both actions then returned nil"
symptoms:
  - "pin prints '1 failed' and the process exits 0"
  - "upgrade does the same, and --parse regenerates YAML from a directory it failed on"
  - "a CI step calling cinzel github pin passes over an unpinned action"
tags:
  - pin
  - upgrade
  - reporting
  - exit code
status: fixed
created_date: "2026-09-25"
updated_date: "2026-09-25"
---

# Pin and upgrade reported a failure at exit 0

## What happened

A run that could not resolve an action printed the failure twice, once as a
warning and once in the count, and then succeeded:

```
warning: could not pin ./.github/actions/build: "./.github/actions/build" does
not name a GitHub repository, so there is nothing to resolve

Pin summary: 0 pinned, 0 already pinned, 1 failed
```

Exit 0. `upgrade` printed the same shape, also at exit 0.

The place this matters is CI. Pinning exists so a workflow cannot silently
follow a moving tag, and the usual way to enforce that is a step running
`cinzel github pin` and trusting its exit code. That step went green on a
repository still holding an unpinned action, which is the one outcome pinning is
there to prevent.

`upgrade` carried it further. Its `--parse` gate reads the same results, so a
run that failed on one action still regenerated YAML for the whole directory and
reported nothing wrong.

## Why

Both summaries count a failed result and print it, and neither had a way to say
so to the caller:

```go
func (cmd *Cli) printPinSummary(results []pin.PinResult) {
	...
	_, _ = fmt.Fprintf(cmd.Writer, "\nPin summary: %d pinned, %d already pinned, %d failed\n", pinned, skipped, failed)
}
```

The Action called it and returned `nil` on the next line. The count was right.
Nothing read it.

This is the second half of `a-skipped-file-counted-as-no-failure.md`, which
fixed the counting and said as much: "The exit code is left alone. A partial run
is still a run, and the question of whether pin should exit non-zero on a
skipped file is a separate decision from whether it should say so." This is that
decision.

## The fix

Both summaries return an error when the count is not zero, and both actions
return it:

```go
if failed > 0 {
	return fmt.Errorf("%w: %d of %d", errPinFailed, failed, len(results))
}

return nil
```

The sentinels are marked `cinzelerror.UserInput`, so the message ends without
the invitation to file a bug. An action that names no repository, or a version
that is not a tag, is the author's to fix.

Two call sites needed a decision rather than a return.

`upgrade --parse` writes the regenerated YAML first and returns the failure
after. What upgraded is correct, and holding the write back would leave the
directory half converted, with the HCL updated and the YAML not. The failure is
still the command's outcome.

`assist` pins as a post-step, after the generated HCL is already on disk, and
reports a failed action as a warning instead. The file the prompt paid for is
not thrown away over an action that would not resolve.

## Prevention

A count printed to the screen is not a result. Anywhere a function's whole job
is to describe how something went, ask what the caller does with that
description: if the answer is nothing, the description is decoration and the
program is claiming success it did not check.

The warning makes this hard to see, the same way it did in the note above. The
information is on screen, so the code looks like it reports the failure. The
exit code is the only part of it another program can read.

## Tests

`internal/command/pin_exit_code_test.go` drives the real CLI through
`Cli.Execute`, so what it asserts on is the error the process exits with rather
than what a summary helper returned.

The failing case uses a local action, which names no repository: both commands
refuse it without a request, so the test needs no network and no stub. The
control for `pin` uses an action already on a full SHA, which is skipped the
same way. `upgrade` has no offline success path with a real action in it, since
it asks for the latest release whatever the version already says, so its control
uses a step with no `uses` block at all.
