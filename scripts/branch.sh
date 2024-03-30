#!/bin/bash

if git rev-parse --git-dir > /dev/null 2>&1; then
  if [[ -z "${BRANCH}" ]]; then
      branch=`git rev-parse --abbrev-ref HEAD`
      if [ "$branch" = "main" ]; then
          echo "latest"
      else
          echo "$branch"
      fi
  else
      echo "${BRANCH}"
      exit 0
  fi
else
  echo "0.0"
fi