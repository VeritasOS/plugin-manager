# Graph support for Plugin Manager (pm-graph)

Graph support for plugin manager using 'dot' command.
If dot command is available / installed on the system, then a dot file and an svg image file would be generated. If dot is not available / accessible, then only a dot file would be generated.

Each plugin in the svg image points (url links) to either the code (with pm list) or the log file (with pm run).
In case of pm run, it indicates the status of the plugin run - i.e., Succeeded, Failed, Skipped...
