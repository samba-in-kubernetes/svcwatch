# svcwatch Release Process

## Preparation

Currently there is no dedicated branch for releases. svcwatch is simple enough,
has few dependencies, and we're not planning on doing backports. Therefore
we apply release tags to the master branch.

### Tagging

Example:
```
git checkout master
git pull --ff-only
git tag -a -m 'Release v0.3' v0.3
```

This creates an annotated tag. Release tags must be annotated tags.

### Build

Perform a final check that the image is built correctly. First, ensure that
base images are up-to-date:

```
podman pull docker.io/golang:1.18
podman pull registry.access.redhat.com/ubi8/ubi-minimal:latest
```

Note that the base images are expected to change over time. Make sure
you pull the images that are used by the project's Containerfile.


Build the image by running `make image-build`.

Double check that the build contains the proper git hash and tag. This can be
done by running the container and looking at the initial logging the svcwatch
binary produces.

### Image Push

Apply a temporary pre-release tag to the image:
```
podman tag quay.io/samba.org/svcwatch:latest quay.io/samba.org/svcwatch:v0.3pre1
```

Log in to quay.io. Example:
```
podman login quay.io
```

Push the image using the temporary pre-release tag. Example:
```
podman push quay.io/samba.org/svcwatch:v0.3pre1
```

Wait for the security scan to complete. There shouldn't be any issues if you
properly updated the base images before building. If there are issues and you
are sure you used the newest base images, check the base images on quay.io and
make sure that the number of issues are identical. The security scan can take
some time, while it runs you may want to do other things.


## GitHub Release

When you are satisfied that the tagged version is suitable for release, you
can push the tag to the public repo:
```
git push --follow-tags
```

Draft a new set of release notes. Select the recently pushed tag. Start with
the auto-generated release notes from GitHub (activate the `Generate release
notes` button/link). Add an introductory section (see previous notes for an
example). Add a "Highlights" section if there are any notable features or fixes
in the release. The Highlights section can be skipped if the content of the
release is unremarkable (e.g. few changes occurred since the previous release).

Because this is a container based release we do not provide any build artifacts
on GitHub (beyond the sources automatically provided there). Instead we add
a Download section that notes the exact tag and digest that the image can
be found at on quay.io.

Use the following as an example:
```
This release can be acquired from the quay.io image registry:
* By tag: quay.io/samba.org/svcwatch:v0.3
* By digest: quay.io/samba.org/svcwatch@sha256:cdc38bd80ce2bc0eef6b883c2f0697f5d693c14155ddb2725b095ecf9cd0c608
```

The tag is pretty obvious - it should match the image tag (minus any pre-release
marker). You can get the digest from the tag using the quay.io UI (do not use
any local digest hashes). Click on the SHA256 link and then copy the full
manifest hash using the UI widget that appears.

Perform a final round of reviews, as needed, for the release notes and then
publish the release.

Once the release notes are drafted and then either immediately before or after
publishing them, use the quay.io UI to copy the pre-release tag to the "latest"
tag and a final "vX.Y" tag. Delete the temporary pre-release tag using the
quay.io UI as it is no longer needed.
