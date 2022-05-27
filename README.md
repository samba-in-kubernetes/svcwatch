
# svcwatch - A Service Watcher

svcwatch is a small tool intended to help an application, running in a
Kubernetes Pod, keep track of external IP address assigned to it by
a Service.

svcwatch subscribes to the Kubernetes API and maintains a small JSON file
on disk, that other processes can consume.

## Usage

svcwatch can be run as a binary from the command line, or more typically from a
container image.

```
$ podman run --rm -it quay.io/samba.org/svcwatch:latest -h
Usage of /svcwatch:
      --destination string   JSON file to update (default from: DESTINATION_PATH)
      --label-key string     Label key to watch (default from: SERVICE_LABEL_KEY)
      --label-value string   Label value (default from: SERVICE_LABEL_VALUE)
      --namespace string     Namespace (default from: SERVICE_NAMESPACE)
```

All of the options available as CLI options will default to sourcing
settings from the environment variables listed above.

* The `--destination` option determines where svcwatch will write the
  output JSON file.
* The `--label-key` and `--label-value` options are used to determine
  what Service is to be monitored.
* The `--namespace` option determines what kubernetes namespace the
  Service will be found in.


## Development Note

I looked for an existing tool to do this but, somewhat to my surprise,  didn't
find anything.  If I overlooked something that already exists please let me
know in the project's issues.
