# Release Process

1. Ensure that the milestone for correct version number exists on GitHub.
2. Assign all closed PRs for this release to the milestone.
3. Ensure `main` is clean locally.
4. Create changelog using `phly/changelog-generator` and copy to clipboard:
    
        changelog-generator -u akrabat -r rodeo -m 2 | pbcopy

5. Tag the new version: `git tag -s 0.1.1` and paste in changelog.
6. Push tag to GitHub: `git push --follow-tags`.
7. Close milestone and create next one.
8. Create a [Release](https://github.com/akrabat/rodeo/releases) with the tag name as the release title.

   This will automatically run a GitHub Action that creates the binaries for this release with
   the tag set as the version number as described by `./rodeo --version`.

