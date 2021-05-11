
# svcwatch - A Service Watcher

svcwatch is a small tool intended to help an application, running in a
Kubernetes Pod, keep track of it's own external IP address as managed by a
Service associated with the Pod.

svcwatch subscribes to the Kubernetes API and maintains a small JSON file
on disk, that other processes can consume.

## Usage

TODO


## Development Note

I looked for an existing tool to do this but, somewhat to my surprise,  didn't
find anything.  If I overlooked something that already exists please let me
know in the project's issues.
