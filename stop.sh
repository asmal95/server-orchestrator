#!/bin/bash

kill $(ps aux | grep 'server-orchestrator' | grep -v grep | awk '{print $2}')