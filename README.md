# Plugin Manager (PM)

[![Go](https://github.com/VeritasOS/plugin-manager/actions/workflows/go.yml/badge.svg?branch=v1)](https://github.com/VeritasOS/plugin-manager/actions/workflows/go.yml)

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
    - [Example: Plugin Manager (PM) `run -plugins`](#example-plugin-manager-pm-run--plugins)
      - [Specify `-plugins` details as a json string](#specify--plugins-details-as-a-json-string)
      - [Specify `-plugins` details via json file](#specify--plugins-details-via-json-file)
    - [Example: Plugin Manager (PM) `run -type`](#example-plugin-manager-pm-run--type)
    - [Example: Plugin Manager (PM) with `sequential` flag](#example-plugin-manager-pm-with-sequential-flag)
    - [Example: Overriding Plugin Manager (PM) configuration - `library`, `log-dir` and `log-file`](#example-overriding-plugin-manager-pm-configuration---library-log-dir-and-log-file)
    - [Example: Writing plugins result to a `output-file` in `output-format` {json, yaml} format](#example-writing-plugins-result-to-a-output-file-in-output-format-json-yaml-format)

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
  [-log-tag=<TagOfSysLog>]
  [-log-dir=<LogDirectory>]
  [-log-file=<NameOfLogFile>]
```

where

- **`type`**: Indicates the plugin type.
- **`library`**: Indicates the location of plugins library.
    **Overrides** value present in PM configuration.
- **`log-tag`**: Indicates the log tag written by rsyslog.
    Note: rsyslog is used as default logger for both main and plugin logs.
    It will be overwritten if `log-file` option set.
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
Plugin Manager configuration file `/opt/veritas/appliance/asum/pm.config.yml`.

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
  log file: "pm.log"
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
pm run [-plugins <PluginInformation>]
  [-type <PluginType>]
  [-library=<PluginsLibraryPath>]
  [-sequential[={true|1|false|0}]]
  [-log-tag=<TagOfSysLog>]
  [-log-dir=<LogDirectory>]
  [-log-file=<NameOfLogFile>]
  [-output={json|yaml}]
  [-output-file=<NameOfOutputFile>]
```

where

- **`plugins`**: A json string or a json file containing plugins and its dependencies.
- **`type`**: Indicates the plugin type.
- **`library`**: Indicates the location of plugins library.
    **Overrides** value present in PM configuration.
    **NOTE** The specified value gets set as an environment variable `PM_LIBRARY` for the plugins being run. The plugin file can access any scripts in the same folder via `PM_LIBRARY` variable.
- **`-sequential`**: Indicates PM to execute only one plugin at a time
    regardless of how many plugins' dependencies are met.
    **Default: Disabled**. To enable, specify `-sequential=true` or just
    `-sequential` while running PM.
- **`log-tag`**: Indicates the log tag written by rsyslog. The `log-tag` option will supercede `log-dir` and `log-file` options.
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

### Example: Plugin Manager (PM) `run -plugins`

```json
$ jq -n "$plugins" | tee sample/plugins-prereboot.json
{
  "Plugins": [
    {
      "Name": "A/a.prereboot",
      "Description": "Applying \"A\" settings",
      "ExecStart": "/usr/bin/ls -l -t",
      "Requires": [
        "C/c.prereboot",
        "D/d.prereboot"
      ]
    },
    {
      "Name": "B/b.prereboot",
      "Description": "Applying \"B\" settings...",
      "ExecStart": "/bin/echo \"Running B...\"",
      "RequiredBy": [
        "D/d.prereboot"
      ]
    },
    {
      "Name": "C/c.prereboot",
      "Description": "Applying \"C\" settings...",
      "ExecStart": "/bin/echo \"Running C...\"",
      "RequiredBy": [
        "A/a.prereboot"
      ]
    },
    {
      "Name": "D/d.prereboot",
      "Description": "Applying \"D\" settings...",
      "ExecStart": "/bin/echo 'Running D...!'",
      "RequiredBy": [
        "A/a.prereboot"
      ],
      "Requires": [
        "B/b.prereboot"
      ]
    }
  ]
}
$
```

#### Specify `-plugins` details as a json string

```bash
$ $GOBIN/pm run -plugins "$plugins"
Applying "B" settings...: Starting
Applying "C" settings...: Starting
Applying "B" settings...: Succeeded
Applying "D" settings...: Starting
Applying "C" settings...: Succeeded
Applying "D" settings...: Succeeded
Applying "A" settings: Starting
Applying "A" settings: Succeeded
Running  plugins: Succeeded
bash-5.1$
```

#### Specify `-plugins` details via json file

```bash
$ $GOBIN/pm run -plugins "./sample/plugins-prereboot.json" -library sample/library/
Applying "C" settings...: Starting
Applying "B" settings...: Starting
Applying "C" settings...: Succeeded
Applying "B" settings...: Succeeded
Applying "D" settings...: Starting
Applying "D" settings...: Succeeded
Applying "A" settings: Starting
Applying "A" settings: Succeeded
Running  plugins: Succeeded
$ 
```

### Example: Plugin Manager (PM) `run -type`

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
Log: /var/log/asum/pm.2021-01-29T17:46:57.6904918-08:00.log

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
      "Name": "D/d.preupgrade",
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
      "Name": "A/a.preupgrade",
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
Log: /var/log/asum/pm.2021-01-29T17:53:15.8128937-08:00.log

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
  name: D/d.preupgrade
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
