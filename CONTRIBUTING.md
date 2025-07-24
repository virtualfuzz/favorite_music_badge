<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

# Contributing

Contributing for contributing to favorite_music_badge.

## Dependencies

- [dprint](https://github.com/dprint/dprint)
- [pre-commit](https://github.com/pre-commit/pre-commit): To automatically run
  pre-commit hooks, run `pre-commit install` and
  `pre-commit install --hook-type commit-msg`

There are other dependencies, but they are automatically installed by
pre-commit.

## Specification

1. Config files filenames MUST start with `.` to be hidden by default.

2. File names, folder names, and project names MUST be in alphanumeric
   characters, without any spaces, using dashes instead of spaces.

3. Folder names, and project names MUST be using lowercase characters.

4. File names SHOULD be using lowercase characters. Some files may use uppercase
   characters such as `README.md` or `LICENSE.md`.

5. Every file MUST have an
   [SPDX identifier](https://spdx.dev/learn/handling-license-info/) on top of
   the file following the latest SPDX specification.

6. All licenses MUST be referred using their
   [SPDX identifiers](https://spdx.org/licenses/).

7. You MUST use [semantic versioning](https://semver.org/) to version your
   software.

8. You MUST use [conventional commits](https://conventionalcommits.org/) with
   the
   [config-conventional](https://npmjs.com/package/@commitlint/config-conventional)
   ruleset to format your commit messages.

9. Each commit MUST pass the
   [pre-commit](https://github.com/pre-commit/pre-commit) checks with the config
   files.
