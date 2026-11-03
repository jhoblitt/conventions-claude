# Building Ceph

Owns how a Ceph change is built and its unit tests are run locally before
it goes up for review: the containerized build, the knobs that pass through
it, capping it on a shared machine, trimming configure, building only the
targets a change touches, and the pitfalls that break a build for reasons
unrelated to the change. Runs under `SKILL.md`'s precedence and routing.
What CI builds and tests on the PR is `references/ci.md`; a local build is
not the running-system reproduction `references/tracker.md`, "Before
filing", asks for.

Every file named here was checked against `ceph/ceph` `main` at
`203e7c52025d`. The scripts change; on another tree, re-read the named file
before relying on a flag or a default.

## Build in a container, never on the host

Ceph is built with `src/script/build-with-container.py`, never with
`install-deps.sh`, `do_cmake.sh`, or `run-make.sh` run on the host.
`install-deps.sh` installs a distribution's worth of packages as root, and
the image the script builds carries all of it instead.

- `-d <distro>` (`--distro`) picks the base image; the default is
  `centos9`.
- `-e <step>` (`--execute`, repeatable) picks the step: `configure`,
  `build`, `custom`, `interactive`, and the test and package steps that
  `--help-build-steps` lists. `custom` runs the command after `--`, which
  the script joins with spaces and hands to `bash -c` in the source root,
  so quoting inside it is lost.
- `-b <dir>` (`--build-dir`) names the build directory, relative to the
  source root.
- `--extra=<arg>` (`-x`, repeatable) passes one argument to the container
  engine's `run`. Write it with `=`: `-x --memory=24g` is rejected, because
  argparse reads a value starting with `--` as another option.
- `--env-file <file>` passes environment to every run container. The script
  forwards no build variable itself, so this file, or
  `--extra=--env=<NAME>=<value>`, is how `CONFIGURE_ARGS`,
  `DEFAULT_MAKEOPTS`, and the rest below reach the scripts.
- `--dry-run` prints the engine commands without running them.

The image is built by `Dockerfile.build`, whose setup script sources
`src/script/run-make.sh` and runs its `prepare`, which runs
`install-deps.sh`. The `configure` step then runs `run-make.sh`'s
`configure` with `${CONFIGURE_ARGS}` appended, and `build` runs its
`build vstart`. `configure` is skipped when the build directory already
holds a `build.ninja` (`has_build_dir`, `src/script/lib-build.sh`), and
`do_cmake.sh` refuses an existing build directory, so a change to
`CONFIGURE_ARGS` needs a new `-b` directory or the old one removed.

Boost does not depend on the distro. `WITH_SYSTEM_BOOST` is `OFF` by
default (`CMakeLists.txt`), and CMake then builds the Boost version Ceph
pins, from `src/boost` or a tarball on `download.ceph.com`
(`cmake/modules/BuildBoost.cmake`). `run-make.sh` turns it on only when
Ceph's own `ceph-libboost` packages are installed under `/opt/ceph`.

### Compile cache

`do_cmake.sh` enables sccache when it is on `PATH`, and ccache only
otherwise; `Dockerfile.build` installs sccache, on `ppc64le` only from a
distro package. `--ccache-dir` sets only `CCACHE_DIR` and `CCACHE_BASEDIR`
inside the container, so in that image it does not reach the compiler
launcher. Run containers are `--rm` and only the source tree is mounted, so
a cache survives between runs only under a mounted path: set sccache's own
`SCCACHE_DIR` in the env file to a directory under the source root, or
mount one with `--extra=--volume=<host-dir>:<ctr-dir>:Z`.

## Run rootless

The script adds `--user=0` to every run whenever its caller is not root,
and mounts the source tree read-write. Under rootless podman, container
root maps to the calling user. If `CONTAINER_HOST` points podman at a
rootful socket, container root is real root, and every file the build
writes into the source tree is root-owned. Run the script under
`env -u CONTAINER_HOST` so the rootless engine is used.

## Cap it on a shared machine

An uncapped build takes every core and enough memory to stall a desktop.
Cap each layer, and serialize builds:

```sh
cat > <env-file> <<'EOF'
DEFAULT_MAKEOPTS=-j<jobs>
NPROC=<jobs>
CONFIGURE_ARGS=<cmake -D options>
EOF

flock <lock-file> env -u CONTAINER_HOST \
  ./src/script/build-with-container.py -d centos9 -b <build-dir> \
  --env-file <env-file> \
  --extra=--cpuset-cpus=<cpu-list> --extra=--memory=<limit> \
  -e configure
```

- `--extra=--cpuset-cpus=` and `--extra=--memory=` cap the run container.
  They do not reach the image build, which takes no `--extra`.
- `DEFAULT_MAKEOPTS` is the job flag `run-make.sh`'s `build` passes to the
  generator (`BUILD_MAKEOPTS` wins over it when set). Its default is half
  of `nproc`, or `NPROC` when that is set, and `NPROC` also caps the jobs
  of the bundled Boost build.
- A `custom` step runs no `run-make.sh`, so its ninja takes the cap itself:
  `ninja -C <build-dir> -j<jobs> <targets>`.
- Write env-file values unquoted; a `CONFIGURE_ARGS` line holds the options
  separated by spaces.
- With the Ninja generator, CMake also sizes compile and link job pools
  from the memory and cores it sees (`cmake/modules/LimitJobs.cmake`).
  `-DNINJA_MAX_LINK_JOBS=<n>` in `CONFIGURE_ARGS` keeps concurrent links
  inside a memory cap.
- `flock <lock-file>` around every build command, the same file for all of
  them, lets only one build run on the machine at a time.

## Trim configure

`run-make.sh`'s `configure` forces several optional components on, and
`CONFIGURE_ARGS` comes after its options on the `cmake` command line, so a
later `-D…=OFF` there wins. For a check that needs only RGW, these can go:

| Component | Off with |
|---|---|
| mgr dashboard frontend (an `npm` build) | `-DWITH_MGR_DASHBOARD_FRONTEND=OFF`, or `WITHOUT_DASHBOARD=1` in the env file |
| SPDK | `-DWITH_SPDK=OFF` |
| the fio plugin | `-DWITH_FIO=OFF` |
| rbd-mirror | `-DWITH_RBD_MIRROR=OFF` |
| cephfs-shell | `-DWITH_CEPHFS_SHELL=OFF` |
| grafana dashboards | `-DWITH_GRAFANA=OFF` |
| S3 Select on Parquet, and the bundled Arrow it pulls in | `-DWITH_RADOSGW_SELECT_PARQUET=OFF` |
| lancedb (needs cargo 1.91 or later) | `-DWITH_RADOSGW_LANCEDB=OFF` |

Crimson is an image choice as well as a CMake one. With `WITH_CRIMSON=`
(set, empty) in the script's own environment, the image is built without
Crimson's dependencies. Set it on the command line, not in the env file:
the script reads `WITH_CRIMSON` from there too, and rejects an empty value.

Keep any component the change touches. Turning the dashboard frontend off
also drops the dashboard's tox tests from the build (see "Python changes").

## Build only what the change touches

Configure once, then build the change's targets with a `custom` step:

```sh
... -e custom -- ninja -C <build-dir> -j<jobs> radosgw radosgw-admin unittest_rgw_<name>
... -e custom -- <build-dir>/bin/unittest_rgw_<name>
```

Binaries land in `<build-dir>/bin`. To check further branches, check each
out in the same tree and rerun the same ninja command; ninja rebuilds only
what the checkout changed. A branch that moves a submodule needs
`git submodule update --init --recursive` first.

For scale, one observed run at `-j4`: a base build of `radosgw`,
`radosgw-admin`, `ceph-mgr`, and a handful of unit tests took about 29
minutes and 638 ninja steps, and each later branch rebuilt in 7 seconds to
15 minutes.

## The source tree must be self-contained

Only the source root is mounted into the container, and `do_cmake.sh` runs
`git submodule update` whenever `.git` is a directory. A clone made with
`git clone --reference` points at objects through
`.git/objects/info/alternates`, whose path does not exist inside the
container: git there failed with `fatal: bad object HEAD`, and
`do_cmake.sh` with it. Clone without `--reference`, or run
`git repack -a -d` and delete `.git/objects/info/alternates` before
building.

## -Werror on the default compiler

`run-make.sh` adds `-Werror` through `CMAKE_CXX_FLAGS_DEBUG` unless
`WITHOUT_WERROR` is set. It applies to a Debug build, which CMake picks by
default when `.git` exists (`CMakeLists.txt`).

`discover_compiler` (`src/script/lib-build.sh`) picks a `clang-<N>` from 21
down to 12 if one is installed, and otherwise the base `gcc`/`g++`; an
existing gcc-toolset-11 `enable` file overrides either. The el9 image installs
gcc-toolset-13 (`ceph.spec.in`), which `discover_compiler` does not look
for, so a `centos9` build uses the base gcc.

At `203e7c52025d` that compiler, gcc 11.5.0, failed the default build at
`src/rgw/rgw_lua_utils.h` with `-Werror=address` ("the address of … will
never be NULL") on `if constexpr (PushElement == nullptr)`, a line no RGW
change need touch. To get past it, set `WITHOUT_WERROR=1` in the env file,
or pass an empty `-DCMAKE_CXX_FLAGS_DEBUG=` in `CONFIGURE_ARGS`. The build
then no longer fails on a warning, so grep its log for `warning:` in every
file the change touches and fix any it introduced.

## Python changes

A change only to Python, such as `src/pybind/mgr/dashboard`, is not checked
by building C++ targets. Its tests are tox environments, registered as
ctest tests labelled `PythonTox` (`run-tox-mgr`, `run-tox-mgr-dashboard-py3`,
…; `cmake/modules/AddCephTest.cmake`) only when tests are configured, and
the dashboard's only with its frontend on. Each run creates a virtualenv
and pip-installs that environment's requirements, which needs network
access; having the build image is not enough. Run them through ctest from
the build directory, or with `tox` in the module's directory.
