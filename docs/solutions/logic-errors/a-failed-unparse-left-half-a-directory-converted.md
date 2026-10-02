---
title: "A failed unparse left half a directory converted"
module: "provider/github, provider/gitlab"
problem_type: logic_error
component: "provider/github/github.go, provider/gitlab/gitlab.go"
severity: high
root_cause: "the write happened inside the conversion loop, so a failure returned with earlier files already on disk"
symptoms:
  - "unparse exits 1 and the output directory holds some of the files"
  - "which files survive depends on where the failing file sorts by name"
  - "only the first failing file is named, however many are broken"
tags:
  - unparse
  - partial-write
  - reporting
status: fixed
created_date: "2026-10-02"
updated_date: "2026-10-02"
---

# A failed unparse left half a directory converted

## The symptom

Two workflows, one of them holding a `needs` that names no job in the file:

```
$ ls in; cinzel github unparse -d in --output-directory out; ls out
a-good.yml  z-bad.yml
error in file '.../z-bad.yml': jobs.build.needs: cannot find needed job 'nope'
a-good.hcl
```

Exit 1, and `a-good.hcl` is on disk. Rename the two files so the broken one
sorts first and the same input produces a different directory:

```
$ ls in; cinzel github unparse -d in --output-directory out; ls out
a-bad.yml   z-good.yml
error in file '.../a-bad.yml': jobs.build.needs: cannot find needed job 'nope'
ls: out: No such file or directory
```

Same two documents, same failure, two different results. What survived was
decided by alphabetical order, and the message named only the one file at fault.

A half-converted directory reads as a finished one. The usual next step after
unparse is deleting the YAML and keeping the HCL, and the workflows the run
never reached go with it. These are the files GitHub actually runs.

GitLab's `Unparse` had the same loop and the same result.

## The cause

The write was inside the conversion loop:

```go
for _, file := range files {
    hclBytes, name, err := unparseYAMLFile(...)
    if err != nil {
        return fmt.Errorf("error in file '%s': %w", file, err)
    }
    ...
    if err := fsutil.WriteFile(outputPath, hclBytes); err != nil {
        return err
    }
}
```

So the run did stop at the failure — it always had. Stopping was never the
missing part. What was missing is that *stop* did not mean *nothing written*:
every iteration before the failing one had already completed its write, and
every iteration after it was never reached.

`Parse` on the same provider does not have this. `checkOutputPaths` runs over
the fully-converted slices before either write loop, under a comment saying so:
"Checked before anything is written, so a collision leaves the directory as it
was rather than partway through the run." The two directions had different
guarantees for no reason anyone chose.

## The fix

Both providers convert every file into memory first, collecting failures rather
than returning at the first, and write only once every recognised file has
converted. A failure anywhere means the write loop is never reached.

The failures are joined with `errors.Join` instead of returning the first. A
directory of five broken workflows is read through in one run rather than five,
one key at a time. Three properties make that safe to drop in:

- A single failure is byte-identical to before, since `errors.Join` with one
  element adds no separator. Every `.error.txt` fixture matched unchanged.
- `errors.Is` and `errors.As` traverse a join, so the sentinel assertions and
  `cinzelerror.IsUserInput` keep working.
- `SafeForTerminal` deliberately leaves `\n` alone, so the joined message
  reaches the terminal as separate lines.

GitLab's `found = true` sat before its conversion rather than after it, so a run
whose only pipeline failed had already recorded that it found one. Nothing read
that while the failure returned on the spot. Now that the run reaches the end,
the failures are checked before the no-definitions guard, and the ordering stops
mattering.

## Skipping is not failing

The two outcomes stay different, and that difference is the whole shape of the
change. A file that is not a workflow, an action or a set of steps — a
dependabot config, an issue template — is warned about and passed over, and the
run goes on. Nothing is lost there: cinzel never had that file. Only a file
cinzel recognises and then fails on stops the run.

This is the opposite direction from
`a-skipped-file-counted-as-no-failure.md`, where a file-level failure was warned
about and dropped so the summary counted nothing. There, a failure was being
treated as a skip. Here, a failure was being reported correctly and the files
around it were not.

## The drifted IDs, left alone

`usedStepIDs` and `usedJobIDs` are claimed as each block is written, during
conversion rather than after it. Continuing past a failure therefore leaves the
failed file's IDs taken, and a later file's `checkout` can come out as
`checkout_2`.

That is left as it is, on purpose:

- The bytes are discarded. Any failure means nothing is written.
- Drift cannot cause a failure. `naming.UniqueIdentifierInSet` returns a string
  and has no error path.
- The only trace is a drifted ID quoted inside a later file's error message, in
  a run that is already failing.

Clearing the sets per file would be worse: sharing them across a directory is
what stops two workflows each holding a "mise setup" step from producing two
`step "mise_setup"` blocks that cinzel then cannot read back.

## Prevention

A loop that validates and writes in the same pass has no failure mode between
"all" and "some". The guarantee comes from ordering the phases, not from any
machinery: convert everything, decide, then write. `Parse` already had it.

The ordering dependence is what hid this. Half the test runs of such a bug look
correct, and which half depends on a filename.

## Tests

`provider/github/unparse_all_or_nothing_test.go` and
`provider/gitlab/unparse_all_or_nothing_test.go`. Each covers both orderings:
the good-file-first case is the bug and was confirmed failing before the fix,
and the bad-file-first case passed before it too, which makes it a control
rather than a restatement — on its own it would only prove the run stops.

Also covered: every failure is named rather than the first; a skipped file
beside a good one still lets the run write; a dry run with a bad file prints
nothing, since printing up to the failure is the same partial result on stdout.
