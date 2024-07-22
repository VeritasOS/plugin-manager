# Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

# A python library for invoking Plugin Manager (PM) to run plugins in parallel while specifying dependencies between plugins.

import json
import logging
import os
import sys
import subprocess
import tempfile
import traceback
from pathlib import Path


DEFAULT_PM_PATH = "/opt/veritas/appliance/asum/bin/pm"

logger = logging.getLogger()


def get_pm_path():
    """
    get_pm_path returns Plugin Manager binary path based on:

    1. Returns PM path in current directory if present.
        This enables upgrade RPMs that bundle latest PM to use it.
    2. Returns default PM path if present in default path.
    3. Return "" when it cannot be found in current dir or default dir.
    """
    pm_path = os.path.dirname(os.path.abspath(__file__)) + "/pm"
    if os.path.exists(pm_path):
        return pm_path
    if os.path.exists(DEFAULT_PM_PATH):
        return DEFAULT_PM_PATH
    return ""


def run(plugins=None,
        type=None,
        library=os.path.dirname(os.path.abspath(__file__)),
        display_stdout=True,
        log_file=None,
        output_format="json",
        output_file=None
        ):
    """
    Run commands in parallel depending on the order specified.
    Returns 0 on success, and 1 on failure

    "plugins"       A map of plugins containing their description, command and any dependencies information.
                        Format: {"plugin-name": {"Description": "Description of plugin", "ExecStart": "command to run",
                            "Requires": "List of space separated plugins to be run before current one", "RequiredBy": "List of plugins that want current plugin to be run before them."}}
    "display_stdout"    Display Plugin Manager output.
                            Set to 'False' when you do not want output to be displayed on console and just want results to be returned.
    """
    logger.debug(
        f"Entering run(plugins={plugins}, type={type}, library={library}, display_stdout={display_stdout}, log_file={log_file}, output_format={output_format}, output_file={output_file})...")

    pm_path = get_pm_path()
    logger.debug(f"PM binary path: {pm_path}")
    if pm_path == "":
        err_msg = f"Failed to find plugin manager 'pm' binary. Install VRTSvxos-asum*.rpm and then try again."
        logger.error(err_msg)
        return 1, ""

    cmd = [pm_path, "run"]
    if plugins != None:
        cmd = cmd + ["-plugins", str(plugins)]
    if type != None:
        cmd = cmd + ["-type", type]
    if library != None:
        cmd = cmd + ["-library", library]
    log_dir = None
    if log_file != None:
        log_dir = os.path.dirname(log_file)
        cmd = cmd + ["-log-dir", log_dir,
                     "-log-file", os.path.basename(log_file)]
    if output_format == None:
        output_format = "json"
    if output_file == None:
        prefix = ""
        if type != None:
            prefix = type+"-"
        tmp_dir = "/tmp"
        if log_dir != None:
            tmp_dir = log_dir
            # INFO: tmp dir needs to exist for mkstemp to work, so create if it
            # doesn't exist.
            if not os.path.exists(tmp_dir):
                os.makedirs(tmp_dir, exist_ok=True)
        _, output_file = tempfile.mkstemp(
            dir=tmp_dir, prefix=prefix, suffix="."+output_format)
        logger.debug("File name: %s", output_file)
    cmd = cmd + ["-output-format", output_format,
                 "-output-file", output_file]
    logger.debug("COMMAND: %s", ' '.join(cmd))

    popen_options = {}
    if display_stdout == False:
        output_file_no_ext = Path(output_file).with_suffix("").as_posix()
        output_file = output_file_no_ext+"."+output_format
        popen_options["stdout"] = open(
            output_file_no_ext+"_pm_stdout.log", mode="w+")
        popen_options["stderr"] = open(
            output_file_no_ext+"_pm_stderr.log", mode="w+")
    logger.debug(f"subprocess.Popen() options: {popen_options}")
    process = subprocess.Popen(cmd, **popen_options)
    process.wait()  # Wait for the subprocess to finish
    logger.debug(f"CMD output: {process}")
    if process.returncode != 0:
        logger.debug(
            f"COMMAND: {' '.join(cmd)} exited with status: {process.returncode}")
        # INFO: Do not return here because of non-zero exit!!!
        #   Plugin Manager (PM) would return non-zero in case if any of the
        #   plugins fail. But, we still want to send those failures back to
        #   caller, so continue down below...

    pluginsStatus = ""
    # For now, supporting only 'json'. "Maybe" support 'yaml' in future when required!
    if output_format == "json":
        try:
            with open(output_file, 'r') as json_fp:
                pluginsStatus = json.load(json_fp)
        except Exception as e:
            logger.error(
                f"Failed to load data from json file {output_file}.")
            logger.debug(f"Error: {str(e)}")
            logger.debug(traceback.format_exc())

    return process.returncode, pluginsStatus
