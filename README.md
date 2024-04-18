# Plugin Manager (PM)

[![Go](https://github.com/VeritasOS/plugin-manager/actions/workflows/go.yml/badge.svg?branch=v2)](https://github.com/VeritasOS/plugin-manager/actions/workflows/go.yml)

Plugin Manager (PM) provides a way for components to define actions and
validations via plugins. PM provides dependency management similar to Red Hat
systemd.

**Table of Contents**

- [Plugin Manager (PM)](#plugin-manager-pm)
  - [Plugins](#plugins)
  - [Plugin Types and File Extensions](#plugin-types-and-file-extensions)
  - [Plugin Dependencies](#plugin-dependencies)
    - [Viewing Plugin and its dependencies](#viewing-plugin-and-its-dependencies)
      - [Example: Plugin Manager (PM) `list`](#example-plugin-manager-pm-list)
  - [Configuring Plugin Manager](#configuring-plugin-manager)
  - [Running Plugins](#running-plugins)
    - [Example: Plugin Manager (PM) `run`](#example-plugin-manager-pm-run)
    - [Example: Plugin Manager (PM) with `sequential` flag](#example-plugin-manager-pm-with-sequential-flag)
    - [Example: Overriding Plugin Manager (PM) configuration - `library`, `log-dir` and `log-file`](#example-overriding-plugin-manager-pm-configuration---library-log-dir-and-log-file)
    - [Example: Writing plugins result to a `output-file` in `output-format` {json, yaml} format](#example-writing-plugins-result-to-a-output-file-in-output-format-json-yaml-format)
  - [Workflow](#workflow)
    - [Running workflow](#running-workflow)
    - [Example](#example)
  - [Plugin Manager Web Server](#plugin-manager-web-server)
    - [Start the server](#start-the-server)
    - [Service to start `pm server`](#service-to-start-pm-server)
    - [Logs](#logs)

## Plugins

Plugin Manager basically uses config that are known as *Plugins* to inform
about what it does, it’s dependencies, and the action to be performed.

- **`Description`**: about the action i.e., what the plugin does.
- **`ExecStart`**: refers to action to be performed i.e., binary to be executed.
  - **Note**: `ExecStart` should have absolute path to the binary, so that PM
    knows where the binary is located. In case, binaries are located in the
    same plugin directory, then you could specify path using the PM's plugins
    library path that would be updated in the environment variable
    i.e., `PM_LIBRARY`.
  - **Example**: `ExecStart=/bin/sh ${PM_LIBRARY}/<component>/example.sh`.
- **`RequiredBy`**: informs that the current plugin must be run before the
  specified plugins.
  In other words, the specified plugins must be run after the current plugin.
- **`Requires`**: informs that the current plugin must be run after the
  specified plugins.
  In other words, the specified plugins must be run before the current plugin.

All plugins must be installed (extracted) into
`${PM_LIBRARY}/<component-plugin-dir>` folder. If you would like to customize
this path, you could either

1. Create/update the config file to look for a different plugins library path,
    and set the `PM_CONF_FILE` to the config file.
2. Specify the plugins library path i.e., `-library` while calling PM.

The `PM_LIBRARY` would be set to the `library` location specified in the PM configuration file. The configuration file currently has `library` as `/system/upgrade/repository/plugins/`, but could change in future. And hence one must use the environment variable `${PM_LIBRARY}` to access the plugins library path.

## Plugin Types and File Extensions

The type of a plugin is identified basically based on plugin file's extension, and it's up to the consumer of the plugin manager to define the plugin types for their actions.  

**Example**: To perform pre-upgrade tasks by various services/components/features in the upgrade workflow, one could define `.preupgrade` plugin type, and have all these plugins called through plugin manager by specifying  `-type` as `preupgrade`.

## Plugin Dependencies

Plugin Manager allows specifying dependencies between plugins.
Basically it allows a plugin to be run before or after certain plugins are run.
The below sample show that "d" plugins requires “b” and “c” plugins to be run
first before it runs, and “A” must be run after "D".

```bash
$ cat <plugins_library>/A/a.prereboot
Description=Applying “A” settings
ExecStart=<path_to_binary>
```

```bash
$ cat <plugins_library>/D/d.prereboot
Description=Applying “D” settings
RequiredBy=A/a.prereboot
Requires=B/b.prereboot C/c.prereboot
ExecStart=${PM_LIBRARY}/D/example.sh
```

### Viewing Plugin and its dependencies

The plugins and its dependencies can be viewed visually in a svg image by running the `list` command of Plugin Manager.

The PM list command syntax / usage is as shown below:

```bash
pm list -type <PluginType>
  [-library=<PluginsLibraryPath>]
  [-log-dir=<LogDirectory>]
  [-log-file=<NameOfLogFile>]
```

where

- **`type`**: Indicates the plugin type.
- **`library`**: Indicates the location of plugins library.
    **Overrides** value present in PM configuration.
- **`log-dir`**: Indicates the log directory path.
    **Overrides** value present in PM configuration.
- **`log-file`**: Indicates the name of the log file.
    **Overrides** value present in PM configuration.

#### Example: Plugin Manager (PM) `list`

```bash
$ $GOBIN/pm list -type=preupgrade
Log: pm.2020-01-13T15:56:46.6817949-08:00.log
The list of plugins are mapped in .//preupgrade.2020-01-13T15:56:46.725348-08:00.svg
```

## Configuring Plugin Manager

Plugin Manager can be configured to look for plugins at a specific location,
and to write logs to a specific file by specifying those details in the
Plugin Manager configuration file `/etc/asum/pm.config.yml`.

Instead of updating the default config file, one can choose to provide his/her
own custom config file.
This can be done by setting the environment variable `PM_CONF_FILE` to the
custom config file path as shown below.

```bash
export PM_CONF_FILE=sample/pm.config.yaml
```

The config file could be either a `yaml` or a `json` file.
Below is a sample `yaml` configuration file.

```bash
$ cat pm.config.yaml
---
PluginManager:
  # `library` is the location where plugin directories containing plugins are expected to be present
  library: "./sample/library"
  log dir: "./"
  # `log file` indicates the name of the log file.
  #   The timestamp and '.log' extension would be appended to this name.
  #   I.e., The format of the log file generated would be: "<log file>.<timestamp>.log"
  #   Example: The below value results in following log file: pm.2020-01-13T16:11:58.6006565-08:00.log
  log file: "pm"
...
```

## Running Plugins

Plugin Manager tries to executes all available plugins of a certain type as
mentioned by the `-type` argument. The result of the execution of a plugin
(i.e., result of execution of binary which is specified in `ExecStart`) is
marked as `Succeeded` or `Failed` to mean success or failure respectively.
The PM checks for exit status of the binary to infer success or failure.
If the binary exits with 0, then plugin execution is marked as `Succeeded`,
while any non zero exit value is considered as `Failed`. In case of non zero
exit value of plugins, the PM exits with 1.

The PM run command syntax / usage is as shown below:

```bash
pm run -type <PluginType>
  [-library=<PluginsLibraryPath>]
  [-sequential[={true|1|false|0}]]
  [-log-dir=<LogDirectory>]
  [-log-file=<NameOfLogFile>]
  [-output={json|yaml}]
  [-output-file=<NameOfOutputFile>]
```

where

- **`type`**: Indicates the plugin type.
- **`library`**: Indicates the location of plugins library.
    **Overrides** value present in PM configuration.
- **`-sequential`**: Indicates PM to execute only one plugin at a time
    regardless of how many plugins' dependencies are met.
    **Default: Disabled**. To enable, specify `-sequential=true` or just
    `-sequential` while running PM.
- **`log-dir`**: Indicates the log directory path.
    **Overrides** value present in PM configuration.
- **`log-file`**: Indicates the name of the log file.
    **Overrides** value present in PM configuration.
- **`output`**: Indicates the format to write the plugins run results.
    Supported formats: "json", "yaml".
- **`output-file`**: Indicates the name of the output file.
    **Note** Specified in conjunction with `output`.
    If `output` format is specified, and `output-file` is not specified,
    then result will be displayed on console.

### Example: Plugin Manager (PM) `run`

```bash
$ $GOBIN/pm run -type=prereboot
Log: pm.2019-07-12T15:23:07.3494206-07:00.log

Applying "B" settings: Starting

Applying "C" settings: Starting
Applying "C" settings: Succeeded
Applying "B" settings: Succeeded

Applying "D" settings: Starting
Applying "D" settings: Succeeded

Applying "A" settings: Starting
Applying "A" settings: Succeeded
Running prereboot plugins: Succeeded
$
```

### Example: Plugin Manager (PM) with `sequential` flag

The `sequential` option informs Plugin Manager to execute one plugin at a time.
By default, this is disabled, and multiple plugins whose dependencies are met
would be run in parallel.

```bash
$ $GOBIN/pm run -type=prereboot -sequential
Log: pm.2019-07-12T15:36:33.7415514-07:00.log

Applying "B" settings: Starting
Applying "B" settings: Succeeded

Applying "C" settings: Starting
Applying "C" settings: Succeeded

Applying "D" settings: Starting
Applying "D" settings: Succeeded

Applying "A" settings: Starting
Applying "A" settings: Succeeded
Running prereboot plugins: Succeeded
$
```

### Example: Overriding Plugin Manager (PM) configuration - `library`, `log-dir` and `log-file`

To override the values in the PM configuration, specify one or many of the
following optional arguments: `library`, `log-dir` and `log-file`

```bash
$ $GOBIN/pm run -type postreboot -library=sample/library/ -log-dir=testlogs/ -log-file=test.log
Log: pm.2019-07-12T15:39:08.1145946-07:00.log
Log: testlogs/test.2019-07-12T15:39:08.1209416-07:00.log

Validating "A's" configuration: Starting
Validating "A's" configuration: Succeeded
Running postreboot plugins: Succeeded
$ ls testlogs/
test.2019-07-12T15:39:08.1209416-07:00.log
$
```

### Example: Writing plugins result to a `output-file` in `output-format` {json, yaml} format

```bash
$ $GOBIN/pm run -type preupgrade -output-format=json -output-file=a.json -library ./sample/library/
Log: /log/asum/pm.2021-01-29T17:46:57.6904918-08:00.log

Checking for "D" settings...: Starting
Checking for "D" settings...: Succeeded

Checking for "A" settings: Starting
Checking for "A" settings: Succeeded
Running preupgrade plugins: Succeeded
$
```

```json
$ cat a.json
{
  "Type": "preupgrade",
  "Plugins": [
    {
      "Description": "Checking for \"D\" settings...",
      "FileName": "D/d.preupgrade",
      "ExecStart": "$PM_LIBRARY/D/preupgrade.sh",
      "RequiredBy": [
        "A/a.preupgrade"
      ],
      "Requires": null,
      "Status": "Succeeded",
      "StdOutErr": "Running preupgrade.sh (path: sample/library//D/preupgrade.sh) with status(0)...\nDisplaying Plugin Manager (PM) Config file path: \nDone(0)!\n"
    },
    {
      "Description": "Checking for \"A\" settings",
      "FileName": "A/a.preupgrade",
      "ExecStart": "/bin/echo \"Checking A...\"",
      "RequiredBy": null,
      "Requires": [
        "D/d.preupgrade"
      ],
      "Status": "Succeeded",
      "StdOutErr": "\"Checking A...\"\n"
    }
  ],
  "Status": "Succeeded",
  "StdOutErr": ""
}$
```

```bash
$ $GOBIN/pm run -type preupgrade -output-format=yaml -output-file=a.yaml -library ./sample/library/
Log: /log/asum/pm.2021-01-29T17:53:15.8128937-08:00.log

Checking for "D" settings...: Starting
Checking for "D" settings...: Failed
Running preupgrade plugins: Failed
$
```

```yaml
$ cat a.yaml 
type: preupgrade
plugins:
- description: Checking for "D" settings...
  filename: D/d.preupgrade
  execstart: $PM_LIBRARY/D/preupgrade.sh
  requiredby:
  - A/a.preupgrade
  requires: []
  status: Failed
  stdouterr: "Running preupgrade.sh (path: sample/library//D/preupgrade.sh) with
    status(1)...\nDisplaying Plugin Manager (PM) Config file path: \nFail(1)\n"
status: Failed
stdouterr: 'Running preupgrade plugins: Failed'
$
```

## Workflow

Plugin Manager (PM) provides workflow option, where one can specify multiple action and rollback plugin types to run. Without `Workflow` support, one has to wait for one set of plugins execution to be complete before triggering the next one.
The Workflow option helps one to automatically trigger either the next `Action` or `Rollback` depending on current execution status.

### Running workflow

```bash
./bin/pm run --workflow '[{"action": "preupgrade", "rollback": "requiredby"}, {"action": "postreboot", "rollback": "prereboot"}]' -library sample/library/ -output-format json -output-file /tmp/ab.json;
echo; echo; echo;
jq . /tmp/ab.json;
```

### Example

```bash
@abhijithda ➜ /workspaces/plugin-manager (workflow) $ ./bin/pm run --workflow '[{"action": "preupgrade", "rollback": "requiredby"}, {"action": "postreboot", "rollback": "prereboot"}]' -library sample/library/ -output-format json -output-file /tmp/ab.json 
Log: pm.2024-04-02T03:10:05.356691548Z.log

Running action plugins: preupgrade [1/2]...

Checking for "D" settings...: Starting
Checking for "D" settings...: Succeeded

Checking for "E" settings: Starting

Checking for "A" settings: Starting
Checking for "E" settings: Succeeded
Checking for "A" settings: Succeeded
Running preupgrade plugins: Succeeded

Running action plugins: postreboot [2/2]...

Validating "A's" configuration: Starting
Validating "A's" configuration: Failed
Running postreboot plugins: Failed
ERROR: Running postreboot plugins: Failed

Starting rollback...
Running rollback plugins: prereboot [1/2]...

Applying "B" settings: Starting

Applying "C" settings: Starting
Applying "C" settings: Succeeded
Applying "B" settings: Succeeded

Applying "D" settings: Starting
Applying "D" settings: Succeeded

Applying "A" settings: Starting
Applying "A" settings: Succeeded
Running prereboot plugins: Succeeded

Running rollback plugins: requiredby [2/2]...

Applying "A" settings: Starting
Applying "A" settings: Succeeded

Applying "B" settings: Starting
Applying "B" settings: Succeeded

Applying "C" settings: Starting
Applying "C" settings: Succeeded
Running requiredby plugins: Succeeded
ERROR: Running Workflow: Failed
@abhijithda ➜ /workspaces/plugin-manager (workflow) $ 
```

## Plugin Manager Web Server

Plugin Manager can be run as a service.
A minimal UI exists to list or run the specified list of `Plugin Type`s in the given `Plugin Library`.

### Start the server

```bash
$ mkdir ./log
$ ./pm server -port 8080 -log-dir ./log
```

Open given port in browser
http://localhost:8080/

### Service to start `pm server`

To start automatically on `pm server` on say RHEL, you can create a systemd unit file as follows:

```bash
$ cat /etc/systemd/system/pm-server.service
[Unit]
Description=Plugin Manager web server
After=network.target
Environment="PM_WEB=/home/abhijith/web"  ## Make sure to update this environment variable appropriately.

[Service]
ExecStart=/usr/local/bin/pm server -log-dir /var/log/pm-server/ 
ExecStart=/storage/bin/pm-server
Type=simple

[Install]
WantedBy=default.target
$
```

### Logs

To tail the `pm-server.service` logs, run the following command:

```bash
journalctl -u pm-server -e -f
```
