#!/usr/bin/python3

# Kubernetes SDK
from kubernetes import client as k8sclient, config as k8sconfig

import isodate
import sys
import argparse
import asyncio
import logging

class ProgramArgs:
    # _min_free: int
    # _rate: datetime.timedelta
    # _resource_type: string
    # _name: string
    # _namespace: string

    def __init__(self):
        self._iam_mappings = {}
        self._users_to_preserve = []
        self._user_arns_to_preserve = []

        # Use ArgParse library to parse arguments
        parser = argparse.ArgumentParser(description="Azure Pipeline Agent Autoscaler", add_help=True, allow_abbrev=True, prefix_chars='-')
        parser.add_argument("--min", "-m", action='store', default=1, type=int, help="Minimum number of free agents to keep alive. Minimum of 1.")
        parser.add_argument("--rate", "-r", action='store', default='PT10S', help="ISO 8601 duration to check the number of agents.")
        parser.add_argument("--type", "--resource-type", "-t", action='store', default='StatefulSet', help="Resource type of the agent. Only StatefulSet is supported.")
        parser.add_argument("--name", "-n", action='store', default='azp-agent', help="The name of the StatefulSet. Defaults tp azp-agent")
        parser.add_argument("--namespace", "-ns", action='store', default='azp', help="The name of the StatefulSet. Defaults to azp.")

        args = parser.parse_args()

        self._min_free = args.min
        if self._min_free < 1:
            raise Exception("Invalid arguments - at least one agent must be kept alive.")
        
        rateStr = args.rate
        try:
            self._rate = isodate.parse_duration(rateStr)
        except Exception:
            exec_type, exec_value, exec_traceback = sys.exc_info()
            msg = f"{exec_type} exception raised when parsing rate string '{rateStr}': {exec_value}"
            logging.error(msg)
            raise Exception(msg)
        
        self._resource_type = args.type
        self._name = args.name
        self._namespace = args.namespace

        if self._resource_type != "StatefulSet":
            raise Exception(f"Unknown resource type '{self._resource}'")

async def main():
    logging.basicConfig(format='[%(levelname)s] %(asctime)s: %(message)s', level='INFO')
    
    logging.info("Starting azp-agent-autoscaler")

    args = ProgramArgs()
    
    try:
        k8sconfig.load_incluster_config()
    except k8sconfig.config_exception.ConfigException:
        exec_type, exec_value, exec_traceback = sys.exc_info()
        logging.warning(f"Received exception type {exec_type} with value {exec_value} when getting in-cluster Kubernetes config. Attempting to load ~/.kube config")
        try:
          k8sconfig.load_kube_config()
        except k8sconfig.config_exception.ConfigException:
            exec_type, exec_value, exec_traceback = sys.exc_info()
            logging.error(f"Received exception type {exec_type} with value {exec_value} when loading ~/.kube config")
    
    k8sv1 = k8sclient.CoreV1Api()

asyncio.run(main())
