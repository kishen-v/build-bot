# Build Bot
Bot for tracking the builds

# How to Build and Push container images

```shell script
$ make all-build
$ make all-push
$ make push-manifest
```

# S3 credentials file

The path to this JSON file is passed with `--s3-credentials-file`. Example for IBM Cloud Object Storage:

```json
{
  "region": "us-south",
  "endpoint": "https://s3.us-south.cloud-object-storage.appdomain.cloud",
  "insecure": false,
  "s3_force_path_style": true,
  "access_key": "<access-key>",
  "secret_key": "<secret-key>"
}
```

- `region`, `endpoint`, `access_key` and `secret_key` are required.
- `endpoint` must be a full URL including the scheme (`https://...`).
- Credentials are read only from this file. Environment variables and instance roles are not used.
