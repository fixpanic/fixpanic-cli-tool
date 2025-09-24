# Release Management Guide

This document explains how to manage Git tags for releases in the `fixpanic-cli-tool` repository. Follow these steps to create, push, delete, and reuse tags for versioned releases.

---

## 1. Creating a Tag

To create a new tag (e.g., `v1.2.3`) on your current commit:

```sh
git tag v1.2.3
```

To create an annotated tag with a message:

```sh
git tag -a v1.2.3 -m "Release v1.2.3"
```

---

## 2. Pushing a Tag to GitHub

After creating a tag, push it to the remote repository to trigger the release workflow:

```sh
git push origin v1.2.3
```

---

## 3. Deleting a Tag

### Delete a tag locally:
```sh
git tag -d v1.2.3
```

### Delete a tag from GitHub (remote):
```sh
git push origin :refs/tags/v1.2.3
```

---

## 4. Reusing a Tag (Overwrite a Tag)

If you need to reuse a tag (e.g., after fixing a release):

1. Delete the tag locally and remotely (see above).
2. Re-create the tag on the desired commit:
   ```sh
   git tag v1.2.3
   git push origin v1.2.3
   ```

---

## 5. Listing Tags

To see all tags in the repository:
```sh
git tag
```

---

## 6. Checking Out a Tag

To check out a specific tag (detached HEAD):
```sh
git checkout v1.2.3
```

---

## Notes
- Tags are used to trigger the release workflow in GitHub Actions.
- Always use the `v` prefix (e.g., `v1.2.3`) for release tags.
- Only push tags you intend to release.
- Deleting a tag from GitHub does not delete the associated release automatically; you may need to delete the release manually from the GitHub web UI if needed.

---

For more details, see the [Git documentation on tagging](https://git-scm.com/book/en/v2/Git-Basics-Tagging).
