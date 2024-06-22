# Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

# A python library for invoking Plugin Manager (PM) to run plugins in parallel while specifying dependencies between plugins.

import logging
import os
import sys
from pathlib import Path

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "../../../", "opt/veritas/appliance/bin"))
import pm

DEFAULT_LOG_DIR = "/var/log/asum/"

logger = logging.getLogger()


def help(exit_status=None):
    """ Prints help on command line """
    cmd = Path(__file__).name
    print(f"""
Plugin Manager allows one to run multiple commands in parallel while specifying any dependencies among them.

Usage:

  {cmd} --plugins <PLUGINS_WITH_DEPENDENCY>
    [-type <"Type of plugin e.g. 'precheck'">]
    [--json-output-file <"path/to/output/file.json">])

Where,
  output-format     The format of output to display the results. Supported output formats are 'json', 'yaml'.
  output-file       path to file to write the results.
  library           path to plugins' library contains plugin directories and
                     plugin files.
                    Default: Current directory of script. I.e., {os.path.dirname(os.path.abspath(__file__))}.
                    Library directory structure:
                        $PM_LIBRARY/$pluginDir/$pluginFile.$pluginType
  plugins           plugins and its dependencies.""")

    print("""
                    Format:
                    {
                      "plugin-name": {
                        "Description": "Description of plugin",
                        "ExecStart": "command to run",
                        "Requires": "[space separated plugins to be run before current one]",
                        "RequiredBy": "[space separated plugins that want current plugin to be run before them.]"
                      }
                    }
  type              type of plugins e.g. "precheck".

Example:

$ plugins='{"plugin1": {"Description": "plugin 1 description...", "ExecStart": "echo command to run..."}, "plugin2": {"Description": "plugin 2 description...", "ExecStart": "echo command to run..."}, "plugin3": {"Description": "Plugin 3 depends on 1 and 2", "ExecStart": "echo Running plugin 3", "Requires": [ "plugin1" , "plugin2"]}}'
    """)

    print(f'''$ {cmd} --plugins "$plugins" --log-dir /tmp --json-output-file /tmp/ab.json
          ''')

    if exit_status != None:
        exit(exit_status)


def init_logging(log_file_path):
    log_dir = os.path.dirname(log_file_path)
    if not os.path.exists(log_dir):
        os.makedirs(log_dir)
    logging.basicConfig(filename=log_file_path,
                        encoding='utf-8', level=logging.DEBUG)
    print(f"Log file: {log_file_path}")


if __name__ == "__main__":
    log_file = None
    output_file = None
    output_format = None
    plugins = None
    type = None

    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument("--plugins",
                        type=str,
                        help="Plugins' and its dependencies in json format as a string or in a file (Ex: './plugins.json')")
    parser.add_argument("--type",
                        help="Type of plugin.")
    parser.add_argument("--log-file",
                        default=DEFAULT_LOG_DIR + Path(__file__).stem + ".log",
                        type=str,
                        help="Path to the log file.")
    parser.add_argument("--output-file",
                        help="Name of the file to write the results.")
    parser.add_argument("--output-format",
                        choices=['json', 'yaml'],
                        default="json",
                        help="The format of output to display the results.")
    args = parser.parse_args()

    init_logging(args.log_file)

    logger.debug("Args:")
    logger.debug(args)

    status, plugins_result = pm.run(
        plugins=args.plugins,
        type=args.type,
        log_file=args.log_file,
        output_file=args.output_file,
        output_format=args.output_format)
    logger.debug("Status:")
    logger.debug(status)
    if status != 0:
        logger.error("Failed to run specified plugins.")
        print("Failed to run specified plugins.")
        exit(1)
    logger.info("Successfully ran specified plugins.")
    logger.info("Plugins Status")
    logger.info(plugins_result)
    # print("Plugins Status")
    # print(plugins_result)
