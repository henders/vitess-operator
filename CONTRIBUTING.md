## Contributing to Vitess Operator

This file is an introduction to the vitess-operator codebase for those looking
to contribute.

### Developer Certificate of Origin

This project uses a [Developer Certificate of Origin](https://wiki.linuxfoundation.org/dco)
instead of a Contributor License Agreement.

Please certify each contribution meets the requirements in the
`DCO` file in the root of this repository by committing with
the `--signoff` flag (or the short form: `-s`):

```sh
git commit --signoff
```

If you contribute often, you may find it useful to install a git hook
to prompt you to sign off if you forget to add the flag:

```sh
tools/git/install-hooks.sh
```

### CRD API Documentation

It may help to start by reading the user-facing docs for the public API,
which can be found in `docs/api`.

### Go Package Documentation

You can browse the Go package docs at:

https://godoc.org/planetscale.dev/vitess-operator

### Directory Structure

This section summarizes how files are organized.

* `build`

  > This directory was generated by operator-sdk when initializing the project.

  * `_output`

    > When you run `make generate` or `make build`, the output artifacts are
    placed here first and then copied into the Docker image.

  * `bin`

    > These are some extra files that were generated by operator-sdk.
    They are also copied into the Docker image and form a custom entrypoint script.

* `cmd`

  > As is conventional for Go projects, this is where you'll find the `main()`
  function for each binary/command.

  * `manager`

    > This is the main binary for the operator, containing all controllers.
    The name refers to the "controller manager", which is the system that runs
    a set of controllers with shared Kubernetes API clients and caches.
    The `main()` function runs all controllers registered with `pkg/controller`.

* `deploy`

  > This directory is part of the mandatory operator-sdk project structure and
  cannot be renamed. It contains a generic configuration for launching the
  operator into any given Kubernetes cluster.

  > The `kustomization.yaml` file makes this directory work with [kustomize](https://github.com/kubernetes-sigs/kustomize).

  * `crds`

    > This directory contains the CustomResourceDefinition specs.

    > The files ending in `_crd.yaml` are regenerated by operator-sdk every time
    `make generate` is run, so manual changes cannot be made to them.
    To change anything in those CRD files, you must modify the Go structs and
    their Go doc comments in `pkg/apis/*` and then run `make generate`.
    There are also some k8s-code-generator-specific directives that appear in
    regular Go code comments in those files, such as those beginning with
    `+k8s` or `+kubebuilder`.

* `docs`

  > Documentation for using the operator.

  * `api`

    > HTML-formatted documentation explaining all the fields in the CRD API,
    directed at users of the operator who need to write the CRD spec to describe
    a particular Vitess cluster.

* `pkg`

  > Most Go code for the operator should live in here.

  * `apis`

    > This directory tree contains the Go structs that define the operator's
    public API. The CustomResourceDefinitions are generated from these structs.

    > The `apis.go` file contains a mechanism for each API package to register
    itself, so the `main()` function can easily install all APIs.
    The individual `addtoscheme_*.go` files are generated by operator-sdk when
    you add a new API group.

    * `planetscale`

      > This contains a subdirectory for each version of the CRD API.
      We started vitess-operator's version at `v2` because `v1` was taken by an
      earlier, unreleased operator version.

      * `v2`

        > This contains the Go structs for all `planetscale.com/v2` CRD APIs.
        The main starting point is `vitesscluster_types.go`, which defines
        the top-level VitessCluster CRD.
        You may also want to start with the user-facing docs that are generated
        from the comments in these Go files, found at `docs/api`.

  * `controller`

    > This directory tree contains the main reconciliation loop of each
    controller, which is like the entrypoint of the controller.
    You can think of this directory as being like `cmd` but for controllers that
    get registered into the controller manager, instead of for individual
    binaries/commands that get executed directly.

    > Each subdirectory generally corresponds to one of the CRDs that make up
    the operator's public API, although this could be a many-to-one mapping.
    Each controller is the implementation of (some portion of) the interface
    defined by the CRD.

    > The `controller.go` file contains a mechanism for each controller package
    to register itself, so the `main()` function can easily install all controllers.
    The individual `add_*.go` files are generated by operator-sdk when you use
    it to add a new controller.

  * `operator`

    > This directory tree contains Go packages that abstract various
    functionality needed by the controller reconciliation loops.
    Anything that's needed by more than one controller should generally be
    abstracted into a package here.

    > Each package should have its own documentation in the form of a Go doc
    comment on the package statement in one of the files. By convention, the
    doc comment is placed either in a `.go` file with the same name as the
    package directory, or in a file named `doc.go`.

    > You can browse these package docs in HTML form as described above.

### Generated Files

If you change anything in `pkg/apis`, you should run the code generators again
before making a new build or sending a PR.

Running the generators is kept separate from the build process so people who
just want to build the operator from source can use the generated files that are
already checked in.

```sh
make generate
make build
```

If `make generate` changes any files, make sure to commit them with your PR.
