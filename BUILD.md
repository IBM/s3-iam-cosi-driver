# Building `s3-iam-cosi-driver` images

The `s3-iam-cosi-driver` images are automatically built and pushed to `icr.io` through our CI/CD pipeline for each commit or merge into the `main` branch.  Manual building is typically not necessary.

If you need to build the images manually, here's the instructions to manually build the driver.

To build, run the following:

```bash
❯ make docker-build
```

This should product a locally build image called `icr.io/cosi-research/s3-iam-cosi-driver`

```console
❯ docker images icr.io/cosi-research/s3-iam-cosi-driver
REPOSITORY                                TAG             IMAGE ID       CREATED         SIZE
icr.io/cosi-research/s3-iam-cosi-driver   1.1.0-e241271   c0982678d94b   2 minutes ago   74.7MB
icr.io/cosi-research/s3-iam-cosi-driver   latest          c0982678d94b   2 minutes ago   74.7MB
```

To push the image to the image repository, first log into ICR:

```bash
❯ ibmcloud cr login
```

Then push the images:

```bash
❯ make docker-push
```
